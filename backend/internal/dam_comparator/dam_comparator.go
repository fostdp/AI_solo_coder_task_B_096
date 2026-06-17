package dam_comparator

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
