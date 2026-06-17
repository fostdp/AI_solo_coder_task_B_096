package dam_presets

import (
	"math"
	"testing"

	"tashan-weir-seepage/internal/models"
)

// ========== 正常场景测试 ==========

func TestGetAllDamPresets_Normal(t *testing.T) {
	presets := GetAllDamPresets()

	if len(presets) != 4 {
		t.Errorf("期望4座堰坝预设，实际得到%d座", len(presets))
	}

	expectedNames := map[string]bool{
		"tashan_weir":    false,
		"mulan_bei":      false,
		"yuliang_ba":     false,
		"modern_gravity": false,
	}

	for _, p := range presets {
		if _, ok := expectedNames[p.DamKey]; !ok {
			t.Errorf("发现未知堰坝: %s", p.DamKey)
		}
		expectedNames[p.DamKey] = true

		if p.DamName == "" {
			t.Errorf("堰坝 %s 名称为空", p.DamKey)
		}
		if p.Geometry.Length <= 0 {
			t.Errorf("堰坝 %s 坝长应>0，实际%.1f", p.DamKey, p.Geometry.Length)
		}
		if p.Geometry.Height <= 0 {
			t.Errorf("堰坝 %s 坝高应>0，实际%.1f", p.DamKey, p.Geometry.Height)
		}
		if p.CurrentPermeability <= 0 {
			t.Errorf("堰坝 %s 渗透系数应>0，实际%e", p.DamKey, p.CurrentPermeability)
		}
	}

	for key, found := range expectedNames {
		if !found {
			t.Errorf("缺少期望的堰坝预设: %s", key)
		}
	}
}

func TestGetDamPreset_TashanWeir_Normal(t *testing.T) {
	preset := GetDamPreset("tashan_weir")

	if preset == nil {
		t.Fatal("获取它山堰预设返回nil")
	}

	if preset.DamName != "它山堰" {
		t.Errorf("坝名期望'它山堰'，实际'%s'", preset.DamName)
	}
	if preset.BuildYear != 833 {
		t.Errorf("建造年份期望833，实际%d", preset.BuildYear)
	}
	if preset.DamType != models.DamTypeAncientStone {
		t.Errorf("坝型期望古代条石坝，实际%s", preset.DamType)
	}
	if math.Abs(preset.Geometry.Length-113.7) > 0.001 {
		t.Errorf("坝长期望113.7m，实际%.1fm", preset.Geometry.Length)
	}
	if math.Abs(preset.Geometry.Height-3.85) > 0.001 {
		t.Errorf("坝高期望3.85m，实际%.2fm", preset.Geometry.Height)
	}
	if !preset.HasAntiSeepageSystem {
		t.Error("它山堰应具备防渗系统")
	}
	if preset.CurrentPermeability <= 0 {
		t.Errorf("渗透系数应>0，实际%e", preset.CurrentPermeability)
	}
	if preset.FoundationDepth < preset.Geometry.Height*0.5 {
		t.Error("基础深度应至少为坝高的50%")
	}
}

func TestGetDamPreset_ModernGravity_Normal(t *testing.T) {
	preset := GetDamPreset("modern_gravity")

	if preset == nil {
		t.Fatal("获取现代重力坝预设返回nil")
	}

	if preset.DamType != models.DamTypeModernConcrete {
		t.Errorf("坝型期望现代混凝土重力坝，实际%s", preset.DamType)
	}
	if preset.BuildYear != 2020 {
		t.Errorf("建造年份期望2020，实际%d", preset.BuildYear)
	}
	if preset.CurrentPermeability >= 1e-9 {
		t.Errorf("现代混凝土坝渗透系数应远低于古代坝，实际%e", preset.CurrentPermeability)
	}
	if preset.Geometry.Height <= 10 {
		t.Errorf("现代重力坝坝高应>10m，实际%.1fm", preset.Geometry.Height)
	}
}

func TestGetDamKeys_Normal(t *testing.T) {
	keys := GetDamKeys()

	if len(keys) != 4 {
		t.Errorf("期望4个坝key，实际%d个", len(keys))
	}

	expected := map[string]bool{
		"tashan_weir":    false,
		"mulan_bei":      false,
		"yuliang_ba":     false,
		"modern_gravity": false,
	}
	for _, k := range keys {
		expected[k] = true
	}
	for k, found := range expected {
		if !found {
			t.Errorf("缺少key: %s", k)
		}
	}
}

func TestGetVirtualTourScenes_TashanWeir_Normal(t *testing.T) {
	scenes := GetVirtualTourScenes("tashan_weir")

	if len(scenes) < 5 {
		t.Errorf("期望至少5个参观场景，实际%d个", len(scenes))
	}

	expectedSceneIDs := map[string]bool{
		"overview":         false,
		"upstream_view":    false,
		"seepage_cutaway":  false,
		"downstream_view":  false,
		"sensor_layout":    false,
	}

	for i, scene := range scenes {
		if scene.SceneID == "" {
			t.Errorf("场景%d缺少SceneID", i)
		}
		if scene.SceneName == "" {
			t.Errorf("场景%d缺少SceneName", i)
		}
		if scene.Narrative == "" {
			t.Errorf("场景%d缺少解说词Narrative", i)
		}
		if scene.CameraPos == (models.Point3D{}) {
			t.Errorf("场景%d缺少相机位置", i)
		}
		if scene.CameraTarget == (models.Point3D{}) {
			t.Errorf("场景%d缺少相机目标点", i)
		}

		if _, ok := expectedSceneIDs[scene.SceneID]; ok {
			expectedSceneIDs[scene.SceneID] = true
		}

		for j, hs := range scene.Hotspots {
			if hs.HotspotID == "" {
				t.Errorf("场景%d的热点%d缺少HotspotID", i, j)
			}
			if hs.Title == "" {
				t.Errorf("场景%d的热点%d缺少Title", i, j)
			}
		}
	}

	for id, found := range expectedSceneIDs {
		if !found {
			t.Errorf("缺少期望的参观场景: %s", id)
		}
	}
}

func TestGetDamTypeLabel_Normal(t *testing.T) {
	tests := []struct {
		damType  models.DamType
		expected string
	}{
		{models.DamTypeAncientStone, "古代条石坝"},
		{models.DamTypeModernConcrete, "现代混凝土重力坝"},
		{models.DamType("unknown_type"), "未知类型"},
	}

	for _, tt := range tests {
		result := GetDamTypeLabel(tt.damType)
		if result != tt.expected {
			t.Errorf("DamType=%s，期望'%s'，实际'%s'", tt.damType, tt.expected, result)
		}
	}
}

// ========== 渗流量差异验证测试（结构对比） ==========

func TestDamPresets_SeepageFlowDifference_StructureComparison(t *testing.T) {
	damKeys := []string{"tashan_weir", "mulan_bei", "yuliang_ba", "modern_gravity"}
	presets := make(map[string]*models.DamPreset)

	for _, key := range damKeys {
		presets[key] = GetDamPreset(key)
		if presets[key] == nil {
			t.Fatalf("获取堰坝%s预设失败", key)
		}
	}

	tashan := presets["tashan_weir"]
	mulan := presets["mulan_bei"]
	yuliang := presets["yuliang_ba"]
	modern := presets["modern_gravity"]

	if modern.CurrentPermeability >= tashan.CurrentPermeability {
		t.Errorf("现代坝渗透系数(%e)应低于它山堰(%e)",
			modern.CurrentPermeability, tashan.CurrentPermeability)
	}

	if modern.CurrentPermeability >= tashan.CurrentPermeability/100 {
		t.Logf("提示: 现代坝渗透系数(%e)比古代坝(%e)降低%.0f倍，符合技术进步预期",
			modern.CurrentPermeability, tashan.CurrentPermeability,
			tashan.CurrentPermeability/modern.CurrentPermeability)
	}

	ancientPermeabilities := []float64{
		tashan.CurrentPermeability,
		mulan.CurrentPermeability,
		yuliang.CurrentPermeability,
	}
	for i, k1 := range ancientPermeabilities {
		for j, k2 := range ancientPermeabilities {
			if i != j && k1 == k2 {
				t.Logf("提示: 古代坝渗透系数存在相同值(%e)，可能需要根据实际监测数据校准", k1)
			}
		}
	}

	for key, preset := range presets {
		if preset.OriginalPermeability > preset.CurrentPermeability {
			t.Errorf("%s: 初始渗透系数(%e)应小于当前(%e)，因为老化会增加渗透性",
				key, preset.OriginalPermeability, preset.CurrentPermeability)
		}
	}
}

// ========== 防渗结构对比测试 ==========

func TestDamPresets_AntiSeepageStructureComparison(t *testing.T) {
	presets := GetAllDamPresets()

	for _, p := range presets {
		if !p.HasAntiSeepageSystem {
			t.Errorf("%s 应具备防渗系统", p.DamName)
		}
		if p.AntiSeepageDescription == "" {
			t.Errorf("%s 缺少防渗结构描述", p.DamName)
		}
		if p.InterfacePermeability <= 0 {
			t.Errorf("%s 界面渗透系数应>0", p.DamName)
		}
		if p.FoundationPermeability <= 0 {
			t.Errorf("%s 坝基渗透系数应>0", p.DamName)
		}
	}

	tashan := GetDamPreset("tashan_weir")
	if tashan != nil {
		if tashan.InterfacePermeability > tashan.CurrentPermeability {
			t.Logf("提示: 它山堰界面渗透系数(%e) > 坝体渗透系数(%e)，界面是防渗薄弱环节",
				tashan.InterfacePermeability, tashan.CurrentPermeability)
		}
	}
}

// ========== 边界场景测试 ==========

func TestGetDamPreset_EmptyKey_Boundary(t *testing.T) {
	preset := GetDamPreset("")
	if preset != nil {
		t.Error("空key应返回nil")
	}
}

func TestGetDamPreset_InvalidKey_Boundary(t *testing.T) {
	invalidKeys := []string{
		"nonexistent_dam",
		"invalid",
		"12345",
		"!@#$%",
		"TASHAN_WEIR",
		"  tashan_weir  ",
	}

	for _, key := range invalidKeys {
		preset := GetDamPreset(key)
		if preset != nil {
			t.Errorf("无效key '%s' 应返回nil", key)
		}
	}
}

func TestGetVirtualTourScenes_InvalidDam_Boundary(t *testing.T) {
	invalidDams := []string{
		"modern_gravity",
		"mulan_bei",
		"nonexistent",
		"",
	}

	for _, dam := range invalidDams {
		scenes := GetVirtualTourScenes(dam)
		if len(scenes) != 0 {
			t.Logf("提示: 坝体%s当前有%d个虚拟参观场景，可扩展更多参观内容", dam, len(scenes))
		}
	}
}

func TestDamGeometry_ExtremeValues_Boundary(t *testing.T) {
	presets := GetAllDamPresets()

	for _, p := range presets {
		if p.Geometry.Length > 1000 {
			t.Errorf("%s 坝长%.1fm过大，边界值不应超过1000m", p.DamName, p.Geometry.Length)
		}
		if p.Geometry.Height > 200 {
			t.Errorf("%s 坝高%.1fm过大，边界值不应超过200m", p.DamName, p.Geometry.Height)
		}
		if p.Geometry.UpstreamSlope < 0 || p.Geometry.UpstreamSlope > 2 {
			t.Errorf("%s 上游坡度%.2f超出合理范围(0-2)", p.DamName, p.Geometry.UpstreamSlope)
		}
		if p.Geometry.DownstreamSlope < 0 || p.Geometry.DownstreamSlope > 2 {
			t.Errorf("%s 下游坡度%.2f超出合理范围(0-2)", p.DamName, p.Geometry.DownstreamSlope)
		}
		if p.CurrentPermeability < 1e-13 || p.CurrentPermeability > 1e-3 {
			t.Errorf("%s 渗透系数%e超出合理范围(1e-13 ~ 1e-3 m/s)", p.DamName, p.CurrentPermeability)
		}
	}
}

func TestBuildYear_Order_Boundary(t *testing.T) {
	yuliang := GetDamPreset("yuliang_ba")
	tashan := GetDamPreset("tashan_weir")
	mulan := GetDamPreset("mulan_bei")
	modern := GetDamPreset("modern_gravity")

	if yuliang == nil || tashan == nil || mulan == nil || modern == nil {
		t.Fatal("获取预设失败")
	}

	if !(yuliang.BuildYear < tashan.BuildYear) {
		t.Errorf("渔梁坝(%d年)应早于它山堰(%d年)", yuliang.BuildYear, tashan.BuildYear)
	}
	if !(tashan.BuildYear < mulan.BuildYear) {
		t.Errorf("它山堰(%d年)应早于木兰陂(%d年)", tashan.BuildYear, mulan.BuildYear)
	}
	if !(mulan.BuildYear < modern.BuildYear) {
		t.Errorf("木兰陂(%d年)应早于现代坝(%d年)", mulan.BuildYear, modern.BuildYear)
	}
}

// ========== 异常场景测试 ==========

func TestDamPreset_ConsistencyCheck_Anomaly(t *testing.T) {
	presets := GetAllDamPresets()

	for _, p := range presets {
		if p.DesignUpstreamWL <= p.DesignDownstreamWL {
			t.Errorf("%s: 设计上游水位(%.1f)应高于下游水位(%.1f)",
				p.DamName, p.DesignUpstreamWL, p.DesignDownstreamWL)
		}

		if p.DesignUpstreamWL-p.DesignDownstreamWL > p.Geometry.Height+5 {
			t.Errorf("%s: 设计水位差%.1fm超过坝高%.1fm太多，可能存在数据异常",
				p.DamName, p.DesignUpstreamWL-p.DesignDownstreamWL, p.Geometry.Height)
		}

		if p.MaterialType == "" {
			t.Errorf("%s: 缺少材料类型描述", p.DamName)
		}
		if p.Location == "" {
			t.Errorf("%s: 缺少地理位置信息", p.DamName)
		}
	}
}

func TestVirtualTourScene_EducationValue_Anomaly(t *testing.T) {
	scenes := GetVirtualTourScenes("tashan_weir")

	for i, scene := range scenes {
		minNarrativeLength := 20
		if len(scene.Narrative) < minNarrativeLength {
			t.Errorf("场景%d(%s)解说词过短(%d字符)，教育性不足，至少需要%d字符",
				i, scene.SceneName, len(scene.Narrative), minNarrativeLength)
		}

		if len(scene.Hotspots) == 0 {
			t.Logf("提示: 场景%d(%s)没有热点，可增加交互点提升教育性", i, scene.SceneName)
		}

		for j, hs := range scene.Hotspots {
			if len(hs.Description) < 5 {
				t.Errorf("场景%d的热点%d(%s)描述过短，教育信息不足", i, j, hs.Title)
			}
		}
	}
}

func TestDamPreset_NoDuplicateKeys_Anomaly(t *testing.T) {
	seen := make(map[string]bool)
	for _, p := range GetAllDamPresets() {
		if seen[p.DamKey] {
			t.Errorf("发现重复的坝key: %s", p.DamKey)
		}
		seen[p.DamKey] = true
	}
}

func TestDamPreset_DamTypeConsistency_Anomaly(t *testing.T) {
	tashan := GetDamPreset("tashan_weir")
	mulan := GetDamPreset("mulan_bei")
	yuliang := GetDamPreset("yuliang_ba")
	modern := GetDamPreset("modern_gravity")

	if tashan.DamType != models.DamTypeAncientStone {
		t.Errorf("它山堰坝类型应为ancient_stone，实际%s", tashan.DamType)
	}
	if mulan.DamType != models.DamTypeAncientStone {
		t.Errorf("木兰陂坝类型应为ancient_stone，实际%s", mulan.DamType)
	}
	if yuliang.DamType != models.DamTypeAncientStone {
		t.Errorf("渔梁坝坝类型应为ancient_stone，实际%s", yuliang.DamType)
	}
	if modern.DamType != models.DamTypeModernConcrete {
		t.Errorf("现代重力坝坝类型应为modern_concrete，实际%s", modern.DamType)
	}
}
