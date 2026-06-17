package era_comparator

import (
	"math"
	"strings"
	"testing"

	"tashan-weir-seepage/internal/models"
)

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

func TestEraComparator_TechnologyGapAccuracy_Normal(t *testing.T) {
	req := &models.CrossEraComparisonRequest{
		AncientDamKey:   "tashan_weir",
		ModernDamKey:    "modern_gravity",
		UpstreamWL:      8.5,
		DownstreamWL:    3.2,
		ScaleToSameSize: false,
	}

	result, err := CrossEraComparison(req)
	if err != nil {
		t.Fatalf("对比失败: %v", err)
	}

	if result.Comparison != nil {
		if gap, ok := result.Comparison["technology_gap_years"].(int); ok {
			if gap < 500 {
				t.Errorf("技术跨度%d年不合理，它山堰vs现代坝应>500年", gap)
			}
			t.Logf("技术跨度验证: %d年", gap)
		}

		if flowRed, ok := result.Comparison["seepage_flow_reduction_pct"].(float64); ok {
			if flowRed < 0 {
				t.Error("渗流量减少率不应为负")
			}
			t.Logf("渗流量减少: %.1f%%", flowRed)
		}
	}
}

func TestEraComparator_ModernDamStandards_Normal(t *testing.T) {
	req := &models.CrossEraComparisonRequest{
		AncientDamKey:   "tashan_weir",
		ModernDamKey:    "modern_gravity",
		UpstreamWL:      8.5,
		DownstreamWL:    3.2,
	}

	result, err := CrossEraComparison(req)
	if err != nil {
		t.Fatalf("对比失败: %v", err)
	}

	if result.ModernMetrics == nil {
		t.Error("现代坝指标不应为nil")
	}

	if result.ModernDam.AntiSeepageEfficiency > 0 {
		t.Logf("现代坝防渗效率: %.1f%%", result.ModernDam.AntiSeepageEfficiency)
	}
}
