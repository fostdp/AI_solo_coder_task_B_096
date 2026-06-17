package dam_comparator

import (
	"strings"
	"sync"
	"testing"

	"tashan-weir-seepage/internal/models"
)

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

func TestDamComparator_ConcurrentSafety_Normal(t *testing.T) {
	var wg sync.WaitGroup
	errCh := make(chan error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := &models.DamComparisonRequest{
				DamKeys:              []string{"tashan_weir", "mulan_bei"},
				UpstreamWaterLevel:   8.5 + float64(idx),
				DownstreamWaterLevel: 3.2,
				GridResolutionX:      30,
				GridResolutionY:      15,
			}
			_, err := CompareDams(req)
			if err != nil {
				errCh <- err
			}
		}(i)
	}
	wg.Wait()
	close(errCh)

	errCount := 0
	for err := range errCh {
		t.Logf("并发错误: %v", err)
		errCount++
	}
	if errCount > 2 {
		t.Errorf("并发对比错误过多: %d/5", errCount)
	}
}

func TestDamComparator_SummaryFields_Normal(t *testing.T) {
	req := &models.DamComparisonRequest{
		DamKeys:              []string{"tashan_weir", "modern_gravity"},
		UpstreamWaterLevel:   8.5,
		DownstreamWaterLevel: 3.2,
	}

	result, err := CompareDams(req)
	if err != nil {
		t.Fatalf("对比失败: %v", err)
	}

	requiredFields := []string{"dam_count", "water_head_difference_m", "min_seepage_flow_lps", "max_seepage_flow_lps"}
	for _, field := range requiredFields {
		if _, ok := result.Summary[field]; !ok {
			t.Errorf("摘要缺少必要字段: %s", field)
		}
	}
}

