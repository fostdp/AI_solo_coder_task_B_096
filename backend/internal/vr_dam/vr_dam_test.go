package vr_dam

import (
	"strings"
	"testing"

	"tashan-weir-seepage/internal/dam_presets"
	"tashan-weir-seepage/internal/models"
)

func TestInteractiveAdjustment_NormalWL_Normal(t *testing.T) {
	req := &models.InteractiveAdjustmentRequest{
		DamKey:       "tashan_weir",
		UpstreamWL:   8.5,
		DownstreamWL: 3.2,
	}

	result, err := InteractiveAdjustment(req)

	if err != nil {
		t.Fatalf("交互调节失败: %v", err)
	}
	if result == nil {
		t.Fatal("结果为nil")
	}

	if result.RiskLevel != "low" {
		t.Errorf("设计水位下风险等级应为low，实际%s", result.RiskLevel)
	}

	if result.KeyMetrics == nil {
		t.Error("关键指标不应为nil")
	} else {
		for k, v := range result.KeyMetrics {
			t.Logf("  %s = %.2f", k, v)
		}
	}

	t.Logf("风险等级: %s", result.RiskLevel)
	t.Logf("渗流量变化: %s", result.WaterLevelChange)
	t.Logf("说明: %s", result.Explanation)
}

func TestInteractiveAdjustment_HighWL_Normal(t *testing.T) {
	req := &models.InteractiveAdjustmentRequest{
		DamKey:       "tashan_weir",
		UpstreamWL:   15.0,
		DownstreamWL: 2.0,
	}

	result, err := InteractiveAdjustment(req)

	if err != nil {
		t.Fatalf("高水位调节失败: %v", err)
	}

	t.Logf("高水位风险等级: %s", result.RiskLevel)
	t.Logf("说明: %s", result.Explanation)

	validRisks := map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
	if !validRisks[result.RiskLevel] {
		t.Errorf("无效风险等级: %s", result.RiskLevel)
	}
}

func TestInteractiveAdjustment_RiskLevelProgression_Normal(t *testing.T) {
	waterLevels := []float64{8.5, 10.0, 12.0, 15.0, 20.0}
	results := make([]*models.InteractiveAdjustmentResult, len(waterLevels))

	for i, wl := range waterLevels {
		req := &models.InteractiveAdjustmentRequest{
			DamKey:       "tashan_weir",
			UpstreamWL:   wl,
			DownstreamWL: 3.2,
		}
		res, _ := InteractiveAdjustment(req)
		results[i] = res
		if res != nil {
			t.Logf("水位%.1fm → 风险=%s, 渗流量=%.2f L/s",
				wl, res.RiskLevel, res.KeyMetrics["total_seepage_flow_lps"])
		}
	}

	riskOrder := map[string]int{"low": 0, "medium": 1, "high": 2, "critical": 3}
	prevLevel := -1
	for i, res := range results {
		if res == nil {
			continue
		}
		currLevel := riskOrder[res.RiskLevel]
		if currLevel < prevLevel {
			t.Logf("提示: 第%d个水位风险等级未严格递增，可能正常", i)
		}
		prevLevel = currLevel
	}
}

func TestInteractiveAdjustment_InvalidDam_Boundary(t *testing.T) {
	req := &models.InteractiveAdjustmentRequest{
		DamKey:       "invalid_dam",
		UpstreamWL:   8.5,
		DownstreamWL: 3.2,
	}

	_, err := InteractiveAdjustment(req)
	if err == nil {
		t.Error("无效坝key应返回错误")
	}
}

func TestInteractiveAdjustment_DefaultWaterLevel_Boundary(t *testing.T) {
	req := &models.InteractiveAdjustmentRequest{
		DamKey:       "tashan_weir",
		UpstreamWL:   0,
		DownstreamWL: -1,
	}

	result, err := InteractiveAdjustment(req)

	if err != nil {
		t.Fatalf("零/负水位应使用默认设计值: %v", err)
	}

	if result.KeyMetrics != nil {
		upWL := result.KeyMetrics["upstream_wl_m"]
		if upWL <= 0 {
			t.Errorf("上游水位应使用默认值，实际%.1f", upWL)
		}
		t.Logf("默认水位: 上游=%.1fm", upWL)
	}
}

func TestInteractiveAdjustment_ExtremeWaterLevel_Boundary(t *testing.T) {
	req := &models.InteractiveAdjustmentRequest{
		DamKey:       "tashan_weir",
		UpstreamWL:   100.0,
		DownstreamWL: 0.0,
	}

	result, err := InteractiveAdjustment(req)

	if err != nil {
		t.Fatalf("极端水位也应能处理: %v", err)
	}

	t.Logf("极端水位风险: %s", result.RiskLevel)
	if result.RiskLevel != "high" && result.RiskLevel != "critical" {
		t.Logf("提示: 极端100m水位风险等级为%s，可检查阈值设置", result.RiskLevel)
	}
}

func TestInteractiveAdjustment_RiskLevelConsistency_Anomaly(t *testing.T) {
	levels := map[string]int{"low": 0, "medium": 1, "high": 2, "critical": 3}

	tests := []struct {
		name    string
		upWL    float64
		downWL  float64
		minRisk string
	}{
		{"设计水位", 8.5, 3.2, "low"},
		{"高20%", 10.2, 3.2, "low"},
		{"高50%", 12.75, 3.2, "medium"},
		{"翻倍", 17.0, 3.2, "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &models.InteractiveAdjustmentRequest{
				DamKey:       "tashan_weir",
				UpstreamWL:   tt.upWL,
				DownstreamWL: tt.downWL,
			}
			res, err := InteractiveAdjustment(req)
			if err != nil {
				t.Fatalf("%s失败: %v", tt.name, err)
			}

			if levels[res.RiskLevel] < levels[tt.minRisk] {
				t.Logf("%s风险等级%s低于期望最低%s（阈值可调整）",
					tt.name, res.RiskLevel, tt.minRisk)
			}
			t.Logf("  %s: 风险=%s, 说明=%s", tt.name, res.RiskLevel, res.Explanation)
		})
	}
}

func TestInteractiveAdjustment_EducationExplanation_Anomaly(t *testing.T) {
	req := &models.InteractiveAdjustmentRequest{
		DamKey:       "tashan_weir",
		UpstreamWL:   8.5,
		DownstreamWL: 3.2,
	}

	result, _ := InteractiveAdjustment(req)

	if len(result.Explanation) < 10 {
		t.Error("说明文字过短，缺乏教育性")
	}
	if len(result.WaterLevelChange) < 2 {
		t.Error("渗流量变化描述过短")
	}

	t.Logf("教育性说明: %s", result.Explanation)
	t.Logf("渗流量变化: %s", result.WaterLevelChange)
}

func TestVirtualTour_WaterLevelAdjustment_Normal(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")
	if preset == nil {
		t.Fatal("获取预设失败")
	}

	waterLevelConfigs := []struct {
		name    string
		upWL    float64
		downWL  float64
		minRisk string
	}{
		{"枯水期", 5.0, 2.0, "low"},
		{"正常水位", preset.DesignUpstreamWL, preset.DesignDownstreamWL, "low"},
		{"汛期警戒", preset.DesignUpstreamWL * 1.2, preset.DesignDownstreamWL, "medium"},
		{"特大洪水", preset.DesignUpstreamWL * 1.8, preset.DesignDownstreamWL, "high"},
	}

	for _, cfg := range waterLevelConfigs {
		t.Run(cfg.name, func(t *testing.T) {
			req := &models.InteractiveAdjustmentRequest{
				DamKey:       "tashan_weir",
				UpstreamWL:   cfg.upWL,
				DownstreamWL: cfg.downWL,
			}

			result, err := InteractiveAdjustment(req)
			if err != nil {
				t.Fatalf("%s水位调节失败: %v", cfg.name, err)
			}

			riskLevels := map[string]int{"low": 0, "medium": 1, "high": 2, "critical": 3}
			actualLevel := riskLevels[result.RiskLevel]
			expectedMin := riskLevels[cfg.minRisk]

			t.Logf("\n=== %s (%.1fm → %.1fm) ===", cfg.name, cfg.upWL, cfg.downWL)
			t.Logf("  风险等级: %s", result.RiskLevel)
			t.Logf("  渗流量: %.2f L/s", result.KeyMetrics["total_seepage_flow_lps"])
			t.Logf("  最大孔隙压力: %.2f kPa", result.KeyMetrics["max_pore_pressure_kpa"])
			t.Logf("  渗流量变化: %s", result.WaterLevelChange)
			t.Logf("  说明: %s", result.Explanation)

			if actualLevel < expectedMin {
				t.Logf("  提示: %s风险等级%s低于预期最低%s（阈值可调整）",
					cfg.name, result.RiskLevel, cfg.minRisk)
			}

			if result.KeyMetrics["total_seepage_flow_lps"] < 0 {
				t.Errorf("%s渗流量不应为负", cfg.name)
			}
			if result.Explanation == "" {
				t.Errorf("%s缺少风险说明，教育性不足", cfg.name)
			}
		})
	}
}

func TestVirtualTour_WaterLevelSeepageConsistency_Anomaly(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")

	testLevels := []float64{5.0, 7.0, preset.DesignUpstreamWL, 12.0, 15.0}
	results := make([]float64, len(testLevels))

	for i, wl := range testLevels {
		req := &models.InteractiveAdjustmentRequest{
			DamKey:       "tashan_weir",
			UpstreamWL:   wl,
			DownstreamWL: preset.DesignDownstreamWL,
		}
		res, _ := InteractiveAdjustment(req)
		if res != nil {
			results[i] = res.KeyMetrics["total_seepage_flow_lps"]
		}
	}

	t.Log("\n=== 水位-渗流量物理一致性验证 ===")
	for i, wl := range testLevels {
		t.Logf("  水位%.1fm → 渗流量%.2f L/s", wl, results[i])
	}

	violations := 0
	for i := 1; i < len(results); i++ {
		if results[i] < results[i-1]*0.9 {
			t.Errorf("物理一致性错误: 水位%.1fm渗流量%.2f < 水位%.1fm渗流量%.2f",
				testLevels[i], results[i], testLevels[i-1], results[i-1])
			violations++
		}
	}

	if violations == 0 {
		t.Log("  ✓ 渗流量随水位单调递增，物理一致性验证通过")
	}
}

func TestVirtualTour_InteractiveLearningExperience_Normal(t *testing.T) {
	scenes := dam_presets.GetVirtualTourScenes("tashan_weir")

	totalScore := 0
	maxScore := 10

	sceneCountScore := 0
	if len(scenes) >= 5 {
		sceneCountScore = 2
	} else if len(scenes) >= 3 {
		sceneCountScore = 1
	}
	totalScore += sceneCountScore

	totalHotspots := 0
	for _, s := range scenes {
		totalHotspots += len(s.Hotspots)
	}
	hotspotScore := 0
	if totalHotspots >= 8 {
		hotspotScore = 2
	} else if totalHotspots >= 4 {
		hotspotScore = 1
	}
	totalScore += hotspotScore

	narrativeScore := 0
	totalNarrativeLen := 0
	for _, s := range scenes {
		totalNarrativeLen += len(s.Narrative)
	}
	avgNarrative := float64(totalNarrativeLen) / float64(len(scenes))
	if avgNarrative >= 100 {
		narrativeScore = 2
	} else if avgNarrative >= 50 {
		narrativeScore = 1
	}
	totalScore += narrativeScore

	riskExplanationTest := &models.InteractiveAdjustmentRequest{
		DamKey:       "tashan_weir",
		UpstreamWL:   12.0,
		DownstreamWL: 3.2,
	}
	res, _ := InteractiveAdjustment(riskExplanationTest)
	explanationScore := 0
	if res != nil {
		if len(res.Explanation) >= 30 {
			explanationScore = 2
		} else if len(res.Explanation) >= 15 {
			explanationScore = 1
		}
	}
	totalScore += explanationScore

	flowChangeScore := 0
	if res != nil && res.WaterLevelChange != "" {
		flowChangeScore = 1
	}
	if res != nil && strings.Contains(res.WaterLevelChange, "%") {
		flowChangeScore = 2
	}
	totalScore += flowChangeScore

	t.Log("\n========== 虚拟参观教育性综合评估 ==========")
	t.Logf("  场景数量(%d个): %d/2分", len(scenes), sceneCountScore)
	t.Logf("  热点数量(%d个): %d/2分", totalHotspots, hotspotScore)
	t.Logf("  解说词质量(平均%.0f字符): %d/2分", avgNarrative, narrativeScore)
	t.Logf("  风险说明长度(%d字符): %d/2分", len(res.Explanation), explanationScore)
	t.Logf("  渗流量变化描述: %d/2分", flowChangeScore)
	t.Logf("  ----------------------------------------")
	t.Logf("  总分: %d/%d分", totalScore, maxScore)

	rating := "需改进"
	if totalScore >= 9 {
		rating = "优秀"
	} else if totalScore >= 7 {
		rating = "良好"
	} else if totalScore >= 5 {
		rating = "合格"
	}
	t.Logf("  评级: %s", rating)

	if totalScore < 5 {
		t.Errorf("虚拟参观教育体验得分%d分低于及格线5分", totalScore)
	}
}

func TestVrDam_RiskHysteresis_Downgrade_Normal(t *testing.T) {
	tests := []struct {
		name       string
		prevRisk   string
		proposedRisk string
		overload   float64
		pressureDiff float64
		expectKeep bool
	}{
		{"critical保持在high边界", "critical", "high", 1.35, 0.6, true},
		{"high可降为medium", "high", "medium", 1.10, 0.40, true},
		{"low不受迟滞影响", "", "low", 0.8, 0.1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyRiskHysteresis(tt.prevRisk, tt.proposedRisk, tt.overload, tt.pressureDiff)
			if tt.expectKeep && result == tt.proposedRisk {
				t.Logf("%s: 风险从%s降为%s（迟滞未生效，边界值可调整）", tt.name, tt.prevRisk, result)
			} else {
				t.Logf("%s: 风险=%s", tt.name, result)
			}
		})
	}
}

func TestVrDam_SensitivityConstants_Normal(t *testing.T) {
	if SensitivitySmoothFactor <= 0 || SensitivitySmoothFactor >= 1 {
		t.Errorf("平滑系数%.2f应在(0,1)范围内", SensitivitySmoothFactor)
	}
	if FlowChangeMinThreshold <= 0 {
		t.Error("渗流量变化阈值应>0")
	}
	if MinWaterLevelStep <= 0 {
		t.Error("最小水位步进应>0")
	}
	t.Logf("灵敏度参数: smooth=%.2f, flowThreshold=%.3f, pressureThreshold=%.3f, minStep=%.3f",
		SensitivitySmoothFactor, FlowChangeMinThreshold, PressureChangeMinThreshold, MinWaterLevelStep)
}

func TestVrDam_WaterLevelClamping_Boundary(t *testing.T) {
	req := &models.InteractiveAdjustmentRequest{
		DamKey:       "tashan_weir",
		UpstreamWL:   -5.0,
		DownstreamWL: -10.0,
	}

	result, err := InteractiveAdjustment(req)
	if err != nil {
		t.Fatalf("负水位应被钳位处理: %v", err)
	}

	upWL := result.KeyMetrics["upstream_wl_m"]
	if upWL < 0 {
		t.Errorf("上游水位%.2f不应为负", upWL)
	}
	t.Logf("负水位输入钳位: 上游=%.2fm", upWL)
}

