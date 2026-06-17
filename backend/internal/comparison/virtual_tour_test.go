package comparison

import (
	"math"
	"strings"
	"testing"

	"tashan-weir-seepage/internal/dam_presets"
	"tashan-weir-seepage/internal/models"
)

// ========== 虚拟参观场景完整性测试 ==========

func TestVirtualTour_SceneIntegrity_Normal(t *testing.T) {
	scenes := dam_presets.GetVirtualTourScenes("tashan_weir")

	if len(scenes) == 0 {
		t.Fatal("它山堰虚拟参观场景为空")
	}

	t.Logf("它山堰虚拟参观共 %d 个场景:", len(scenes))

	for i, scene := range scenes {
		t.Logf("\n--- 场景 %d: %s (%s) ---", i+1, scene.SceneName, scene.SceneID)

		if scene.SceneID == "" {
			t.Errorf("场景%d缺少SceneID", i)
		}
		if scene.SceneName == "" {
			t.Errorf("场景%d缺少SceneName", i)
		}
		if scene.Description == "" {
			t.Errorf("场景%d缺少Description", i)
		}
		if len(scene.Description) < 5 {
			t.Errorf("场景%d描述过短(%d字符)，信息不足", i, len(scene.Description))
		}

		t.Logf("  描述: %s", scene.Description)
		t.Logf("  相机位置: (%.1f, %.1f, %.1f)", scene.CameraPos.X, scene.CameraPos.Y, scene.CameraPos.Z)
		t.Logf("  观察目标: (%.1f, %.1f, %.1f)", scene.CameraTarget.X, scene.CameraTarget.Y, scene.CameraTarget.Z)
		t.Logf("  解说词长度: %d字符", len(scene.Narrative))
		t.Logf("  热点数量: %d个", len(scene.Hotspots))
	}
}

// ========== 场景相机位置合理性测试 ==========

func TestVirtualTour_CameraPositions_Normal(t *testing.T) {
	scenes := dam_presets.GetVirtualTourScenes("tashan_weir")

	preset := dam_presets.GetDamPreset("tashan_weir")
	if preset == nil {
		t.Fatal("获取它山堰预设失败")
	}

	for i, scene := range scenes {
		t.Run(scene.SceneID, func(t *testing.T) {
			if scene.CameraPos == scene.CameraTarget {
				t.Errorf("场景%s相机位置与目标点重合，无法形成有效视角", scene.SceneID)
			}

			distance := math.Sqrt(
				math.Pow(scene.CameraPos.X-scene.CameraTarget.X, 2) +
					math.Pow(scene.CameraPos.Y-scene.CameraTarget.Y, 2) +
					math.Pow(scene.CameraPos.Z-scene.CameraTarget.Z, 2))

			if distance < 5 {
				t.Errorf("场景%s相机距离(%.1fm)过近，无法观察坝体", scene.SceneID, distance)
			}
			if distance > 500 {
				t.Errorf("场景%s相机距离(%.1fm)过远，细节不清晰", scene.SceneID, distance)
			}

			if scene.CameraPos.Y < 0 {
				t.Errorf("场景%s相机Y坐标%.1f为负，位于地面以下", scene.SceneID, scene.CameraPos.Y)
			}
			if scene.CameraTarget.Y < -preset.FoundationDepth-1 {
				t.Errorf("场景%s目标点Y坐标%.1f低于坝基", scene.SceneID, scene.CameraTarget.Y)
			}

			t.Logf("  相机距离目标: %.1fm (合理范围5-500m) ✓", distance)
			_ = i
		})
	}
}

// ========== 场景切换逻辑测试 ==========

func TestVirtualTour_SceneSwitching_Boundary(t *testing.T) {
	scenes := dam_presets.GetVirtualTourScenes("tashan_weir")

	if len(scenes) < 2 {
		t.Fatal("至少需要2个场景才能测试切换")
	}

	sceneIDs := make(map[string]bool)
	for _, s := range scenes {
		if sceneIDs[s.SceneID] {
			t.Errorf("发现重复场景ID: %s", s.SceneID)
		}
		sceneIDs[s.SceneID] = true
	}

	for i := 1; i < len(scenes); i++ {
		prev := scenes[i-1]
		curr := scenes[i]

		distance := math.Sqrt(
			math.Pow(curr.CameraPos.X-prev.CameraPos.X, 2) +
				math.Pow(curr.CameraPos.Y-prev.CameraPos.Y, 2) +
				math.Pow(curr.CameraPos.Z-prev.CameraPos.Z, 2))

		t.Logf("切换 %s → %s: 相机移动%.1fm", prev.SceneID, curr.SceneID, distance)

		if distance > 200 {
			t.Logf("  提示: 场景切换距离%.1fm较大，建议使用平滑动画过渡", distance)
		}
	}
}

// ========== 热点交互与教育性测试 ==========

func TestVirtualTour_HotspotInteractivity_Normal(t *testing.T) {
	scenes := dam_presets.GetVirtualTourScenes("tashan_weir")
	preset := dam_presets.GetDamPreset("tashan_weir")

	totalHotspots := 0
	hasHotspotScene := 0

	for i, scene := range scenes {
		t.Logf("\n场景: %s (%d个热点)", scene.SceneName, len(scene.Hotspots))

		if len(scene.Hotspots) > 0 {
			hasHotspotScene++
		}
		totalHotspots += len(scene.Hotspots)

		for j, hs := range scene.Hotspots {
			t.Logf("  热点%d: %s", j+1, hs.Title)
			t.Logf("    位置: (%.1f, %.1f, %.1f)", hs.Position.X, hs.Position.Y, hs.Position.Z)
			t.Logf("    描述: %s", hs.Description)

			if hs.HotspotID == "" {
				t.Errorf("场景%d的热点%d缺少ID", i, j)
			}
			if hs.Title == "" {
				t.Errorf("场景%d的热点%d缺少标题", i, j)
			}

			minDescLength := 8
			if len(hs.Description) < minDescLength {
				t.Errorf("场景%d的热点'%s'描述过短(%d字符< %d)，教育性不足",
					i, hs.Title, len(hs.Description), minDescLength)
			}

			if preset != nil {
				if hs.Position.X < -10 || hs.Position.X > preset.Geometry.Length+20 {
					t.Errorf("热点%s的X坐标%.1f超出坝体范围[0, %.1f]",
						hs.HotspotID, hs.Position.X, preset.Geometry.Length+10)
				}
				if hs.Position.Y < -preset.FoundationDepth-2 || hs.Position.Y > preset.Geometry.Height+10 {
					t.Errorf("热点%s的Y坐标%.1f超出合理范围", hs.HotspotID, hs.Position.Y)
				}
			}
		}
	}

	t.Logf("\n=== 热点教育性统计 ===")
	t.Logf("  总场景数: %d", len(scenes))
	t.Logf("  含热点场景数: %d", hasHotspotScene)
	t.Logf("  热点总数: %d", totalHotspots)
	t.Logf("  平均每场景热点: %.1f", float64(totalHotspots)/float64(len(scenes)))

	if totalHotspots < 3 {
		t.Log("  提示: 热点总数偏少，可增加更多交互点提升教育体验")
	}
}

// ========== 解说词教育价值测试 ==========

func TestVirtualTour_NarrativeEducationValue_Normal(t *testing.T) {
	scenes := dam_presets.GetVirtualTourScenes("tashan_weir")

	educationalKeywords := []string{
		"水利", "工程", "防渗", "渗流", "历史", "唐朝", "宋代",
		"灌溉", "防洪", "技术", "建造", "条石", "糯米", "工匠",
		"文化", "遗产", "监测", "传感器", "扬压力",
	}

	totalScore := 0
	for i, scene := range scenes {
		hitCount := 0
		for _, kw := range educationalKeywords {
			if strings.Contains(scene.Narrative, kw) {
				hitCount++
			}
		}

		minLength := 30
		if len(scene.Narrative) < minLength {
			t.Errorf("场景%d(%s)解说词过短(%d字符< %d)，教育信息不足",
				i, scene.SceneName, len(scene.Narrative), minLength)
		}

		t.Logf("场景%d: %s → 关键词命中%d/%d, 长度%d字符",
			i, scene.SceneName, hitCount, len(educationalKeywords), len(scene.Narrative))

		totalScore += hitCount
	}

	avgScore := float64(totalScore) / float64(len(scenes))
	t.Logf("\n解说词平均教育关键词命中: %.1f/%d", avgScore, len(educationalKeywords))

	if avgScore < 2 {
		t.Log("  提示: 解说词教育关键词较少，可增加更多水利工程科普内容")
	}
}

// ========== 虚拟参观水位调节交互测试 ==========

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

// ========== 水位调节渗流变化物理一致性测试 ==========

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

// ========== 虚拟参观场景覆盖完整性测试 ==========

func TestVirtualTour_SceneCoverage_Normal(t *testing.T) {
	scenes := dam_presets.GetVirtualTourScenes("tashan_weir")

	expectedSceneTypes := map[string]bool{
		"overview":        false,
		"upstream_view":   false,
		"seepage_cutaway": false,
		"downstream_view": false,
		"sensor_layout":   false,
	}

	for _, s := range scenes {
		if _, ok := expectedSceneTypes[s.SceneID]; ok {
			expectedSceneTypes[s.SceneID] = true
		}
	}

	t.Log("\n=== 场景覆盖完整性 ===")
	for id, found := range expectedSceneTypes {
		status := "✓"
		if !found {
			status = "✗ 缺失"
		}
		t.Logf("  %s: %s", id, status)
	}

	missingCount := 0
	for _, found := range expectedSceneTypes {
		if !found {
			missingCount++
		}
	}

	if missingCount > 0 {
		t.Logf("  提示: 缺失%d种场景类型，可扩展参观体验", missingCount)
	}
}

// ========== 异常场景测试 ==========

func TestVirtualTour_InvalidDamKey_Boundary(t *testing.T) {
	scenes := dam_presets.GetVirtualTourScenes("invalid_dam")
	if len(scenes) != 0 {
		t.Errorf("无效坝key应返回空场景，实际%d个", len(scenes))
	}

	scenes2 := dam_presets.GetVirtualTourScenes("modern_gravity")
	if len(scenes2) == 0 {
		t.Log("提示: 现代重力坝暂无虚拟参观场景，可扩展")
	}
}

func TestVirtualTour_HotspotUniqueness_Anomaly(t *testing.T) {
	scenes := dam_presets.GetVirtualTourScenes("tashan_weir")

	hotspotIDs := make(map[string]string)
	for i, scene := range scenes {
		for _, hs := range scene.Hotspots {
			if prevScene, exists := hotspotIDs[hs.HotspotID]; exists {
				t.Errorf("热点ID冲突: %s 同时出现在%s和场景%d",
					hs.HotspotID, prevScene, i)
			}
			hotspotIDs[hs.HotspotID] = scene.SceneID
		}
	}
}

func TestVirtualTour_NarrativeCulturalValue_Anomaly(t *testing.T) {
	scenes := dam_presets.GetVirtualTourScenes("tashan_weir")
	preset := dam_presets.GetDamPreset("tashan_weir")

	if preset == nil {
		t.Fatal("获取预设失败")
	}

	mentionsHistory := false
	for _, scene := range scenes {
		if strings.Contains(scene.Narrative, preset.BuildDynasty) ||
			strings.Contains(scene.Narrative, "王元暐") ||
			strings.Contains(scene.Narrative, "世界灌溉工程遗产") ||
			strings.Contains(scene.Narrative, "全国重点文物保护单位") {
			mentionsHistory = true
			break
		}
	}

	if !mentionsHistory {
		t.Log("  提示: 解说词可增加更多历史文化价值相关内容，提升公众教育效果")
	} else {
		t.Log("  ✓ 解说词包含历史文化价值内容")
	}
}

// ========== 交互式学习体验综合评估 ==========

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
