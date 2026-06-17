package models

import (
	"time"
)

type DamInfo struct {
	ID                      int       `json:"id"`
	DamName                 string    `json:"dam_name"`
	BuildDynasty            string    `json:"build_dynasty"`
	BuildYear               int       `json:"build_year"`
	DamLength               float64   `json:"dam_length"`
	DamHeight               float64   `json:"dam_height"`
	DamTopWidth             float64   `json:"dam_top_width"`
	DamBottomWidth          float64   `json:"dam_bottom_width"`
	UpstreamSlope           float64   `json:"upstream_slope"`
	DownstreamSlope         float64   `json:"downstream_slope"`
	DesignUpstreamWaterLevel float64  `json:"design_upstream_water_level"`
	DesignDownstreamWaterLevel float64 `json:"design_downstream_water_level"`
	MaterialType            string    `json:"material_type"`
	PermeabilityCoefficient float64   `json:"permeability_coefficient"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

type SensorConfig struct {
	SensorID          string    `json:"sensor_id"`
	SensorType        string    `json:"sensor_type"`
	SensorName        string    `json:"sensor_name"`
	LocationX         float64   `json:"location_x"`
	LocationY         float64   `json:"location_y"`
	LocationZ         float64   `json:"location_z"`
	InstallationDate  string    `json:"installation_date"`
	WarningThreshold  *float64  `json:"warning_threshold"`
	DangerThreshold   *float64  `json:"danger_threshold"`
	Unit              string    `json:"unit"`
	IsActive          bool      `json:"is_active"`
	DTUID             string    `json:"dtu_id"`
	CreatedAt         time.Time `json:"created_at"`
}

type SensorData struct {
	Time        time.Time            `json:"time"`
	SensorID    string               `json:"sensor_id"`
	SensorValue float64              `json:"sensor_value"`
	Quality     int                  `json:"quality"`
	RawData     map[string]interface{} `json:"raw_data,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
}

type DTUPayload struct {
	DTUID     string       `json:"dtu_id"`
	Timestamp time.Time    `json:"timestamp"`
	Sensors   []SensorData `json:"sensors"`
	Signal    float64      `json:"signal_strength"`
	Battery   float64      `json:"battery_level"`
}

type AlarmRecord struct {
	ID             int64     `json:"id"`
	AlarmTime      time.Time `json:"alarm_time"`
	AlarmLevel     string    `json:"alarm_level"`
	AlarmType      string    `json:"alarm_type"`
	SensorID       *string   `json:"sensor_id"`
	SensorValue    *float64  `json:"sensor_value"`
	ThresholdValue *float64  `json:"threshold_value"`
	AlarmMessage   string    `json:"alarm_message"`
	IsHandled      bool      `json:"is_handled"`
	HandledBy      *string   `json:"handled_by"`
	HandledTime    *time.Time `json:"handled_time"`
	HandleNote     *string   `json:"handle_note"`
	MQTTPublished  bool      `json:"mqtt_published"`
	MQTTTopic      *string   `json:"mqtt_topic"`
	CreatedAt      time.Time `json:"created_at"`
}

type SeepageSimulation struct {
	ID                  int64                  `json:"id"`
	SimulationName      string                 `json:"simulation_name"`
	UpstreamWaterLevel  float64                `json:"upstream_water_level"`
	DownstreamWaterLevel float64               `json:"downstream_water_level"`
	SimulationTime      time.Time              `json:"simulation_time"`
	TotalSeepageFlow    float64                `json:"total_seepage_flow"`
	MaxPorePressure     float64                `json:"max_pore_pressure"`
	GridCount           int                    `json:"grid_count"`
	CalculationTimeMs   int64                  `json:"calculation_time_ms"`
	Parameters          map[string]interface{} `json:"parameters"`
	ResultSummary       map[string]interface{} `json:"result_summary"`
	CreatedAt           time.Time              `json:"created_at"`
}

type SimulationGrid struct {
	ID                int64   `json:"id"`
	SimulationID      int64   `json:"simulation_id"`
	GridX             float64 `json:"grid_x"`
	GridY             float64 `json:"grid_y"`
	WaterHead         float64 `json:"water_head"`
	PorePressure      float64 `json:"pore_pressure"`
	VelocityX         float64 `json:"velocity_x"`
	VelocityY         float64 `json:"velocity_y"`
	VelocityMagnitude float64 `json:"velocity_magnitude"`
	IsSaturated       bool    `json:"is_saturated"`
}

type OptimizationResult struct {
	ID                    int64                  `json:"id"`
	OptimizationName      string                 `json:"optimization_name"`
	Algorithm             string                 `json:"algorithm"`
	UpstreamWaterLevel    float64                `json:"upstream_water_level"`
	DownstreamWaterLevel  float64                `json:"downstream_water_level"`
	BlanketLength         float64                `json:"blanket_length"`
	BlanketThickness      float64                `json:"blanket_thickness"`
	BlanketPermeability   float64                `json:"blanket_permeability"`
	OptimizedSeepageFlow  float64                `json:"optimized_seepage_flow"`
	BaselineSeepageFlow   float64                `json:"baseline_seepage_flow"`
	FlowReductionRate     float64                `json:"flow_reduction_rate"`
	GenerationCount       int                    `json:"generation_count"`
	PopulationSize        int                    `json:"population_size"`
	BestFitness           float64                `json:"best_fitness"`
	OptimizationTimeMs    int64                  `json:"optimization_time_ms"`
	Parameters            map[string]interface{} `json:"parameters"`
	ConvergenceCurve      []float64              `json:"convergence_curve"`
	CreatedAt             time.Time              `json:"created_at"`
}

type SimulationRequest struct {
	UpstreamWaterLevel   float64                `json:"upstream_water_level"`
	DownstreamWaterLevel float64                `json:"downstream_water_level"`
	GridResolutionX      int                    `json:"grid_resolution_x"`
	GridResolutionY      int                    `json:"grid_resolution_y"`
	PermeabilityK        float64                `json:"permeability_k"`
	BlanketLength        *float64               `json:"blanket_length,omitempty"`
	BlanketThickness     *float64               `json:"blanket_thickness,omitempty"`
	BlanketPermeability  *float64               `json:"blanket_permeability,omitempty"`
	SimulationName       string                 `json:"simulation_name"`
	Parameters           map[string]interface{} `json:"parameters"`
}

type OptimizationRequest struct {
	UpstreamWaterLevel   float64 `json:"upstream_water_level"`
	DownstreamWaterLevel float64 `json:"downstream_water_level"`
	MinBlanketLength     float64 `json:"min_blanket_length"`
	MaxBlanketLength     float64 `json:"max_blanket_length"`
	MinBlanketThickness  float64 `json:"min_blanket_thickness"`
	MaxBlanketThickness  float64 `json:"max_blanket_thickness"`
	PopulationSize       int     `json:"population_size"`
	MaxGenerations       int     `json:"max_generations"`
	MutationRate         float64 `json:"mutation_rate"`
	CrossoverRate        float64 `json:"crossover_rate"`
	OptimizationName     string  `json:"optimization_name"`
}

type ParetoSolution struct {
	BlanketLength    float64 `json:"blanket_length"`
	BlanketThickness float64 `json:"blanket_thickness"`
	SeepageFlow      float64 `json:"seepage_flow"`
	MaterialCost     float64 `json:"material_cost"`
	FlowReduction    float64 `json:"flow_reduction"`
	Rank             int     `json:"rank"`
}

// ===== 新增Feature: 多堰坝对比 =====

type DamType string

const (
	DamTypeAncientStone   DamType = "ancient_stone"   // 古代条石坝
	DamTypeModernConcrete DamType = "modern_concrete" // 现代混凝土重力坝
)

type DamPreset struct {
	DamKey                   string      `json:"dam_key"`
	DamName                  string      `json:"dam_name"`
	DamType                  DamType     `json:"dam_type"`
	BuildDynasty             string      `json:"build_dynasty"`
	BuildYear                int         `json:"build_year"`
	Location                 string      `json:"location"`
	Description              string      `json:"description"`
	HistoricalSignificance   string      `json:"historical_significance"`
	Geometry                 DamGeometry `json:"geometry"`
	FoundationDepth          float64     `json:"foundation_depth"`
	DesignUpstreamWL         float64     `json:"design_upstream_wl"`
	DesignDownstreamWL       float64     `json:"design_downstream_wl"`
	MaterialType             string      `json:"material_type"`
	OriginalPermeability     float64     `json:"original_permeability"`
	CurrentPermeability      float64     `json:"current_permeability"`
	FoundationPermeability   float64     `json:"foundation_permeability"`
	InterfacePermeability    float64     `json:"interface_permeability"`
	HasAntiSeepageSystem     bool        `json:"has_anti_seepage_system"`
	AntiSeepageDescription   string      `json:"anti_seepage_description"`
	CulturalValue            string      `json:"cultural_value"`
	ImageURL                 string      `json:"image_url,omitempty"`
}

type DamGeometry struct {
	Length          float64 `json:"length"`
	Height          float64 `json:"height"`
	TopWidth        float64 `json:"top_width"`
	UpstreamSlope   float64 `json:"upstream_slope"`
	DownstreamSlope float64 `json:"downstream_slope"`
}

// ===== 新增Feature: 坝体老化预测 =====

type AgingPredictionRequest struct {
	DamKey               string  `json:"dam_key"`
	PredictionYears      int     `json:"prediction_years"`
	TimeStepYears        int     `json:"time_step_years"`
	InitialPermeability  float64 `json:"initial_permeability"`
	ConsiderClimate      bool    `json:"consider_climate"`
	ConsiderMaintenance  bool    `json:"consider_maintenance"`
	MaintenanceFrequency string  `json:"maintenance_frequency"` // none, low, medium, high
}

type AgingDataPoint struct {
	Year                   int     `json:"year"`
	AgeYears               int     `json:"age_years"`
	Permeability           float64 `json:"permeability"`
	PermeabilityRatio      float64 `json:"permeability_ratio"`
	SeepageFlow            float64 `json:"seepage_flow"`
	SeepageFlowRatio       float64 `json:"seepage_flow_ratio"`
	MaxPorePressure        float64 `json:"max_pore_pressure"`
	DegreeOfAging          float64 `json:"degree_of_aging"` // 0-100%
	FailureProbability     float64 `json:"failure_probability"`
	RecommendedAction      string  `json:"recommended_action"`
}

type AgingPredictionResult struct {
	DamKey             string            `json:"dam_key"`
	DamName            string            `json:"dam_name"`
	InitialAge         int               `json:"initial_age_years"`
	PredictionYears    int               `json:"prediction_years"`
	DataPoints         []AgingDataPoint  `json:"data_points"`
	AgingRate          float64           `json:"aging_rate"`
	CriticalYear       int               `json:"critical_year,omitempty"`
	Summary            string            `json:"summary"`
	Recommendations    []string          `json:"recommendations"`
	UpstreamWL         float64           `json:"upstream_wl"`
	DownstreamWL       float64           `json:"downstream_wl"`
	CalculationTimeMs  int64             `json:"calculation_time_ms"`
}

// ===== 新增Feature: 对比分析 =====

type DamComparisonRequest struct {
	DamKeys             []string `json:"dam_keys"`
	UpstreamWaterLevel  float64  `json:"upstream_water_level"`
	DownstreamWaterLevel float64 `json:"downstream_water_level"`
	GridResolutionX     int      `json:"grid_resolution_x"`
	GridResolutionY     int      `json:"grid_resolution_y"`
	IncludeCurrentAging bool     `json:"include_current_aging"`
}

type DamComparisonItem struct {
	DamKey              string                  `json:"dam_key"`
	DamName             string                  `json:"dam_name"`
	DamType             DamType                 `json:"dam_type"`
	BuildDynasty        string                  `json:"build_dynasty"`
	Geometry            DamGeometry             `json:"geometry"`
	Permeability        float64                 `json:"permeability"`
	TotalSeepageFlow    float64                 `json:"total_seepage_flow"`
	SeepageFlowPerMeter float64                 `json:"seepage_flow_per_meter"`
	MaxPorePressure     float64                 `json:"max_pore_pressure"`
	AvgPorePressure     float64                 `json:"avg_pore_pressure"`
	InfiltrationLine    []Point2D               `json:"infiltration_line"`
	ExitGradient        float64                 `json:"exit_gradient"`
	UpliftForce         float64                 `json:"uplift_force"`
	AntiSeepageEfficiency float64               `json:"anti_seepage_efficiency"`
	Simulation          *SeepageSimulation      `json:"simulation,omitempty"`
	Grids               []SimulationGrid        `json:"grids,omitempty"`
}

type Point2D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type DamComparisonResult struct {
	RequestID            string                `json:"request_id"`
	ComparisonName       string                `json:"comparison_name"`
	UpstreamWaterLevel   float64               `json:"upstream_water_level"`
	DownstreamWaterLevel float64               `json:"downstream_water_level"`
	Items                []DamComparisonItem   `json:"items"`
	Summary              map[string]interface{} `json:"summary"`
	CalculationTimeMs    int64                 `json:"calculation_time_ms"`
}

// ===== 新增Feature: 跨时代对比 =====

type CrossEraComparisonRequest struct {
	AncientDamKey    string  `json:"ancient_dam_key"`
	ModernDamKey     string  `json:"modern_dam_key"`
	UpstreamWL       float64 `json:"upstream_wl"`
	DownstreamWL     float64 `json:"downstream_wl"`
	ScaleToSameSize  bool    `json:"scale_to_same_size"`
}

type CrossEraComparisonResult struct {
	AncientDam     DamComparisonItem      `json:"ancient_dam"`
	ModernDam      DamComparisonItem      `json:"modern_dam"`
	AncientMetrics map[string]interface{} `json:"ancient_metrics"`
	ModernMetrics  map[string]interface{} `json:"modern_metrics"`
	Comparison     map[string]interface{} `json:"comparison"`
	Insights       []string               `json:"insights"`
}

// ===== 新增Feature: 虚拟参观 =====

type VirtualTourScene struct {
	SceneID      string   `json:"scene_id"`
	SceneName    string   `json:"scene_name"`
	Description  string   `json:"description"`
	CameraPos    Point3D  `json:"camera_position"`
	CameraTarget Point3D  `json:"camera_target"`
	Hotspots     []Hotspot `json:"hotspots"`
	Narrative    string   `json:"narrative"`
}

type Point3D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type Hotspot struct {
	HotspotID   string  `json:"hotspot_id"`
	Position    Point3D `json:"position"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	MediaType   string  `json:"media_type"` // text, image, video
	MediaURL    string  `json:"media_url,omitempty"`
}

type VirtualTourRequest struct {
	DamKey        string  `json:"dam_key"`
	WaterLevel    float64 `json:"water_level"`
	Season        string  `json:"season"` // spring, summer, autumn, winter
	TimeOfDay     string  `json:"time_of_day"` // morning, afternoon, evening, night
	ShowSeepage   bool    `json:"show_seepage"`
	ShowSensors   bool    `json:"show_sensors"`
}

type InteractiveAdjustmentRequest struct {
	DamKey           string  `json:"dam_key"`
	UpstreamWL       float64 `json:"upstream_wl"`
	DownstreamWL     float64 `json:"downstream_wl"`
	HighlightArea    string  `json:"highlight_area,omitempty"` // none, upstream, downstream, foundation, blanket
	VisualizationMode string `json:"visualization_mode"` // pressure, velocity, streamline, both
}

type InteractiveAdjustmentResult struct {
	Simulation      *SeepageSimulation `json:"simulation"`
	Grids           []SimulationGrid   `json:"grids"`
	KeyMetrics      map[string]float64 `json:"key_metrics"`
	WaterLevelChange string            `json:"water_level_change"`
	RiskLevel       string            `json:"risk_level"` // low, medium, high, critical
	Explanation     string            `json:"explanation"`
}
