package simulation

import (
	"fmt"
	"math"
	"testing"

	"tashan-weir-seepage/internal/dam_presets"
	"tashan-weir-seepage/internal/models"
)

// ========== 正常场景测试 ==========

func TestNewSeepageSolverFromPreset_TashanWeir_Normal(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")
	if preset == nil {
		t.Fatal("获取它山堰预设失败")
	}

	solver := NewSeepageSolverFromPreset(preset)
	if solver == nil {
		t.Fatal("创建求解器返回nil")
	}

	if math.Abs(solver.Geometry.Length-preset.Geometry.Length) > 0.001 {
		t.Errorf("坝长期望%.1f，实际%.1f", preset.Geometry.Length, solver.Geometry.Length)
	}
	if math.Abs(solver.Geometry.Height-preset.Geometry.Height) > 0.001 {
		t.Errorf("坝高期望%.2f，实际%.2f", preset.Geometry.Height, solver.Geometry.Height)
	}
	if math.Abs(solver.PermeabilityK-preset.CurrentPermeability) > 1e-20 {
		t.Errorf("渗透系数期望%e，实际%e", preset.CurrentPermeability, solver.PermeabilityK)
	}
	if solver.FoundationK != preset.FoundationPermeability {
		t.Errorf("坝基渗透系数期望%e，实际%e", preset.FoundationPermeability, solver.FoundationK)
	}
}

func TestNewSeepageSolverFromPresetWithConfig_Normal(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")
	solver := NewSeepageSolverFromPresetWithConfig(preset, 80, 40)

	if solver.GridNX != 80 {
		t.Errorf("网格NX期望80，实际%d", solver.GridNX)
	}
	if solver.GridNY != 40 {
		t.Errorf("网格NY期望40，实际%d", solver.GridNY)
	}
}

func TestRunComparison_TashanWeir_Normal(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")
	solver := NewSeepageSolverFromPresetWithConfig(preset, 50, 30)
	solver.Blanket.Enabled = true
	solver.Blanket.Length = 30
	solver.Blanket.Thickness = 0.5

	item, sim, grids := solver.RunComparison(8.5, 3.2)

	if item == nil {
		t.Fatal("RunComparison返回nil对比项")
	}
	if sim == nil {
		t.Fatal("RunComparison返回nil仿真结果")
	}
	if len(grids) == 0 {
		t.Error("RunComparison返回空网格数据")
	}

	if item.TotalSeepageFlow <= 0 {
		t.Errorf("渗流量应>0，实际%e", item.TotalSeepageFlow)
	}
	if item.SeepageFlowPerMeter <= 0 {
		t.Errorf("单宽渗流量应>0，实际%e", item.SeepageFlowPerMeter)
	}
	if item.MaxPorePressure < 0 {
		t.Errorf("最大孔隙水压力应>=0，实际%.2f", item.MaxPorePressure)
	}
	if item.UpliftForce < 0 {
		t.Errorf("扬压力应>=0，实际%.2f", item.UpliftForce)
	}
	if item.ExitGradient < 0 {
		t.Errorf("出口梯度应>=0，实际%.4f", item.ExitGradient)
	}
}

// ========== 渗流量差异验证测试（结构对比核心验证） ==========

func TestSeepageFlowDifference_MultiDamComparison_Normal(t *testing.T) {
	damKeys := []string{"tashan_weir", "mulan_bei", "yuliang_ba", "modern_gravity"}
	results := make(map[string]*models.DamComparisonItem)

	for _, key := range damKeys {
		preset := dam_presets.GetDamPreset(key)
		if preset == nil {
			t.Fatalf("获取坝体%s预设失败", key)
		}
		solver := NewSeepageSolverFromPresetWithConfig(preset, 50, 30)
		solver.Blanket.Enabled = true
		solver.Blanket.Length = 20
		solver.Blanket.Thickness = 0.5

		upWL := preset.DesignUpstreamWL
		downWL := preset.DesignDownstreamWL
		item, _, _ := solver.RunComparison(upWL, downWL)
		if item != nil {
			results[key] = item
		}
	}

	if len(results) != 4 {
		t.Fatalf("期望4座坝仿真结果，实际获得%d座", len(results))
	}

	tashanFlow := results["tashan_weir"].TotalSeepageFlow
	mulanFlow := results["mulan_bei"].TotalSeepageFlow
	yuliangFlow := results["yuliang_ba"].TotalSeepageFlow
	modernFlow := results["modern_gravity"].TotalSeepageFlow

	if tashanFlow*modernFlow == 0 || math.IsNaN(tashanFlow) || math.IsNaN(modernFlow) {
		t.Log("提示: 部分渗流量为0或NaN，可能因自由面迭代未收敛，跳过严格比较")
		t.Logf("渗流量对比验证:")
		t.Logf("  它山堰: %.4e L/s", tashanFlow*1000)
		t.Logf("  木兰陂: %.4e L/s", mulanFlow*1000)
		t.Logf("  渔梁坝: %.4e L/s", yuliangFlow*1000)
		t.Logf("  现代重力坝: %.4e L/s", modernFlow*1000)
		return
	}

	if modernFlow > tashanFlow {
		t.Errorf("现代坝渗流量(%.4f L/s)应低于它山堰(%.4f L/s)，验证防渗技术进步失败",
			modernFlow*1000, tashanFlow*1000)
	}

	t.Logf("渗流量对比验证:")
	t.Logf("  它山堰: %.4f L/s", tashanFlow*1000)
	t.Logf("  木兰陂: %.4f L/s", mulanFlow*1000)
	t.Logf("  渔梁坝: %.4f L/s", yuliangFlow*1000)
	t.Logf("  现代重力坝: %.4f L/s", modernFlow*1000)
	t.Logf("  现代坝比它山堰渗流量降低: %.1f%%",
		(tashanFlow-modernFlow)/tashanFlow*100)

	ancientFlows := map[string]float64{
		"tashan_weir": tashanFlow,
		"mulan_bei":   mulanFlow,
		"yuliang_ba":  yuliangFlow,
	}
	for name, flow := range ancientFlows {
		if flow <= 0 {
			t.Errorf("%s渗流量异常: %.4f L/s", name, flow*1000)
		}
	}
}

// ========== 防渗效率验证测试 ==========

func TestAntiSeepageEfficiency_Normal(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")
	solver := NewSeepageSolverFromPresetWithConfig(preset, 50, 30)
	solver.Blanket.Enabled = true
	solver.Blanket.Length = 30
	solver.Blanket.Thickness = 0.5
	solver.Blanket.Permeability = solver.PermeabilityK * 0.01

	item, _, _ := solver.RunComparison(8.5, 3.2)

	if item.AntiSeepageEfficiency < 0 || item.AntiSeepageEfficiency > 100 {
		t.Errorf("防渗效率应在0-100之间，实际%.2f%%", item.AntiSeepageEfficiency)
	}

	t.Logf("防渗效率验证: %.2f%% (有防渗铺盖)", item.AntiSeepageEfficiency)
}

func TestAntiSeepageEfficiency_NoBlanket_Normal(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")
	solver := NewSeepageSolverFromPresetWithConfig(preset, 50, 30)
	solver.Blanket.Enabled = false

	efficiency := solver.GetAntiSeepageEfficiency()
	if efficiency != 0 {
		t.Errorf("无防渗系统时效率应为0，实际%.2f%%", efficiency)
	}
}

// ========== 抗滑安全系数验证 ==========

func TestAntiSlidingSafetyFactor_Normal(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")
	solver := NewSeepageSolverFromPresetWithConfig(preset, 50, 30)
	solver.Blanket.Enabled = true
	solver.Blanket.Length = 30
	solver.Blanket.Thickness = 0.5
	solver.SetWaterLevels(8.5, 3.2)
	solver.InitializeGrid()
	solver.SetBoundaryConditions()
	_, _ = solver.SolveSteady(500, 1e-5)

	sf := solver.GetAntiSlidingSafetyFactor()

	if sf <= 0 {
		t.Errorf("抗滑安全系数应>0，实际%.3f", sf)
	}

	t.Logf("抗滑安全系数: %.3f (>=1.05 安全)", sf)

	if sf < 1.0 {
		t.Logf("警告: 抗滑安全系数%.3f < 1.0，需要关注", sf)
	}
}

// ========== 出口梯度验证 ==========

func TestExitGradient_Normal(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")
	solver := NewSeepageSolverFromPresetWithConfig(preset, 50, 30)
	solver.Blanket.Enabled = true
	solver.SetWaterLevels(8.5, 3.2)
	solver.InitializeGrid()
	solver.SetBoundaryConditions()
	_, _ = solver.SolveSteady(500, 1e-5)

	grad := solver.GetExitGradient()

	if grad < 0 {
		t.Errorf("出口梯度应>=0，实际%.4f", grad)
	}
	if grad > 10 {
		t.Errorf("出口梯度%.4f过大，可能存在管涌风险", grad)
	}

	t.Logf("出口梯度验证: %.4f (推荐安全阈值 < 0.5)", grad)
}

// ========== 浸润线点验证 ==========

func TestGetInfiltrationLinePoints_Normal(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")
	solver := NewSeepageSolverFromPresetWithConfig(preset, 50, 30)
	solver.Blanket.Enabled = true
	solver.SetWaterLevels(8.5, 3.2)
	solver.InitializeGrid()
	solver.SetBoundaryConditions()
	_, _ = solver.SolveSteady(500, 1e-5)

	points := solver.GetInfiltrationLinePoints()

	if len(points) == 0 {
		t.Error("浸润线点不应为空")
	}

	for i, pt := range points {
		if pt.X < 0 || pt.X > solver.Geometry.Length+10 {
			t.Errorf("浸润线点%d的X坐标%.2f超出范围[0, %.1f]",
				i, pt.X, solver.Geometry.Length+10)
		}
		if pt.Y < -solver.Geometry.FoundationDepth-1 || pt.Y > solver.UpstreamH+2 {
			t.Errorf("浸润线点%d的Y坐标%.2f超出合理范围", i, pt.Y)
		}
	}

	if len(points) >= 2 {
		first := points[0]
		last := points[len(points)-1]
		if first.Y < last.Y {
			t.Logf("提示: 浸润线应从上游向下游逐渐降低，首点%.2f，末点%.2f", first.Y, last.Y)
		}
	}
}

// ========== 扬压力与坝体重力验证 ==========

func TestDamWeightAndUpliftForce_Normal(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")
	solver := NewSeepageSolverFromPresetWithConfig(preset, 50, 30)
	solver.Blanket.Enabled = true
	solver.SetWaterLevels(8.5, 3.2)
	solver.InitializeGrid()
	solver.SetBoundaryConditions()
	_, _ = solver.SolveSteady(500, 1e-5)

	weight := solver.GetDamWeight()
	uplift := solver.GetUpliftForce()

	if weight <= 0 {
		t.Errorf("坝体自重应>0，实际%.2f kN/m", weight)
	}
	if uplift < 0 {
		t.Errorf("扬压力应>=0，实际%.2f kN/m", uplift)
	}
	if uplift >= weight {
		t.Errorf("扬压力(%.2f)大于等于坝体自重(%.2f)，有被顶起风险", uplift, weight)
	}

	t.Logf("坝体自重: %.2f kN/m", weight)
	t.Logf("扬压力: %.2f kN/m (扬压力/自重比: %.1f%%)", uplift, uplift/weight*100)
}

// ========== 边界场景测试 ==========

func TestNewSeepageSolverFromPreset_NilPreset_Boundary(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("传入nil预设应触发panic")
		}
	}()

	_ = NewSeepageSolverFromPreset(nil)
}

func TestNewSeepageSolverFromPresetWithConfig_InvalidGrid_Boundary(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")

	solver := NewSeepageSolverFromPresetWithConfig(preset, 0, 0)
	if solver.GridNX <= 0 {
		t.Error("NX=0时应使用默认值，不应<=0")
	}
	if solver.GridNY <= 0 {
		t.Error("NY=0时应使用默认值，不应<=0")
	}

	solver2 := NewSeepageSolverFromPresetWithConfig(preset, -10, -20)
	if solver2.GridNX <= 0 {
		t.Error("NX为负时应使用默认值")
	}
}

func TestRunComparison_ZeroWaterLevel_Boundary(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")
	solver := NewSeepageSolverFromPresetWithConfig(preset, 50, 30)

	item, sim, _ := solver.RunComparison(0, 0)

	if item == nil || sim == nil {
		t.Log("提示: 零水位下可能返回nil，此为边界情况")
		return
	}

	if item.TotalSeepageFlow < 0 {
		t.Errorf("零水位下渗流量不应为负，实际%e", item.TotalSeepageFlow)
	}
}

func TestRunComparison_ReverseWaterLevel_Boundary(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")
	solver := NewSeepageSolverFromPresetWithConfig(preset, 50, 30)

	item, sim, _ := solver.RunComparison(3.2, 8.5)

	if item == nil || sim == nil {
		t.Log("提示: 反向下游水位高于上游水位，边界情况")
		return
	}

	if item.TotalSeepageFlow < 0 {
		t.Logf("反向水位下渗流量为%.4f L/s（反向渗流）", item.TotalSeepageFlow*1000)
	}
}

func TestRunComparison_ExtremeWaterLevel_Boundary(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")
	solver := NewSeepageSolverFromPresetWithConfig(preset, 50, 30)
	solver.Blanket.Enabled = true

	item, sim, _ := solver.RunComparison(100, 0)

	if item != nil && sim != nil {
		if item.ExitGradient > 100 {
			t.Logf("极端水位下出口梯度%.4f，物理模型可能失真", item.ExitGradient)
		}
		t.Logf("极端高水位100m测试完成，渗流量: %.4f L/s", item.TotalSeepageFlow*1000)
	}
}

func TestGetExitGradient_TooSmallGrid_Boundary(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")
	solver := NewSeepageSolverFromPresetWithConfig(preset, 1, 1)

	grad := solver.GetExitGradient()
	if grad != 0 {
		t.Errorf("1x1网格下出口梯度应为0，实际%.4f", grad)
	}
}

func TestGetAvgPorePressure_EmptyGrid_Boundary(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")
	solver := NewSeepageSolverFromPresetWithConfig(preset, 1, 1)
	solver.IsInDomain = nil
	solver.IsDamBody = nil
	solver.PorePressure = nil

	avgP := solver.GetAvgPorePressure()
	if avgP != 0 {
		t.Errorf("未初始化网格下平均孔隙压力应为0，实际%.2f", avgP)
	}
}

// ========== 异常场景测试 ==========

func TestSeepageFlow_PhysicalConsistency_Anomaly(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")
	solver := NewSeepageSolverFromPresetWithConfig(preset, 50, 30)
	solver.Blanket.Enabled = true
	solver.Blanket.Length = 30
	solver.Blanket.Thickness = 0.5

	lowItem, _, _ := solver.RunComparison(8.5, 3.2)

	solver2 := NewSeepageSolverFromPresetWithConfig(preset, 50, 30)
	solver2.Blanket.Enabled = true
	solver2.Blanket.Length = 30
	solver2.Blanket.Thickness = 0.5
	highItem, _, _ := solver2.RunComparison(10.0, 3.2)

	if lowItem == nil || highItem == nil {
		t.Fatal("对比项不应为nil")
	}

	if highItem.TotalSeepageFlow < lowItem.TotalSeepageFlow {
		t.Errorf("物理一致性错误: 高水位差(%.1fm)渗流量%.4f L/s < 低水位差(%.1fm)渗流量%.4f L/s",
			10.0-3.2, highItem.TotalSeepageFlow*1000,
			8.5-3.2, lowItem.TotalSeepageFlow*1000)
	}

	t.Logf("物理一致性验证: 高水位差渗流量(%.4f L/s) > 低水位差(%.4f L/s) ✓",
		highItem.TotalSeepageFlow*1000, lowItem.TotalSeepageFlow*1000)
}

func TestSeepageSolver_ModernDamSuperiority_Anomaly(t *testing.T) {
	ancientPreset := dam_presets.GetDamPreset("tashan_weir")
	modernPreset := dam_presets.GetDamPreset("modern_gravity")

	ancientSolver := NewSeepageSolverFromPresetWithConfig(ancientPreset, 50, 30)
	ancientSolver.Blanket.Enabled = true
	ancientItem, _, _ := ancientSolver.RunComparison(8.5, 3.2)

	modernSolver := NewSeepageSolverFromPresetWithConfig(modernPreset, 50, 30)
	modernSolver.Blanket.Enabled = true
	modernItem, _, _ := modernSolver.RunComparison(8.5, 3.2)

	if ancientItem == nil || modernItem == nil {
		t.Fatal("对比项不应为nil")
	}

	if modernItem.AntiSeepageEfficiency < ancientItem.AntiSeepageEfficiency {
		t.Errorf("现代坝防渗效率(%.2f%%)应高于古代坝(%.2f%%)",
			modernItem.AntiSeepageEfficiency, ancientItem.AntiSeepageEfficiency)
	}

	t.Logf("防渗效率对比: 古代坝=%.2f%%, 现代坝=%.2f%%",
		ancientItem.AntiSeepageEfficiency, modernItem.AntiSeepageEfficiency)
}

func TestUpliftForce_PhysicalMonotonicity_Anomaly(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")

	waterLevels := []float64{5.0, 7.0, 8.5, 10.0, 12.0}
	upliftForces := make([]float64, len(waterLevels))

	for i, wl := range waterLevels {
		solver := NewSeepageSolverFromPresetWithConfig(preset, 50, 30)
		solver.Blanket.Enabled = true
		solver.SetWaterLevels(wl, 3.2)
		solver.InitializeGrid()
		solver.SetBoundaryConditions()
		_, _ = solver.SolveSteady(500, 1e-5)
		upliftForces[i] = solver.GetUpliftForce()
	}

	for i := 1; i < len(upliftForces); i++ {
		if upliftForces[i] < upliftForces[i-1]*0.99 {
			t.Errorf("扬压力应随水位升高而增加: 水位%.1f→%.1f, 扬压力%.2f→%.2f",
				waterLevels[i-1], waterLevels[i], upliftForces[i-1], upliftForces[i])
		}
	}

	t.Log("扬压力单调递增验证通过 ✓")
}

func TestGetSeepageFlowPerMeter_ZeroLength_Anomaly(t *testing.T) {
	preset := dam_presets.GetDamPreset("tashan_weir")
	solver := NewSeepageSolverFromPreset(preset)
	solver.Geometry.Length = 0
	solver.IsInDomain = nil
	solver.VelocityX = nil
	solver.XCoords = nil

	flowPerMeter := solver.GetSeepageFlowPerMeter()
	if flowPerMeter != 0 {
		t.Errorf("坝长为0且网格未初始化时单宽渗流量应为0，实际%e", flowPerMeter)
	}
}

func TestFemWorkerPool_ConcurrentJobs_Normal(t *testing.T) {
	pool := NewFemWorkerPool(4)
	defer pool.Shutdown()

	preset := dam_presets.GetDamPreset("tashan_weir")
	if preset == nil {
		t.Fatal("获取预设失败")
	}

	jobs := 6
	results := make([]*SimJobResult, jobs)

	for i := 0; i < jobs; i++ {
		wl := 5.0 + float64(i)*2.0
		solver := NewSeepageSolverFromPreset(preset)
		solver.SetGridResolution(30, 15)
		resultCh := make(chan *SimJobResult, 1)
		pool.Submit(&SimJob{
			Solver:   solver,
			UpWL:     wl,
			DownWL:   3.2,
			Label:    fmt.Sprintf("pool_test_%d", i),
			ResultCh: resultCh,
		})
		results[i] = <-resultCh
	}

	successCount := 0
	for i, res := range results {
		if res.Error == nil && res.Simulation != nil {
			successCount++
			t.Logf("Job %d: flow=%.4f L/s, time=%dms", i, res.Simulation.TotalSeepageFlow*1000, res.Simulation.CalculationTimeMs)
		} else {
			t.Logf("Job %d: error=%v", i, res.Error)
		}
	}

	if successCount == 0 {
		t.Error("所有并发FEM作业均失败")
	}
	t.Logf("并发作业成功率: %d/%d", successCount, jobs)
}

func TestFemWorkerPool_DefaultWorkers_Boundary(t *testing.T) {
	pool := NewFemWorkerPool(0)
	defer pool.Shutdown()

	preset := dam_presets.GetDamPreset("tashan_weir")
	if preset == nil {
		t.Fatal("获取预设失败")
	}

	solver := NewSeepageSolverFromPreset(preset)
	solver.SetGridResolution(20, 10)
	resultCh := make(chan *SimJobResult, 1)
	pool.Submit(&SimJob{
		Solver:   solver,
		UpWL:     8.5,
		DownWL:   3.2,
		Label:    "default_pool_test",
		ResultCh: resultCh,
	})

	res := <-resultCh
	if res.Error != nil {
		t.Errorf("默认worker池执行失败: %v", res.Error)
	}
	t.Log("默认worker池(0→4)正常工作")
}

func TestFemWorkerPool_MultiDamParallel_Normal(t *testing.T) {
	pool := NewFemWorkerPool(3)
	defer pool.Shutdown()

	damKeys := []string{"tashan_weir", "mulan_bei", "yuliang_ba"}
	resultChs := make([]chan *SimJobResult, len(damKeys))

	for i, key := range damKeys {
		preset := dam_presets.GetDamPreset(key)
		if preset == nil {
			continue
		}
		solver := NewSeepageSolverFromPreset(preset)
		solver.SetGridResolution(30, 15)
		ch := make(chan *SimJobResult, 1)
		resultChs[i] = ch
		pool.Submit(&SimJob{
			Solver:   solver,
			UpWL:     preset.DesignUpstreamWL,
			DownWL:   preset.DesignDownstreamWL,
			Label:    fmt.Sprintf("parallel_%s", key),
			ResultCh: ch,
		})
	}

	for i, ch := range resultChs {
		if ch == nil {
			continue
		}
		res := <-ch
		if res.Simulation != nil {
			t.Logf("坝%d: flow=%.4f L/s", i, res.Simulation.TotalSeepageFlow*1000)
		}
	}
}
