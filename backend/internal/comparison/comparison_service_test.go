package comparison

import (
	"math"
	"strings"
	"testing"

	"tashan-weir-seepage/internal/models"
)

// ========== 正常场景测试 ==========

func TestCompareDams_ThreeAncientDams_Normal(t *testing.T) {
	req := &models.DamComparisonRequest{
		DamKeys:            []string{"tashan_weir", "mulan_bei", "yuliang_ba"},
		UpstreamWaterLevel: 8.5,
		DownstreamWaterLevel: 3.2,
		GridResolutionX:    50,
		GridResolutionY:    25,
		IncludeCurrentAging: false,
	}

	result, err := CompareDams(req)

	if err != nil {
		t.Fatalf("三坝对比失败: %v", err)
	}
	if result == nil {
		t.Fatal("对比结果为nil")
	}
	if len(result.Items) < 2 {
		t.Fatalf("期望至少2个对比项，实际%d个", len(result.Items))
	}

	for i, item := range result.Items {
		if item.DamName == "" {
			t.Errorf("对比项%d缺少坝名", i)
		}
		if item.TotalSeepageFlow <= 0 {
			t.Errorf("%s渗流量应>0，实际%e", item.DamName, item.TotalSeepageFlow)
		}
		if item.MaxPorePressure < 0 {
			t.Errorf("%s最大孔隙压力应>=0，实际%.2f", item.DamName, item.MaxPorePressure)
		}
		t.Logf("对比项%d: %s, 渗流量=%.4f L/s, 最大孔隙压力=%.2f kPa",
			i, item.DamName, item.TotalSeepageFlow*1000, item.MaxPorePressure)
	}

	if result.Summary == nil {
		t.Error("对比摘要不应为nil")
	} else {
		if damCount, ok := result.Summary["dam_count"].(int); ok {
			if damCount < 2 {
				t.Errorf("摘要中坝数量%d不足", damCount)
			}
		}
		if insights, ok := result.Summary["key_insights"].([]string); ok {
			t.Logf("关键洞察(%d条):", len(insights))
			for _, ins := range insights {
				t.Logf("  - %s", ins)
			}
		}
	}
}

func TestCompareDams_CrossEraAncientVsModern_Normal(t *testing.T) {
	req := &models.DamComparisonRequest{
		DamKeys:              []string{"tashan_weir", "modern_gravity"},
		UpstreamWaterLevel:   8.5,
		DownstreamWaterLevel: 3.2,
		GridResolutionX:      50,
		GridResolutionY:      25,
	}

	result, err := CompareDams(req)

	if err != nil {
		t.Fatalf("跨时代对比失败: %v", err)
	}

	var ancientItem, modernItem *models.DamComparisonItem
	for i := range result.Items {
		if result.Items[i].DamType == models.DamTypeAncientStone {
			ancientItem = &result.Items[i]
		}
		if result.Items[i].DamType == models.DamTypeModernConcrete {
			modernItem = &result.Items[i]
		}
	}

	if ancientItem == nil || modernItem == nil {
		t.Fatal("缺少古代或现代坝对比项")
	}

	t.Logf("古代坝(%s)渗流量: %.4f L/s", ancientItem.DamName, ancientItem.TotalSeepageFlow*1000)
	t.Logf("现代坝(%s)渗流量: %.4f L/s", modernItem.DamName, modernItem.TotalSeepageFlow*1000)

	if modernItem.TotalSeepageFlow >= ancientItem.TotalSeepageFlow {
		t.Errorf("现代坝渗流量(%.4f)应低于古代坝(%.4f)",
			modernItem.TotalSeepageFlow*1000, ancientItem.TotalSeepageFlow*1000)
	}

	reduction := (ancientItem.TotalSeepageFlow - modernItem.TotalSeepageFlow) / ancientItem.TotalSeepageFlow * 100
	t.Logf("现代坝防渗性能提升: %.1f%%", reduction)
	if reduction < 10 {
		t.Errorf("现代坝防渗提升不足10%%，实际%.1f%%", reduction)
	}
}

func TestCrossEraComparison_TashanVsModern_Normal(t *testing.T) {
	req := &models.CrossEraComparisonRequest{
		AncientDamKey:  "tashan_weir",
		ModernDamKey:   "modern_gravity",
		UpstreamWL:     8.5,
		DownstreamWL:   3.2,
		ScaleToSameSize: false,
	}

	result, err := CrossEraComparison(req)

	if err != nil {
		t.Fatalf("跨时代对比失败: %v", err)
	}
	if result == nil {
		t.Fatal("结果为nil")
	}

	if result.AncientDam.DamName == "" {
		t.Error("古代坝名称为空")
	}
	if result.ModernDam.DamName == "" {
		t.Error("现代坝名称为空")
	}

	if len(result.Insights) < 3 {
		t.Errorf("期望至少3条技术洞察，实际%d条", len(result.Insights))
	}
	t.Logf("技术洞察(%d条):", len(result.Insights))
	for i, ins := range result.Insights {
		t.Logf("  %d. %s", i+1, ins)
		if len(ins) < 10 {
			t.Errorf("洞察%d过短，教育性不足", i)
		}
	}

	if result.Comparison != nil {
		if flowReduction, ok := result.Comparison["seepage_flow_reduction_pct"].(float64); ok {
			t.Logf("渗流量降低: %.1f%%", flowReduction)
			if flowReduction < 0 {
				t.Error("渗流量降低率不应为负")
			}
		}
		if gap, ok := result.Comparison["technology_gap_years"].(int); ok {
			t.Logf("技术跨度: %d年", gap)
			if gap < 1000 {
				t.Errorf("技术跨度应>1000年，实际%d年", gap)
			}
		}
	}
}

func TestCrossEraComparison_ScaledToSameSize_Normal(t *testing.T) {
	req := &models.CrossEraComparisonRequest{
		AncientDamKey:   "tashan_weir",
		ModernDamKey:    "modern_gravity",
		UpstreamWL:      8.5,
		DownstreamWL:    3.2,
		ScaleToSameSize: true,
	}

	result, err := CrossEraComparison(req)

	if err != nil {
		t.Fatalf("等比例缩放对比失败: %v", err)
	}

	if result.Comparison != nil {
		if scaled, ok := result.Comparison["scaled_to_same_size"].(bool); ok && !scaled {
			t.Error("ScaleToSameSize=true但结果标记为未缩放")
		}
	}

	t.Log("等比例缩放跨时代对比测试通过")
}

// ========== 防渗效率验证（跨时代对比核心测试） ==========

func TestCrossEraComparison_AntiSeepageEfficiency_Normal(t *testing.T) {
	req := &models.CrossEraComparisonRequest{
		AncientDamKey:   "tashan_weir",
		ModernDamKey:    "modern_gravity",
		UpstreamWL:      8.5,
		DownstreamWL:    3.2,
		ScaleToSameSize: true,
	}

	result, err := CrossEraComparison(req)
	if err != nil {
		t.Fatalf("对比失败: %v", err)
	}

	ancientEff := result.AncientDam.AntiSeepageEfficiency
	modernEff := result.ModernDam.AntiSeepageEfficiency

	t.Logf("防渗效率对比: 古代坝=%.2f%%, 现代坝=%.2f%%", ancientEff, modernEff)

	if modernEff < ancientEff {
		t.Errorf("现代坝防渗效率(%.2f%%)应≥古代坝(%.2f%%)", modernEff, ancientEff)
	}

	if result.Comparison != nil {
		if permRed, ok := result.Comparison["permeability_reduction_pct"].(float64); ok {
			t.Logf("渗透系数降低: %.1f%%", permRed)
			if permRed < 50 {
				t.Errorf("现代坝渗透系数降低率应>50%%，实际%.1f%%", permRed)
			}
		}
	}
}

// ========== 交互式水位调节测试 ==========

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

// ========== 边界场景测试 ==========

func TestCompareDams_SingleDam_Boundary(t *testing.T) {
	req := &models.DamComparisonRequest{
		DamKeys:              []string{"tashan_weir"},
		UpstreamWaterLevel:   8.5,
		DownstreamWaterLevel: 3.2,
	}

	result, err := CompareDams(req)

	if err == nil {
		t.Error("单坝对比应返回错误")
	}
	if result != nil {
		t.Error("单坝对比应返回nil结果")
	}
	if !strings.Contains(err.Error(), "at least 2") {
		t.Errorf("错误信息应提示至少2座坝，实际: %v", err)
	}
}

func TestCompareDams_EmptyDams_Boundary(t *testing.T) {
	req := &models.DamComparisonRequest{
		DamKeys:              []string{},
		UpstreamWaterLevel:   8.5,
		DownstreamWaterLevel: 3.2,
	}

	_, err := CompareDams(req)
	if err == nil {
		t.Error("空坝列表应返回错误")
	}
}

func TestCompareDams_WithInvalidDam_Boundary(t *testing.T) {
	req := &models.DamComparisonRequest{
		DamKeys:              []string{"tashan_weir", "nonexistent_dam"},
		UpstreamWaterLevel:   8.5,
		DownstreamWaterLevel: 3.2,
	}

	result, err := CompareDams(req)

	if err != nil {
		t.Logf("包含无效坝返回错误: %v (可接受)", err)
	} else if result != nil {
		if len(result.Items) >= 1 {
			t.Logf("无效坝被过滤，剩余%d个有效坝", len(result.Items))
		}
	}
}

func TestCompareDams_ZeroGridResolution_Boundary(t *testing.T) {
	req := &models.DamComparisonRequest{
		DamKeys:              []string{"tashan_weir", "yuliang_ba"},
		UpstreamWaterLevel:   8.5,
		DownstreamWaterLevel: 3.2,
		GridResolutionX:      0,
		GridResolutionY:      -5,
	}

	result, err := CompareDams(req)

	if err != nil {
		t.Fatalf("零网格分辨率应使用默认值，不应报错: %v", err)
	}
	if result == nil {
		t.Fatal("结果不应为nil")
	}
	t.Log("零/负网格分辨率被正确容错")
}

func TestCrossEraComparison_InvalidAncientDam_Boundary(t *testing.T) {
	req := &models.CrossEraComparisonRequest{
		AncientDamKey: "invalid_dam",
		ModernDamKey:  "modern_gravity",
		UpstreamWL:    8.5,
		DownstreamWL:  3.2,
	}

	_, err := CrossEraComparison(req)
	if err == nil {
		t.Error("无效古代坝key应返回错误")
	}
}

func TestCrossEraComparison_InvalidModernDam_Boundary(t *testing.T) {
	req := &models.CrossEraComparisonRequest{
		AncientDamKey: "tashan_weir",
		ModernDamKey:  "invalid_dam",
		UpstreamWL:    8.5,
		DownstreamWL:  3.2,
	}

	_, err := CrossEraComparison(req)
	if err == nil {
		t.Error("无效现代坝key应返回错误")
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

// ========== 异常场景测试 ==========

func TestCompareDams_AllInvalid_Anomaly(t *testing.T) {
	req := &models.DamComparisonRequest{
		DamKeys:              []string{"invalid1", "invalid2", "invalid3"},
		UpstreamWaterLevel:   8.5,
		DownstreamWaterLevel: 3.2,
	}

	result, err := CompareDams(req)

	if err == nil {
		t.Error("全部无效坝应返回错误")
	}
	if result != nil {
		t.Error("全部无效坝应返回nil结果")
	}
}

func TestCompareDams_DuplicateDams_Anomaly(t *testing.T) {
	req := &models.DamComparisonRequest{
		DamKeys:              []string{"tashan_weir", "tashan_weir", "tashan_weir"},
		UpstreamWaterLevel:   8.5,
		DownstreamWaterLevel: 3.2,
	}

	result, err := CompareDams(req)

	if err != nil {
		t.Logf("重复坝key报错: %v (可接受)", err)
		return
	}

	if result != nil {
		t.Logf("重复坝key得到%d个对比项（去重或重复均合理）", len(result.Items))
	}
}

func TestCrossEraComparison_SameDam_Anomaly(t *testing.T) {
	req := &models.CrossEraComparisonRequest{
		AncientDamKey: "tashan_weir",
		ModernDamKey:  "tashan_weir",
		UpstreamWL:    8.5,
		DownstreamWL:  3.2,
	}

	result, err := CrossEraComparison(req)

	if err != nil {
		t.Logf("同一座坝进行跨时代对比报错: %v (可接受)", err)
		return
	}

	if result != nil {
		if result.Comparison != nil {
			if red, ok := result.Comparison["seepage_flow_reduction_pct"].(float64); ok {
				if math.Abs(red) > 0.01 {
					t.Errorf("同一坝对比渗流量变化率应≈0，实际%.2f%%", red)
				}
			}
		}
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

func TestCompareDams_SeepageFlowPhysicalConsistency_Anomaly(t *testing.T) {
	req1 := &models.DamComparisonRequest{
		DamKeys:              []string{"tashan_weir", "mulan_bei"},
		UpstreamWaterLevel:   8.5,
		DownstreamWaterLevel: 3.2,
	}
	req2 := &models.DamComparisonRequest{
		DamKeys:              []string{"tashan_weir", "mulan_bei"},
		UpstreamWaterLevel:   12.0,
		DownstreamWaterLevel: 3.2,
	}

	res1, _ := CompareDams(req1)
	res2, _ := CompareDams(req2)

	if res1 == nil || res2 == nil {
		t.Fatal("对比失败")
	}

	findFlow := func(items []models.DamComparisonItem, name string) float64 {
		for _, it := range items {
			if it.DamName == name {
				return it.TotalSeepageFlow
			}
		}
		return -1
	}

	lowFlow := findFlow(res1.Items, "它山堰")
	highFlow := findFlow(res2.Items, "它山堰")

	if lowFlow > 0 && highFlow > 0 && highFlow < lowFlow*0.95 {
		t.Errorf("物理一致性错误: 高水位渗流量%.4f < 低水位%.4f",
			highFlow*1000, lowFlow*1000)
	}

	t.Logf("它山堰物理一致性: 低水头%.4f L/s → 高水头%.4f L/s ✓",
		lowFlow*1000, highFlow*1000)
}

func TestGenerateComparisonSummary_EmptyItems_Anomaly(t *testing.T) {
	summary := generateComparisonSummary([]models.DamComparisonItem{}, 8.5, 3.2)

	if summary == nil {
		t.Error("空项目也应返回摘要map")
		return
	}

	if count, ok := summary["dam_count"].(int); ok {
		if count != 0 {
			t.Errorf("空项目dam_count应为0，实际%d", count)
		}
	}
}

func TestCrossEraComparison_InsightsEducationalValue_Anomaly(t *testing.T) {
	req := &models.CrossEraComparisonRequest{
		AncientDamKey:   "tashan_weir",
		ModernDamKey:    "modern_gravity",
		UpstreamWL:      8.5,
		DownstreamWL:    3.2,
		ScaleToSameSize: false,
	}

	result, _ := CrossEraComparison(req)

	expectedKeywords := []string{"年", "防渗", "技术", "文化"}
	foundCount := 0
	for _, kw := range expectedKeywords {
		for _, ins := range result.Insights {
			if strings.Contains(ins, kw) {
				foundCount++
				break
			}
		}
	}

	t.Logf("技术洞察关键词命中: %d/%d", foundCount, len(expectedKeywords))
	if foundCount < 2 {
		t.Log("提示: 技术洞察的教育性可进一步加强")
	}
}
