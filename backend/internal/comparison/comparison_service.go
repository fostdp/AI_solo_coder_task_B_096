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

func InteractiveAdjustment(req *models.InteractiveAdjustmentRequest) (*models.InteractiveAdjustmentResult, error) {
	preset := dam_presets.GetDamPreset(req.DamKey)
	if preset == nil {
		return nil, fmt.Errorf("dam not found: %s", req.DamKey)
	}

	solver := simulation.NewSeepageSolverFromPreset(preset)
	solver.SetGridResolution(50, 25)

	upWL := req.UpstreamWL
	downWL := req.DownstreamWL

	if upWL <= 0 {
		upWL = preset.DesignUpstreamWL
	}
	if downWL <= 0 {
		downWL = preset.DesignDownstreamWL
	}

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

	if baselineResult != nil {
		flowDiff := (simResult.TotalSeepageFlow - baselineResult.TotalSeepageFlow) / baselineResult.TotalSeepageFlow
		if flowDiff > 0.5 {
			flowChange = fmt.Sprintf("渗流量增加%.1f%%，显著上升", flowDiff*100)
		} else if flowDiff > 0.2 {
			flowChange = fmt.Sprintf("渗流量增加%.1f%%，有所上升", flowDiff*100)
		} else if flowDiff < -0.2 {
			flowChange = fmt.Sprintf("渗流量减少%.1f%%，有所下降", -flowDiff*100)
		} else {
			flowChange = "渗流量基本稳定"
		}

		pressureDiff := (simResult.MaxPorePressure - baselineResult.MaxPorePressure) / baselineResult.MaxPorePressure
		headDiff := upWL - downWL
		designDiff := preset.DesignUpstreamWL - preset.DesignDownstreamWL
		overload := headDiff / designDiff

		switch {
		case overload > 1.5 || pressureDiff > 0.8:
			riskLevel = "critical"
			explanation = fmt.Sprintf("当前水位差%.1fm已超过设计值的%.1f倍，扬压力剧增%.1f%%，存在严重安全风险，建议立即降低水位！",
				headDiff, overload, pressureDiff*100)
		case overload > 1.2 || pressureDiff > 0.5:
			riskLevel = "high"
			explanation = fmt.Sprintf("当前水位差%.1fm超过设计值%.1f%%，扬压力增加%.1f%%，需加强监测",
				headDiff, (overload-1)*100, pressureDiff*100)
		case overload > 1.0 || pressureDiff > 0.2:
			riskLevel = "medium"
			explanation = fmt.Sprintf("当前水位差%.1fm略高于设计值，扬压力增加%.1f%%，处于警戒状态",
				headDiff, pressureDiff*100)
		default:
			riskLevel = "low"
			explanation = fmt.Sprintf("当前水位差%.1fm在设计范围内，渗流状态正常，坝体安全", headDiff)
		}
	}

	keyMetrics := map[string]float64{
		"total_seepage_flow_lps":  simResult.TotalSeepageFlow * 1000,
		"max_pore_pressure_kpa":   simResult.MaxPorePressure,
		"upstream_wl_m":           upWL,
		"downstream_wl_m":         downWL,
		"water_head_difference_m": upWL - downWL,
		"grid_count":              float64(simResult.GridCount),
		"calculation_time_ms":     float64(simResult.CalculationTimeMs),
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
