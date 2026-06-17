package comparison

import (
	"fmt"
	"math"
	"sync"
	"time"

	"tashan-weir-seepage/internal/dam_presets"
	"tashan-weir-seepage/internal/models"
	"tashan-weir-seepage/internal/simulation"
)

func CompareDams(req *models.DamComparisonRequest) (*models.DamComparisonResult, error) {
	startTime := time.Now()

	if len(req.DamKeys) < 2 {
		return nil, fmt.Errorf("at least 2 dams required for comparison")
	}

	nx := req.GridResolutionX
	if nx <= 0 {
		nx = 50
	}
	ny := req.GridResolutionY
	if ny <= 0 {
		ny = 25
	}

	upWL := req.UpstreamWaterLevel
	downWL := req.DownstreamWaterLevel

	items := make([]models.DamComparisonItem, 0, len(req.DamKeys))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, damKey := range req.DamKeys {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()

			preset := dam_presets.GetDamPreset(key)
			if preset == nil {
				return
			}

			perm := preset.CurrentPermeability
			if req.IncludeCurrentAging {
				ageFactor := float64(time.Now().Year()-preset.BuildYear) / 1000.0
				perm = preset.OriginalPermeability * (1 + ageFactor*5)
			}

			solver := simulation.NewSeepageSolverFromPresetWithConfig(preset, nx, ny)
			solver.PermeabilityK = perm

			item, _, _ := solver.RunComparison(upWL, downWL)
			if item == nil {
				return
			}

			item.DamKey = preset.DamKey
			item.DamName = preset.DamName
			item.DamType = preset.DamType
			item.BuildDynasty = preset.BuildDynasty
			item.Geometry = preset.Geometry

			mu.Lock()
			items = append(items, *item)
			mu.Unlock()
		}(damKey)
	}

	wg.Wait()

	if len(items) < 2 {
		return nil, fmt.Errorf("not enough valid dam data for comparison")
	}

	summary := generateComparisonSummary(items, upWL, downWL)

	requestID := fmt.Sprintf("cmp_%d", time.Now().Unix())
	comparisonName := fmt.Sprintf("%d坝对比_%.1fm水头", len(items), upWL-downWL)

	calcTime := time.Since(startTime).Milliseconds()

	result := &models.DamComparisonResult{
		RequestID:            requestID,
		ComparisonName:       comparisonName,
		UpstreamWaterLevel:   upWL,
		DownstreamWaterLevel: downWL,
		Items:                items,
		Summary:              summary,
		CalculationTimeMs:    calcTime,
	}

	return result, nil
}

func generateComparisonSummary(items []models.DamComparisonItem, upWL, downWL float64) map[string]interface{} {
	summary := make(map[string]interface{})

	headDiff := upWL - downWL
	summary["water_head_difference_m"] = headDiff
	summary["dam_count"] = len(items)

	minFlow := math.MaxFloat64
	maxFlow := 0.0
	minPressure := math.MaxFloat64
	maxPressure := 0.0
	bestEfficiency := 0.0
	bestDam := ""

	for _, item := range items {
		if item.TotalSeepageFlow < minFlow {
			minFlow = item.TotalSeepageFlow
		}
		if item.TotalSeepageFlow > maxFlow {
			maxFlow = item.TotalSeepageFlow
		}
		if item.MaxPorePressure < minPressure {
			minPressure = item.MaxPorePressure
		}
		if item.MaxPorePressure > maxPressure {
			maxPressure = item.MaxPorePressure
		}
		if item.AntiSeepageEfficiency > bestEfficiency {
			bestEfficiency = item.AntiSeepageEfficiency
			bestDam = item.DamName
		}
	}

	summary["min_seepage_flow_lps"] = minFlow * 1000
	summary["max_seepage_flow_lps"] = maxFlow * 1000
	summary["flow_ratio"] = maxFlow / minFlow
	summary["min_pore_pressure_kpa"] = minPressure
	summary["max_pore_pressure_kpa"] = maxPressure
	summary["best_anti_seepage_dam"] = bestDam
	summary["best_anti_seepage_efficiency_pct"] = bestEfficiency

	ancientCount := 0
	modernCount := 0
	for _, item := range items {
		if item.DamType == models.DamTypeAncientStone {
			ancientCount++
		} else if item.DamType == models.DamTypeModernConcrete {
			modernCount++
		}
	}
	summary["ancient_dam_count"] = ancientCount
	summary["modern_dam_count"] = modernCount

	insights := []string{}
	if modernCount > 0 && ancientCount > 0 {
		insights = append(insights,
			fmt.Sprintf("对比包含%d座古代坝和%d座现代坝，可直观观察技术进步", ancientCount, modernCount))
	}
	if bestEfficiency > 0 {
		insights = append(insights,
			fmt.Sprintf("%s的防渗效率最高，达到%.1f%%", bestDam, bestEfficiency))
	}
	if maxFlow/minFlow > 3 {
		insights = append(insights,
			fmt.Sprintf("各坝渗流量差异显著，最大/最小比值达%.1f倍", maxFlow/minFlow))
	}
	summary["key_insights"] = insights

	return summary
}

func CrossEraComparison(req *models.CrossEraComparisonRequest) (*models.CrossEraComparisonResult, error) {
	startTime := time.Now()

	ancientPreset := dam_presets.GetDamPreset(req.AncientDamKey)
	if ancientPreset == nil {
		return nil, fmt.Errorf("ancient dam not found: %s", req.AncientDamKey)
	}

	modernPreset := dam_presets.GetDamPreset(req.ModernDamKey)
	if modernPreset == nil {
		return nil, fmt.Errorf("modern dam not found: %s", req.ModernDamKey)
	}

	upWL := req.UpstreamWL
	downWL := req.DownstreamWL

	ancientSolver := simulation.NewSeepageSolverFromPreset(ancientPreset)
	ancientSolver.SetGridResolution(60, 30)

	if req.ScaleToSameSize {
		ancientSolver.Geometry = simulation.DamGeometry{
			Length:          modernPreset.Geometry.Length,
			Height:          modernPreset.Geometry.Height,
			TopWidth:        modernPreset.Geometry.TopWidth,
			UpstreamSlope:   modernPreset.Geometry.UpstreamSlope,
			DownstreamSlope: modernPreset.Geometry.DownstreamSlope,
			FoundationDepth: ancientPreset.FoundationDepth,
		}
	}

	modernSolver := simulation.NewSeepageSolverFromPreset(modernPreset)
	modernSolver.SetGridResolution(60, 30)

	ancientItem, _, _ := ancientSolver.RunComparison(upWL, downWL)
	modernItem, _, _ := modernSolver.RunComparison(upWL, downWL)

	if ancientItem == nil || modernItem == nil {
		return nil, fmt.Errorf("simulation failed")
	}

	ancientItem.DamKey = ancientPreset.DamKey
	ancientItem.DamName = ancientPreset.DamName
	ancientItem.DamType = ancientPreset.DamType
	ancientItem.BuildDynasty = ancientPreset.BuildDynasty
	ancientItem.Geometry = ancientPreset.Geometry

	modernItem.DamKey = modernPreset.DamKey
	modernItem.DamName = modernPreset.DamName
	modernItem.DamType = modernPreset.DamType
	modernItem.BuildDynasty = modernPreset.BuildDynasty
	modernItem.Geometry = modernPreset.Geometry

	ancientMetrics := map[string]interface{}{
		"material":                ancientPreset.MaterialType,
		"permeability_ratio":      modernItem.Permeability / ancientItem.Permeability,
		"seepage_flow_ratio":      ancientItem.TotalSeepageFlow / modernItem.TotalSeepageFlow,
		"pressure_ratio":          ancientItem.MaxPorePressure / modernItem.MaxPorePressure,
		"build_year":              ancientPreset.BuildYear,
		"age_years":               time.Now().Year() - ancientPreset.BuildYear,
		"anti_seepage_method":     ancientPreset.AntiSeepageDescription,
		"cultural_value":          ancientPreset.CulturalValue,
		"historical_significance": ancientPreset.HistoricalSignificance,
	}

	modernMetrics := map[string]interface{}{
		"material":               modernPreset.MaterialType,
		"build_year":             modernPreset.BuildYear,
		"age_years":              time.Now().Year() - modernPreset.BuildYear,
		"anti_seepage_method":    modernPreset.AntiSeepageDescription,
		"technology_advantages":  []string{"防渗面板", "帷幕灌浆", "排水廊道", "温控系统"},
		"design_standards":       "GB 50201-2014 防洪标准",
	}

	flowImprovement := (ancientItem.TotalSeepageFlow - modernItem.TotalSeepageFlow) / ancientItem.TotalSeepageFlow * 100
	pressureImprovement := (ancientItem.MaxPorePressure - modernItem.MaxPorePressure) / ancientItem.MaxPorePressure * 100
	permeabilityImprovement := (ancientItem.Permeability - modernItem.Permeability) / ancientItem.Permeability * 100

	comparison := map[string]interface{}{
		"seepage_flow_reduction_pct":    flowImprovement,
		"pore_pressure_reduction_pct":   pressureImprovement,
		"permeability_reduction_pct":    permeabilityImprovement,
		"height_ratio":                  modernPreset.Geometry.Height / ancientPreset.Geometry.Height,
		"water_head":                    upWL - downWL,
		"scaled_to_same_size":           req.ScaleToSameSize,
		"ancient_age":                   time.Now().Year() - ancientPreset.BuildYear,
		"technology_gap_years":          modernPreset.BuildYear - ancientPreset.BuildYear,
		"ancient_dam_total_cost":        "无价（文化遗产）",
		"modern_dam_estimated_cost":     "约2.5亿元（同规模现代坝）",
	}

	insights := []string{
		fmt.Sprintf("从%s到%s，跨越%d年的水利工程技术对比",
			ancientPreset.BuildDynasty, modernPreset.BuildDynasty,
			modernPreset.BuildYear-ancientPreset.BuildYear),
		fmt.Sprintf("现代混凝土坝的渗透系数比古代条石坝降低%.1f%%，防渗性能大幅提升",
			math.Abs(permeabilityImprovement)),
		fmt.Sprintf("在相同水位差下，现代坝渗流量比古代坝减少%.1f%%，扬压力降低%.1f%%",
			flowImprovement, pressureImprovement),
		fmt.Sprintf("古代坝虽防渗性能稍逊，但历经千余年仍发挥作用，体现了古代工匠的卓越智慧"),
		fmt.Sprintf("古代坝的核心价值在于其文化和历史意义，是不可再生的文化遗产"),
		"现代坝在材料、设计方法、施工工艺上全面进步，但古代坝的工程理念仍值得借鉴",
	}

	result := &models.CrossEraComparisonResult{
		AncientDam:     *ancientItem,
		ModernDam:      *modernItem,
		AncientMetrics: ancientMetrics,
		ModernMetrics:  modernMetrics,
		Comparison:     comparison,
		Insights:       insights,
	}

	_ = startTime
	return result, nil
}

// ===== 修复4: 优化水位调节灵敏度参数 =====
const (
	SensitivitySmoothFactor    = 0.15  // 水位变化平滑系数，避免跳动
	FlowChangeMinThreshold     = 0.03  // 渗流量变化最小可感知阈值(3%)
	PressureChangeMinThreshold = 0.02  // 压力变化最小阈值(2%)
	SensorNoiseFloor           = 0.005 // 传感器噪声本底(0.5%)
	MinWaterLevelStep          = 0.01  // 最小水位步进(1cm)
	HysteresisBandLow          = 0.98  // 风险等级迟滞低带
	HysteresisBandHigh         = 1.02  // 风险等级迟滞高带
)

var lastRiskLevelStore = make(map[string]string) // 迟滞记忆

// ===== 修复4: 优化水位调节灵敏度 - 主函数 =====
func InteractiveAdjustment(req *models.InteractiveAdjustmentRequest) (*models.InteractiveAdjustmentResult, error) {
	preset := dam_presets.GetDamPreset(req.DamKey)
	if preset == nil {
		return nil, fmt.Errorf("dam not found: %s", req.DamKey)
	}

	solver := simulation.NewSeepageSolverFromPreset(preset)
	solver.SetGridResolution(50, 25)

	// ===== 修复4: 灵敏度优化 - 水位输入量化与钳位 =====
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

	// 水位值钳位，避免极端
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

	// 水位值量化到0.01m，避免浮点抖动导致的灵敏度闪烁
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

		// ===== 修复4: 灵敏度优化 - 传感器噪声过滤 =====
		var flowDiff float64
		if baselineFlow > SensorNoiseFloor {
			flowDiffRaw := (currentFlow - baselineFlow) / baselineFlow
			// 噪声过滤：小于阈值视为无变化
			if math.Abs(flowDiffRaw) < FlowChangeMinThreshold {
				flowDiff = 0
			} else {
				// 平滑处理，避免跳变
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

		// ===== 修复4: 灵敏度优化 - 风险等级迟滞机制 =====
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

	// ===== 修复4: 灵敏度优化 - 关键指标精度控制 =====
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

// ===== 修复4: 新增风险等级迟滞函数 =====
func applyRiskHysteresis(prevRisk, proposedRisk string, overload, pressureDiff float64) string {
	riskOrder := map[string]int{"low": 0, "medium": 1, "high": 2, "critical": 3}
	prevLevel := riskOrder[prevRisk]
	proposedLevel := riskOrder[proposedRisk]

	// 相同风险或升级直接采用
	if prevRisk == "" || proposedLevel >= prevLevel {
		return proposedRisk
	}

	// 降级时施加迟滞：需要远离边界足够距离才降级
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
