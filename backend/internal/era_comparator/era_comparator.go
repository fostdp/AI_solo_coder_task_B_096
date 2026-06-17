package era_comparator

import (
	"fmt"
	"math"
	"time"

	"tashan-weir-seepage/internal/dam_presets"
	"tashan-weir-seepage/internal/models"
	"tashan-weir-seepage/internal/simulation"
)

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
