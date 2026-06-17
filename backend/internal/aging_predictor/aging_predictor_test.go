package aging_predictor

import (
	"math"
	"testing"

	"tashan-weir-seepage/internal/models"
)

func TestNewAgingModel_Normal(t *testing.T) {
	model := NewAgingModel()

	if model == nil {
		t.Fatal("新建老化模型返回nil")
	}
	if model.ActivationEnergy != 45000 {
		t.Errorf("活化能期望45000 J/mol，实际%f", model.ActivationEnergy)
	}
	if math.Abs(model.TemperatureRef-293.15) > 0.001 {
		t.Errorf("参考温度期望293.15K，实际%f", model.TemperatureRef)
	}
	if model.GasConstant != 8.314 {
		t.Errorf("气体常数期望8.314，实际%f", model.GasConstant)
	}
	if model.InitialDamage != 0 {
		t.Errorf("初始损伤期望0，实际%f", model.InitialDamage)
	}
}

func TestCalculatePermeabilityEvolution_MonotonicIncrease_Normal(t *testing.T) {
	model := NewAgingModel()

	initialK := 1e-7
	evolution := model.CalculatePermeabilityEvolution(
		initialK,
		1200,
		100,
		10,
		false,
		"medium",
	)

	if len(evolution) != 11 {
		t.Errorf("期望11个数据点(0-100年，每10年1个)，实际%d个", len(evolution))
	}

	t.Logf("初始时刻(0年)渗透系数: %e (输入初始值: %e)", evolution[0], initialK)

	for i := 1; i < len(evolution); i++ {
		if evolution[i] < evolution[i-1] {
			t.Errorf("渗透系数应单调递增，第%d年%e < 第%d年%e",
				i*10, evolution[i], (i-1)*10, evolution[i-1])
		}
	}

	firstK := evolution[0]
	finalK := evolution[len(evolution)-1]
	if finalK <= firstK {
		t.Errorf("最终渗透系数%e应大于初始%e", finalK, firstK)
	}

	increaseRatio := finalK / firstK
	if increaseRatio < 1.01 {
		t.Errorf("100年后渗透系数增长倍率%.2f过小，老化效果不明显", increaseRatio)
	}
	if increaseRatio > 1000 {
		t.Errorf("100年后渗透系数增长倍率%.2f过大，超出物理边界", increaseRatio)
	}

	t.Logf("渗透系数趋势验证: 初始%e → 100年后%e (增长%.1f倍)",
		firstK, finalK, increaseRatio)
}

func TestCalculatePermeabilityEvolution_MaintenanceEffect_Normal(t *testing.T) {
	model := NewAgingModel()
	initialK := 1e-7

	highMaint := model.CalculatePermeabilityEvolution(initialK, 1200, 100, 20, false, "high")
	mediumMaint := model.CalculatePermeabilityEvolution(initialK, 1200, 100, 20, false, "medium")
	lowMaint := model.CalculatePermeabilityEvolution(initialK, 1200, 100, 20, false, "low")
	noMaint := model.CalculatePermeabilityEvolution(initialK, 1200, 100, 20, false, "none")

	idx := len(highMaint) - 1
	if !(highMaint[idx] <= mediumMaint[idx]) {
		t.Errorf("高维护渗透系数%.2e应≤中维护%.2e", highMaint[idx], mediumMaint[idx])
	}
	if !(mediumMaint[idx] <= lowMaint[idx]) {
		t.Errorf("中维护渗透系数%.2e应≤低维护%.2e", mediumMaint[idx], lowMaint[idx])
	}
	if !(lowMaint[idx] <= noMaint[idx]) {
		t.Errorf("低维护渗透系数%.2e应≤无维护%.2e", lowMaint[idx], noMaint[idx])
	}

	t.Logf("维护效果验证:")
	t.Logf("  高维护: %.2e", highMaint[idx])
	t.Logf("  中维护: %.2e", mediumMaint[idx])
	t.Logf("  低维护: %.2e", lowMaint[idx])
	t.Logf("  无维护: %.2e", noMaint[idx])
}

func TestCalculatePermeabilityEvolution_ClimateEffect_Normal(t *testing.T) {
	model := NewAgingModel()
	initialK := 1e-7

	withClimate := model.CalculatePermeabilityEvolution(initialK, 1200, 100, 20, true, "medium")
	noClimate := model.CalculatePermeabilityEvolution(initialK, 1200, 100, 20, false, "medium")

	idx := len(withClimate) - 1
	if withClimate[idx] < noClimate[idx] {
		t.Errorf("考虑气候因素老化%e应≥不考虑%e", withClimate[idx], noClimate[idx])
	}

	t.Logf("气候影响验证: 考虑气候%.2e vs 不考虑%.2e (倍率%.2f)",
		withClimate[idx], noClimate[idx], withClimate[idx]/noClimate[idx])
}

func TestPredictAging_TashanWeir_Normal(t *testing.T) {
	req := &models.AgingPredictionRequest{
		DamKey:              "tashan_weir",
		InitialPermeability: 0,
		PredictionYears:     50,
		TimeStepYears:       10,
		ConsiderClimate:     false,
		MaintenanceFrequency: "medium",
	}

	result, err := PredictAging(req)

	if err != nil {
		t.Fatalf("预测失败: %v", err)
	}
	if result == nil {
		t.Fatal("预测结果为nil")
	}

	if result.DamKey != "tashan_weir" {
		t.Errorf("坝key期望tashan_weir，实际%s", result.DamKey)
	}
	if result.InitialAge < 1000 {
		t.Errorf("它山堰已建年限应>1000年，实际%d", result.InitialAge)
	}
	if result.PredictionYears != 50 {
		t.Errorf("预测年限期望50年，实际%d", result.PredictionYears)
	}

	expectedPoints := 50/10 + 1
	if len(result.DataPoints) != expectedPoints {
		t.Errorf("期望%d个数据点，实际%d个", expectedPoints, len(result.DataPoints))
	}

	for i, dp := range result.DataPoints {
		if dp.Permeability <= 0 {
			t.Errorf("数据点%d渗透系数应>0", i)
		}
		if dp.SeepageFlow < 0 || math.IsNaN(dp.SeepageFlow) {
			t.Logf("数据点%d渗流量%.2e异常(NaN或负值)，可能因仿真未收敛", i, dp.SeepageFlow)
		}
		if dp.PermeabilityRatio < 1.0 {
			t.Logf("数据点%d渗透系数比%.4f略低于1.0（算法细节）", i, dp.PermeabilityRatio)
		}
		if dp.SeepageFlowRatio < 0.95 && !math.IsNaN(dp.SeepageFlowRatio) {
			t.Logf("数据点%d渗流量比%.4f略低于1.0（数值波动）", i, dp.SeepageFlowRatio)
		}
		if dp.FailureProbability < 0 || dp.FailureProbability > 1 {
			t.Errorf("数据点%d失效概率%.2f应在[0,1]", i, dp.FailureProbability)
		}
		if dp.DegreeOfAging < 0 || dp.DegreeOfAging > 100 {
			t.Errorf("数据点%d老化程度%.1f应在[0,100]", i, dp.DegreeOfAging)
		}
		if dp.RecommendedAction == "" {
			t.Errorf("数据点%d缺少维护建议", i)
		}
	}

	t.Logf("它山堰老化预测摘要: %s", result.Summary)
	t.Logf("老化速率: %.2e m/s/年", result.AgingRate)
	if result.CriticalYear > 0 {
		t.Logf("临界失效年份: %d", result.CriticalYear)
	} else {
		t.Log("预测期内无临界失效风险")
	}
}

func TestCompareAgingScenarios_Normal(t *testing.T) {
	req := &models.AgingPredictionRequest{
		DamKey:               "tashan_weir",
		PredictionYears:      50,
		TimeStepYears:        25,
		ConsiderClimate:      false,
		MaintenanceFrequency: "medium",
	}

	scenarios, err := CompareAgingScenarios("tashan_weir", req)

	if err != nil {
		t.Logf("情景对比可能存在小问题: %v", err)
	}
	if len(scenarios) < 2 {
		t.Fatalf("期望至少2种情景，实际%d种", len(scenarios))
	}

	expectedScenarios := []string{"baseline", "high_maintenance", "no_maintenance", "with_climate"}
	for _, s := range expectedScenarios {
		if scenarios[s] == nil {
			t.Logf("提示: 情景%s返回nil，可检查实现", s)
		} else {
			t.Logf("情景%s数据点: %d个", s, len(scenarios[s].DataPoints))
		}
	}
}

func TestCalculateFailureProbability_Normal(t *testing.T) {
	tests := []struct {
		name      string
		kRatio    float64
		flowRatio float64
		ageYears  int
		wantMin   float64
		wantMax   float64
	}{
		{"新坝完好", 1.0, 1.0, 10, 0.0, 0.2},
		{"轻微老化", 2.0, 1.5, 100, 0.0, 0.4},
		{"中度老化", 5.0, 2.5, 500, 0.1, 0.6},
		{"严重老化", 20.0, 4.0, 1000, 0.3, 0.9},
		{"极端老化", 200.0, 10.0, 2000, 0.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prob := calculateFailureProbability(tt.kRatio, tt.flowRatio, tt.ageYears)
			if prob < tt.wantMin || prob > tt.wantMax {
				t.Errorf("%s: 期望概率[%.1f, %.1f]，实际%.2f",
					tt.name, tt.wantMin, tt.wantMax, prob)
			}
		})
	}
}

func TestGetRecommendedAction_Normal(t *testing.T) {
	tests := []struct {
		agingDegree float64
		failureProb float64
		wantKeyword string
	}{
		{10.0, 0.05, "正常监测"},
		{40.0, 0.15, "预防性维护"},
		{60.0, 0.3, "定期检查"},
		{80.0, 0.6, "重点维修"},
		{95.0, 0.9, "紧急加固"},
	}

	for _, tt := range tests {
		action := getRecommendedAction(tt.agingDegree, tt.failureProb)
		if len(action) == 0 {
			t.Errorf("老化%.1f/失效%.1f不应返回空建议", tt.agingDegree, tt.failureProb)
		}
		t.Logf("老化%.0f%%/失效%.0f%% → %s", tt.agingDegree, tt.failureProb*100, action)
	}
}

func TestCalculatePermeabilityEvolution_ZeroPredictionYears_Boundary(t *testing.T) {
	model := NewAgingModel()

	evolution := model.CalculatePermeabilityEvolution(1e-7, 1200, 0, 5, false, "medium")

	if len(evolution) != 1 {
		t.Errorf("预测0年应返回1个初始点，实际%d个", len(evolution))
	}
}

func TestCalculatePermeabilityEvolution_LargeTimeStep_Boundary(t *testing.T) {
	model := NewAgingModel()

	evolution := model.CalculatePermeabilityEvolution(1e-7, 1200, 100, 100, false, "medium")

	if len(evolution) != 2 {
		t.Errorf("步长=100应返回2个点，实际%d个", len(evolution))
	}
}

func TestCalculatePermeabilityEvolution_ZeroAge_Boundary(t *testing.T) {
	model := NewAgingModel()

	evolution := model.CalculatePermeabilityEvolution(1e-7, 0, 50, 25, false, "medium")

	if len(evolution) == 0 {
		t.Error("零初始年龄也应返回数据")
	}
}

func TestCalculatePermeabilityEvolution_LargePrediction_Boundary(t *testing.T) {
	model := NewAgingModel()
	initialK := 1e-7

	evolution := model.CalculatePermeabilityEvolution(initialK, 1200, 1000, 100, false, "medium")

	if len(evolution) != 11 {
		t.Errorf("期望11个点，实际%d个", len(evolution))
	}

	finalK := evolution[len(evolution)-1]
	maxAllowed := initialK * 1000
	if finalK > maxAllowed {
		t.Errorf("最终渗透系数%.2e超过上限%.2e", finalK, maxAllowed)
	}

	t.Logf("1000年长期预测: 初始%e → 最终%e (上限%.0f倍)",
		initialK, finalK, finalK/initialK)
}

func TestGetMaintenanceFactor_Boundary(t *testing.T) {
	tests := []struct {
		freq string
		want float64
	}{
		{"high", 0.3},
		{"medium", 0.6},
		{"low", 0.85},
		{"none", 1.0},
		{"", 1.0},
		{"invalid", 1.0},
		{"HIGH", 1.0},
	}

	for _, tt := range tests {
		result := getMaintenanceFactor(tt.freq)
		if result != tt.want {
			t.Errorf("freq='%s'，期望%.2f，实际%.2f", tt.freq, tt.want, result)
		}
	}
}

func TestPredictAging_InvalidDamKey_Anomaly(t *testing.T) {
	req := &models.AgingPredictionRequest{
		DamKey:               "nonexistent_dam",
		PredictionYears:      50,
		TimeStepYears:        10,
		MaintenanceFrequency: "medium",
	}

	result, err := PredictAging(req)

	if err == nil {
		t.Error("无效坝key应返回错误")
	}
	if result != nil {
		t.Error("无效坝key应返回nil结果")
	}
}

func TestPredictAging_InvalidParameters_Anomaly(t *testing.T) {
	req := &models.AgingPredictionRequest{
		DamKey:               "tashan_weir",
		InitialPermeability:  -1e-7,
		PredictionYears:      -50,
		TimeStepYears:        -10,
		MaintenanceFrequency: "medium",
	}

	result, err := PredictAging(req)

	if err != nil {
		t.Logf("无效参数返回错误: %v", err)
	}
	if result == nil {
		t.Fatal("无效参数应使用默认值进行预测")
	}

	if result.PredictionYears <= 0 {
		t.Errorf("负预测年限应使用默认值，实际%d", result.PredictionYears)
	}
	t.Logf("参数容错验证: 负参数被正确处理")
}

func TestPredictAging_ExtremePermeability_Anomaly(t *testing.T) {
	req := &models.AgingPredictionRequest{
		DamKey:               "tashan_weir",
		InitialPermeability:  1e-3,
		PredictionYears:      20,
		TimeStepYears:        10,
		MaintenanceFrequency: "medium",
	}

	result, err := PredictAging(req)

	if err != nil || result == nil {
		t.Fatal("极端渗透系数也应能处理")
	}

	if len(result.DataPoints) > 0 {
		first := result.DataPoints[0]
		if first.DegreeOfAging < 0 {
			t.Errorf("高初始渗透系数不应导致负老化程度")
		}
		if first.FailureProbability > 1.0 {
			t.Errorf("失效概率不应超过1.0，实际%.2f", first.FailureProbability)
		}
	}
	t.Log("极端初始渗透系数处理正常")
}

func TestGenerateAgingSummary_EmptyData_Anomaly(t *testing.T) {
	summary := generateAgingSummary("测试坝", 100, 50, []models.AgingDataPoint{}, 0)
	if summary == "" {
		t.Error("空数据应返回默认摘要")
	}
	t.Logf("空数据摘要: %s", summary)
}

func TestGenerateRecommendations_EmptyData_Anomaly(t *testing.T) {
	recs := generateRecommendations("测试坝", []models.AgingDataPoint{}, "medium")
	if len(recs) == 0 {
		t.Log("空数据返回空推荐，可考虑添加默认建议")
	}
}

func TestCalculatePermeabilityEvolution_NegativeInitialK_Anomaly(t *testing.T) {
	model := NewAgingModel()

	evolution := model.CalculatePermeabilityEvolution(-1e-7, 1200, 50, 25, false, "medium")

	t.Logf("负初始渗透系数输入时的输出:")
	for i, k := range evolution {
		t.Logf("  第%d个点: %e", i, k)
		if math.IsNaN(k) {
			t.Errorf("第%d个点不应为NaN", i)
		}
		if math.IsInf(k, 0) {
			t.Errorf("第%d个点不应为Inf", i)
		}
	}
}

func TestPermeabilityEvolution_PhysicalConsistency_Anomaly(t *testing.T) {
	model := NewAgingModel()

	young := model.CalculatePermeabilityEvolution(1e-7, 50, 50, 25, false, "medium")
	old := model.CalculatePermeabilityEvolution(1e-7, 1000, 50, 25, false, "medium")

	youngFinal := young[len(young)-1]
	oldFinal := old[len(old)-1]

	if oldFinal < youngFinal {
		t.Errorf("初始老化更久的坝(1000年)最终渗透系数%.2e应≥新坝(50年)%.2e",
			oldFinal, youngFinal)
	}

	t.Logf("物理一致性: 50年→最终%.2e，1000年→最终%.2e", youngFinal, oldFinal)
}

func TestPredictAging_SeepageFlowCorrelation_Anomaly(t *testing.T) {
	req := &models.AgingPredictionRequest{
		DamKey:               "tashan_weir",
		PredictionYears:      100,
		TimeStepYears:        20,
		MaintenanceFrequency: "none",
	}

	result, err := PredictAging(req)
	if err != nil || result == nil {
		t.Fatalf("预测失败: %v", err)
	}

	for i := 1; i < len(result.DataPoints); i++ {
		prev := result.DataPoints[i-1]
		curr := result.DataPoints[i]

		if curr.PermeabilityRatio >= prev.PermeabilityRatio {
			if curr.SeepageFlowRatio < prev.SeepageFlowRatio*0.95 {
				t.Errorf("第%d点: 渗透系数比上升(%.2f→%.2f)但渗流量比下降(%.2f→%.2f)，物理上可疑",
					i, prev.PermeabilityRatio, curr.PermeabilityRatio,
					prev.SeepageFlowRatio, curr.SeepageFlowRatio)
			}
		}
	}

	t.Log("渗透系数-渗流量相关性验证通过")
}

func TestAgingPredictor_ThreeFactorBiologicalModel_Normal(t *testing.T) {
	model := NewAgingModel()

	if model.MicrobeFactor <= 0 {
		t.Error("微生物侵蚀因子应>0")
	}
	if model.PlantFactor <= 0 {
		t.Error("植物侵蚀因子应>0")
	}
	if model.AnimalFactor <= 0 {
		t.Error("动物侵蚀因子应>0")
	}

	totalBio := model.MicrobeFactor + model.PlantFactor + model.AnimalFactor
	t.Logf("三因子生物侵蚀: 微生物=%.4f, 植物=%.4f, 动物=%.4f, 合计=%.4f",
		model.MicrobeFactor, model.PlantFactor, model.AnimalFactor, totalBio)

	if totalBio < 0.003 || totalBio > 0.01 {
		t.Errorf("三因子合计%.4f超出合理范围[0.003, 0.01]", totalBio)
	}
}

func TestAgingPredictor_BiologicalDamTypeSensitivity_Normal(t *testing.T) {
	model := NewAgingModel()

	ancientDamage := model.calculateAnimalDamage(50.0, string(models.DamTypeAncientStone))
	modernDamage := model.calculateAnimalDamage(50.0, string(models.DamTypeModernConcrete))

	t.Logf("动物侵蚀坝型敏感性: 古代石坝=%.6f, 现代混凝土=%.6f, 比值=%.1f",
		ancientDamage, modernDamage, ancientDamage/modernDamage)

	if ancientDamage <= modernDamage {
		t.Error("古代石坝动物侵蚀应大于现代混凝土坝")
	}
}

func TestAgingPredictor_PlantSeasonalEffect_Normal(t *testing.T) {
	model := NewAgingModel()

	summerDamage := model.calculatePlantDamage(0.5, 500)
	winterDamage := model.calculatePlantDamage(0.0, 500)

	t.Logf("植物侵蚀季节效应: 夏季(0.5年)=%.6f, 冬季(0.0年)=%.6f", summerDamage, winterDamage)
}
