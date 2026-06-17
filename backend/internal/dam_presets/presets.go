package dam_presets

import (
	"tashan-weir-seepage/internal/models"
)

var damPresets = map[string]*models.DamPreset{
	"tashan_weir": {
		DamKey:                 "tashan_weir",
		DamName:                "它山堰",
		DamType:                models.DamTypeAncientStone,
		BuildDynasty:           "唐·大和七年",
		BuildYear:              833,
		Location:               "浙江省宁波市鄞州区",
		Description:            "中国古代四大水利工程之一，由县令王元暐主持修建，是一座以灌溉为主，结合防洪、蓄淡、引水等综合利用的大型水利工程。",
		HistoricalSignificance: "它山堰是中国古代水利工程的杰出代表，与都江堰、郑国渠、灵渠并称为中国古代四大水利工程，历经千余年仍发挥作用。",
		Geometry: models.DamGeometry{
			Length:          113.7,
			Height:          3.85,
			TopWidth:        4.8,
			UpstreamSlope:   0.35,
			DownstreamSlope: 0.6,
		},
		FoundationDepth:        5.0,
		DesignUpstreamWL:       8.5,
		DesignDownstreamWL:     3.2,
		MaterialType:           "条石砌筑+黏土心墙",
		OriginalPermeability:   1.2e-8,
		CurrentPermeability:    1.0e-7,
		FoundationPermeability: 5.0e-7,
		InterfacePermeability:  5.0e-8,
		HasAntiSeepageSystem:   true,
		AntiSeepageDescription: "采用黏土心墙防渗，上下游设护坦消能，坝基设防渗帷幕。堰体为重力式结构，利用条石自重抗滑稳定。",
		CulturalValue:          "1988年被列为全国重点文物保护单位，2015年入选世界灌溉工程遗产名录。",
	},
	"mulan_bei": {
		DamKey:                 "mulan_bei",
		DamName:                "木兰陂",
		DamType:                models.DamTypeAncientStone,
		BuildDynasty:           "北宋·治平元年",
		BuildYear:              1064,
		Location:               "福建省莆田市",
		Description:            "木兰陂是中国古代大型水利工程，位于木兰溪下游，是一座引、蓄、灌、排、挡综合利用的大型水利工程，灌溉面积达20余万亩。",
		HistoricalSignificance: "木兰陂是中国现存最完整的古代水利工程之一，历经三次修建方成，是福建水利史上的里程碑，与都江堰、灵渠、它山堰并称中国古代四大水利工程。",
		Geometry: models.DamGeometry{
			Length:          219.0,
			Height:          7.5,
			TopWidth:        5.0,
			UpstreamSlope:   0.4,
			DownstreamSlope: 0.7,
		},
		FoundationDepth:        8.0,
		DesignUpstreamWL:       12.5,
		DesignDownstreamWL:     5.0,
		MaterialType:           "花岗岩条石砌筑+糯米灰浆勾缝",
		OriginalPermeability:   8.0e-9,
		CurrentPermeability:    1.5e-7,
		FoundationPermeability: 2.0e-6,
		InterfacePermeability:  4.0e-8,
		HasAntiSeepageSystem:   true,
		AntiSeepageDescription: "采用糯米灰浆勾缝防渗，坝基采用木桩加密加固，上下游设闸控制水位。堰体分堰顶、堰闸、堰墩三部分，结构精巧。",
		CulturalValue:          "2014年被列为世界灌溉工程遗产，是研究中国古代水利工程技术的重要实物资料。",
	},
	"yuliang_ba": {
		DamKey:                 "yuliang_ba",
		DamName:                "渔梁坝",
		DamType:                models.DamTypeAncientStone,
		BuildDynasty:           "唐·开元年间",
		BuildYear:              730,
		Location:               "安徽省黄山市歙县",
		Description:            "渔梁坝是新安江上游最古老、规模最大的古代拦河坝，是徽州古代最知名的水利工程，被称为'江南第一都江堰'。",
		HistoricalSignificance: "渔梁坝始建于唐代，明代重建，是徽州商人从家乡走向全国的起点，见证了徽商的兴盛与徽州水利文明的辉煌。",
		Geometry: models.DamGeometry{
			Length:          138.0,
			Height:          5.5,
			TopWidth:        6.0,
			UpstreamSlope:   0.45,
			DownstreamSlope: 0.65,
		},
		FoundationDepth:        6.5,
		DesignUpstreamWL:       10.0,
		DesignDownstreamWL:     4.5,
		MaterialType:           "花岗岩条石+石榫连接+桐油灰浆",
		OriginalPermeability:   9.0e-9,
		CurrentPermeability:    8.0e-8,
		FoundationPermeability: 8.0e-7,
		InterfacePermeability:  4.5e-8,
		HasAntiSeepageSystem:   true,
		AntiSeepageDescription: "条石之间用石榫锁合，缝隙用桐油拌石灰、糯米浆填筑，防渗效果极佳。坝底设泄水孔，兼具泄洪和调节水位功能。",
		CulturalValue:          "2001年被列为全国重点文物保护单位，是徽州文化的重要象征之一。",
	},
	"modern_gravity": {
		DamKey:                 "modern_gravity",
		DamName:                "现代混凝土重力坝",
		DamType:                models.DamTypeModernConcrete,
		BuildDynasty:           "当代",
		BuildYear:              2020,
		Location:               "参照现代水利工程标准设计",
		Description:            "采用现代水工设计规范的混凝土重力坝，应用有限元分析、新材料、新工艺，代表当代水利工程的最高水平。",
		HistoricalSignificance: "作为跨时代对比的参照基准，展示从古代条石坝到现代混凝土坝的技术进步。",
		Geometry: models.DamGeometry{
			Length:          113.7,
			Height:          15.0,
			TopWidth:        8.0,
			UpstreamSlope:   0.15,
			DownstreamSlope: 0.75,
		},
		FoundationDepth:        12.0,
		DesignUpstreamWL:       13.5,
		DesignDownstreamWL:     3.2,
		MaterialType:           "C25碾压混凝土+防渗面板+基础灌浆帷幕",
		OriginalPermeability:   1.0e-11,
		CurrentPermeability:    1.0e-11,
		FoundationPermeability: 1.0e-8,
		InterfacePermeability:  1.0e-10,
		HasAntiSeepageSystem:   true,
		AntiSeepageDescription: "上游面设60cm厚防渗面板，坝基进行帷幕灌浆，设排水廊道系统。坝体内部设温度控制冷却水管，防渗体系完善。",
		CulturalValue:          "代表21世纪水利工程技术水平，作为科技进步的参照基准。",
	},
}

var virtualTourScenes = map[string][]*models.VirtualTourScene{
	"tashan_weir": {
		{
			SceneID:      "overview",
			SceneName:    "它山堰全景",
			Description:  "俯瞰它山堰全貌，感受千年前水利工程的宏伟气势",
			CameraPos:    models.Point3D{X: 60, Y: 20, Z: 80},
			CameraTarget: models.Point3D{X: 56.85, Y: 1.925, Z: 0},
			Narrative:    "各位游客，欢迎来到它山堰！您眼前的这座古堰始建于唐朝大和七年，也就是公元833年，由当时的鄮县县令王元暐主持修建。它山堰全长113.7米，高3.85米，是中国古代四大水利工程之一。",
			Hotspots: []models.Hotspot{
				{HotspotID: "hs1", Position: models.Point3D{X: 56.85, Y: 3.85, Z: 0}, Title: "堰顶", Description: "堰顶宽4.8米，采用巨型条石砌筑，每块重约吨。"},
				{HotspotID: "hs2", Position: models.Point3D{X: 10, Y: 2, Z: 0}, Title: "上游护坦", Description: "上游设护坦保护坝基免受水流冲刷。"},
			},
		},
		{
			SceneID:      "upstream_view",
			SceneName:    "上游视角",
			Description:  "站在上游岸边，观察水流与堰体的相互作用",
			CameraPos:    models.Point3D{X: -30, Y: 5, Z: 30},
			CameraTarget: models.Point3D{X: 20, Y: 2, Z: 0},
			Narrative:    "现在我们来到了它山堰的上游。您可以看到，当溪水流经堰体时，被抬高的水位形成了落差。这种设计巧妙地解决了鄞西平原的灌溉用水问题，同时又能在洪水期让水流漫过堰顶，排泄到奉化江。",
			Hotspots: []models.Hotspot{
				{HotspotID: "hs3", Position: models.Point3D{X: 0, Y: 6, Z: 0}, Title: "水位", Description: "当前上游水位8.5米，远超堰顶高程。"},
			},
		},
		{
			SceneID:      "seepage_cutaway",
			SceneName:    "渗流剖面图",
			Description:  "透视坝体内部，观察地下水的渗流路径",
			CameraPos:    models.Point3D{X: 56.85, Y: 10, Z: 50},
			CameraTarget: models.Point3D{X: 56.85, Y: 0, Z: 0},
			Narrative:    "现在我们切换到透视模式，看看坝体内部的情况。您看到的彩色云图是坝体内部的扬压力分布，红色代表高压力，蓝色代表低压力。那些流动的粒子就是地下水的渗流路径。可以看到，防渗铺盖有效地延长了渗流路径，降低了扬压力。",
			Hotspots: []models.Hotspot{
				{HotspotID: "hs4", Position: models.Point3D{X: 30, Y: -2, Z: 0}, Title: "防渗铺盖", Description: "紫色区域是防渗铺盖，渗透系数仅为坝体的1%。"},
				{HotspotID: "hs5", Position: models.Point3D{X: 70, Y: 0, Z: 0}, Title: "渗流出口", Description: "渗流从下游坝脚渗出，流速最快。"},
			},
		},
		{
			SceneID:      "downstream_view",
			SceneName:    "下游视角",
			Description:  "观察下游消能设施和渗流出口",
			CameraPos:    models.Point3D{X: 150, Y: 3, Z: -20},
			CameraTarget: models.Point3D{X: 100, Y: 2, Z: 0},
			Narrative:    "现在我们来到下游。可以看到水流从堰顶跌落，通过下游护坦消能。仔细观察坝脚，您会发现有细小的水流渗出——这就是坝体的渗流。正常的渗流是允许的，但如果渗流量过大，就可能危及坝体安全。",
			Hotspots: []models.Hotspot{
				{HotspotID: "hs6", Position: models.Point3D{X: 113.7, Y: 0, Z: 0}, Title: "坝脚渗流", Description: "渗流从这里逸出，当前渗流量8.5 L/s，在安全范围内。"},
			},
		},
		{
			SceneID:      "sensor_layout",
			SceneName:    "传感器布置",
			Description:  "了解现代化监测系统如何守护千年古堰",
			CameraPos:    models.Point3D{X: 56.85, Y: 15, Z: 60},
			CameraTarget: models.Point3D{X: 56.85, Y: 0, Z: 0},
			Narrative:    "为了守护这座千年古堰，我们安装了15个现代化传感器。5个扬压力计监测坝体内部的水压力，1个渗流量计监测总渗流量，2个水位计监测上下游水位，2个冲刷深度计监测坝基冲刷，还有5个浸润线测点监测自由面位置。所有数据实时传输到监控中心。",
			Hotspots: []models.Hotspot{
				{HotspotID: "hs7", Position: models.Point3D{X: 40, Y: 1, Z: 0}, Title: "PZ-003 扬压力计", Description: "当前读数42.3 kPa，在预警阈值50 kPa以下，正常。"},
				{HotspotID: "hs8", Position: models.Point3D{X: 110, Y: 0.5, Z: 0}, Title: "SM-001 渗流量计", Description: "当前读数8.5 L/s，正常范围。"},
			},
		},
	},
}

func GetAllDamPresets() []*models.DamPreset {
	result := make([]*models.DamPreset, 0, len(damPresets))
	for _, v := range damPresets {
		result = append(result, v)
	}
	return result
}

func GetDamPreset(damKey string) *models.DamPreset {
	return damPresets[damKey]
}

func GetDamKeys() []string {
	keys := make([]string, 0, len(damPresets))
	for k := range damPresets {
		keys = append(keys, k)
	}
	return keys
}

func GetVirtualTourScenes(damKey string) []*models.VirtualTourScene {
	return virtualTourScenes[damKey]
}

func GetDamTypeLabel(damType models.DamType) string {
	switch damType {
	case models.DamTypeAncientStone:
		return "古代条石坝"
	case models.DamTypeModernConcrete:
		return "现代混凝土重力坝"
	default:
		return "未知类型"
	}
}
