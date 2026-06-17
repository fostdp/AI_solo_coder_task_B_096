package vr_dam

import (
	"fmt"
	"math"

	"tashan-weir-seepage/internal/dam_presets"
	"tashan-weir-seepage/internal/models"
	"tashan-weir-seepage/internal/simulation"
)

const (
	SensitivitySmoothFactor    = 0.15
	FlowChangeMinThreshold     = 0.03
	PressureChangeMinThreshold = 0.02
	SensorNoiseFloor           = 0.005
	MinWaterLevelStep          = 0.01
	HysteresisBandLow          = 0.98
	HysteresisBandHigh         = 1.02
)

var lastRiskLevelStore = make(map[string]string)

func InteractiveAdjustment(req *models.InteractiveAdjustmentRequest) (*models.InteractiveAdjustmentResult, error) {
	preset := dam_presets.GetDamPreset(req.DamKey)
	if preset == nil {
		return nil, fmt.Errorf("dam not found: %s", req.DamKey)
	}

	solver := simulation.NewSeepageSolverFromPreset(preset)
	solver.SetGridResolution(50, 25)

	minWL := 0.5
	maxWL := preset.DesignUpstreamWL * 2.5

	upWL := req.UpstreamWL
	downWL := req.DownstreamWL

	if upWL <= 0 {
		upWL = preset.DesignUpstreamWL
	}
	if downWL <= 0 {
		downWL = preset.DesignDownstreamWL
	}

	if upWL < minWL {
		upWL = minWL
	}
	if upWL > maxWL {
		upWL = maxWL
	}
	if downWL < 0 {
		downWL = 0
	}
	if downWL >= upWL-0.1 {
		downWL = upWL - 0.1
	}

	upWL = math.Round(upWL*100) / 100
	downWL = math.Round(downWL*100) / 100

	simResult, grids, err := solver.Run(upWL, downWL, "interactive")
	if err != nil {
		return nil, err
	}

	baselineSolver := simulation.NewSeepageSolverFromPreset(preset)
	baselineSolver.SetGridResolution(50, 25)
	baselineResult, _, _ := baselineSolver.Run(preset.DesignUpstreamWL, preset.DesignDownstreamWL, "baseline")

	flowChange := ""
	riskLevel := "low"
	explanation := ""
	previousRisk := lastRiskLevelStore[req.DamKey]

	if baselineResult != nil {
		baselineFlow := baselineResult.TotalSeepageFlow
		currentFlow := simResult.TotalSeepageFlow

		var flowDiff float64
		if baselineFlow > SensorNoiseFloor {
			flowDiffRaw := (currentFlow - baselineFlow) / baselineFlow
			if math.Abs(flowDiffRaw) < FlowChangeMinThreshold {
				flowDiff = 0
			} else {
				flowDiff = flowDiffRaw * (1 - SensitivitySmoothFactor)
			}
		}

		if flowDiff > 0.5 {
			flowChange = fmt.Sprintf("渗流量增加%.1f%%，显著上升", flowDiff*100)
		} else if flowDiff > 0.2 {
			flowChange = fmt.Sprintf("渗流量增加%.1f%%，有所上升", flowDiff*100)
		} else if flowDiff < -0.2 {
			flowChange = fmt.Sprintf("渗流量减少%.1f%%，有所下降", -flowDiff*100)
		} else {
			flowChange = "渗流量基本稳定"
		}

		baselinePressure := baselineResult.MaxPorePressure
		currentPressure := simResult.MaxPorePressure

		var pressureDiff float64
		if baselinePressure > 0 {
			pressureDiffRaw := (currentPressure - baselinePressure) / baselinePressure
			if math.Abs(pressureDiffRaw) < PressureChangeMinThreshold {
				pressureDiff = 0
			} else {
				pressureDiff = pressureDiffRaw * (1 - SensitivitySmoothFactor)
			}
		}

		headDiff := upWL - downWL
		designDiff := preset.DesignUpstreamWL - preset.DesignDownstreamWL
		overload := headDiff / designDiff

		var proposedRisk string
		switch {
		case overload > 1.5 || pressureDiff > 0.8:
			proposedRisk = "critical"
		case overload > 1.2 || pressureDiff > 0.5:
			proposedRisk = "high"
		case overload > 1.0 || pressureDiff > 0.2:
			proposedRisk = "medium"
		default:
			proposedRisk = "low"
		}

		riskLevel = applyRiskHysteresis(previousRisk, proposedRisk, overload, pressureDiff)
		lastRiskLevelStore[req.DamKey] = riskLevel

		switch riskLevel {
		case "critical":
			explanation = fmt.Sprintf("当前水位差%.2fm已超过设计值的%.1f倍，扬压力剧增%.1f%%，存在严重安全风险，建议立即降低水位！",
				headDiff, overload, pressureDiff*100)
		case "high":
			explanation = fmt.Sprintf("当前水位差%.2fm超过设计值%.1f%%，扬压力增加%.1f%%，需加强监测",
				headDiff, (overload-1)*100, pressureDiff*100)
		case "medium":
			explanation = fmt.Sprintf("当前水位差%.2fm略高于设计值，扬压力增加%.1f%%，处于警戒状态",
				headDiff, pressureDiff*100)
		default:
			explanation = fmt.Sprintf("当前水位差%.2fm在设计范围内，渗流状态正常，坝体安全", headDiff)
		}
	}

	roundTo := func(v float64, places int) float64 {
		p := math.Pow(10, float64(places))
		return math.Round(v*p) / p
	}

	keyMetrics := map[string]float64{
		"total_seepage_flow_lps":  roundTo(simResult.TotalSeepageFlow*1000, 4),
		"max_pore_pressure_kpa":   roundTo(simResult.MaxPorePressure, 2),
		"upstream_wl_m":           roundTo(upWL, 2),
		"downstream_wl_m":         roundTo(downWL, 2),
		"water_head_difference_m": roundTo(upWL-downWL, 2),
		"grid_count":              float64(simResult.GridCount),
		"calculation_time_ms":     roundTo(float64(simResult.CalculationTimeMs), 1),
		"min_water_level_step_m":  MinWaterLevelStep,
		"sensitivity_level":       1.0,
	}

	result := &models.InteractiveAdjustmentResult{
		Simulation:       simResult,
		Grids:            grids,
		KeyMetrics:       keyMetrics,
		WaterLevelChange: flowChange,
		RiskLevel:        riskLevel,
		Explanation:      explanation,
	}

	return result, nil
}

func applyRiskHysteresis(prevRisk, proposedRisk string, overload, pressureDiff float64) string {
	riskOrder := map[string]int{"low": 0, "medium": 1, "high": 2, "critical": 3}
	prevLevel := riskOrder[prevRisk]
	proposedLevel := riskOrder[proposedRisk]

	if prevRisk == "" || proposedLevel >= prevLevel {
		return proposedRisk
	}

	triggerBands := map[string]float64{
		"critical_to_high":    1.4 * HysteresisBandLow,
		"high_to_medium":      1.15 * HysteresisBandLow,
		"medium_to_low":       0.95 * HysteresisBandLow,
	}

	switch prevRisk {
	case "critical":
		if overload < triggerBands["critical_to_high"] && pressureDiff < 0.7 {
			return proposedRisk
		}
		return prevRisk
	case "high":
		if overload < triggerBands["high_to_medium"] && pressureDiff < 0.45 {
			return proposedRisk
		}
		return prevRisk
	case "medium":
		if overload < triggerBands["medium_to_low"] && pressureDiff < 0.18 {
			return proposedRisk
		}
		return prevRisk
	default:
		return proposedRisk
	}
}
