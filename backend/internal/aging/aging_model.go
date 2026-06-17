package aging

import (
	"fmt"
	"math"
	"time"

	"tashan-weir-seepage/internal/dam_presets"
	"tashan-weir-seepage/internal/models"
	"tashan-weir-seepage/internal/simulation"
)

type AgingModel struct {
	ActivationEnergy   float64
	TemperatureRef     float64
	GasConstant        float64
	InitialDamage      float64
	DamageGrowthRate   float64
	WeatheringFactor   float64
	BiologicalFactor   float64
}

func NewAgingModel() *AgingModel {
	return &AgingModel{
		ActivationEnergy: 45000,
		TemperatureRef:   293.15,
		GasConstant:      8.314,
		InitialDamage:    0.0,
		DamageGrowthRate: 0.005,
		WeatheringFactor: 0.003,
		BiologicalFactor: 0.002,
	}
}

func (m *AgingModel) CalculatePermeabilityEvolution(
	initialK float64,
	initialAgeYears int,
	predictionYears int,
	timeStepYears int,
	considerClimate bool,
	maintenanceFreq string,
) []float64 {
	numPoints := predictionYears/timeStepYears + 1
	permeabilities := make([]float64, numPoints)

	currentK := initialK
	currentDamage := m.InitialDamage

	maintenanceFactor := getMaintenanceFactor(maintenanceFreq)
	climateFactor := 1.0
	if considerClimate {
		climateFactor = 1.3
	}

	for i := 0; i < numPoints; i++ {
		year := float64(i * timeStepYears)
		totalAge := float64(initialAgeYears) + year

		ageFactor := math.Log10(totalAge + 1)

		damageGrowth := m.DamageGrowthRate * ageFactor * climateFactor
		weathering := m.WeatheringFactor * math.Sqrt(year+1) * climateFactor
		biological := m.BiologicalFactor * (1 - math.Exp(-year/50))

		currentDamage += (damageGrowth + weathering + biological) * float64(timeStepYears)
		currentDamage *= maintenanceFactor

		tempFactor := 1.0
		if considerClimate {
			seasonalVariation := 1.0 + 0.05*math.Sin(2*math.Pi*year/10)
			tempFactor = seasonalVariation
		}

		arrheniusFactor := math.Exp(-m.ActivationEnergy/m.GasConstant * (1/(m.TemperatureRef+5) - 1/m.TemperatureRef))

		kRatio := 1.0 + currentDamage*arrheniusFactor*tempFactor
		permeabilities[i] = currentK * kRatio

		if i > 0 {
			prevK := permeabilities[i-1]
			increment := kRatio * maintenanceFactor
			permeabilities[i] = math.Max(prevK, prevK*increment)
		}

		maxK := initialK * 1000.0
		if permeabilities[i] > maxK {
			permeabilities[i] = maxK
		}
	}

	return permeabilities
}

func getMaintenanceFactor(freq string) float64 {
	switch freq {
	case "high":
		return 0.3
	case "medium":
		return 0.6
	case "low":
		return 0.85
	default:
		return 1.0
	}
}

func PredictAging(req *models.AgingPredictionRequest) (*models.AgingPredictionResult, error) {
	startTime := time.Now()

	preset := dam_presets.GetDamPreset(req.DamKey)
	if preset == nil {
		return nil, fmt.Errorf("dam not found: %s", req.DamKey)
	}

	currentYear := time.Now().Year()
	initialAge := currentYear - preset.BuildYear
	if initialAge < 0 {
		initialAge = 0
	}

	initialK := req.InitialPermeability
	if initialK <= 0 {
		initialK = preset.CurrentPermeability
	}

	timeStep := req.TimeStepYears
	if timeStep <= 0 {
		timeStep = 5
	}

	predYears := req.PredictionYears
	if predYears <= 0 {
		predYears = 100
	}

	model := NewAgingModel()
	kEvolution := model.CalculatePermeabilityEvolution(
		initialK,
		initialAge,
		predYears,
		timeStep,
		req.ConsiderClimate,
		req.MaintenanceFrequency,
	)

	dataPoints := make([]models.AgingDataPoint, 0, len(kEvolution))
	criticalYear := 0
	baselineFlow := 0.0

	upWL := preset.DesignUpstreamWL
	downWL := preset.DesignDownstreamWL

	for i, k := range kEvolution {
		year := currentYear + i*timeStep
		ageYears := initialAge + i*timeStep

		solver := simulation.NewSeepageSolverFromPreset(preset)
		solver.PermeabilityK = k
		solver.SetGridResolution(40, 20)

		simResult, _, err := solver.Run(upWL, downWL, fmt.Sprintf("aging_year_%d", year))
		if err != nil {
			continue
		}

		if i == 0 {
			baselineFlow = simResult.TotalSeepageFlow
		}

		kRatio := k / initialK
		flowRatio := simResult.TotalSeepageFlow / baselineFlow
		agingDegree := math.Min(100.0, (kRatio-1.0)*20.0)

		failureProb := calculateFailureProbability(kRatio, flowRatio, ageYears)

		action := getRecommendedAction(agingDegree, failureProb)

		dp := models.AgingDataPoint{
			Year:               year,
			AgeYears:           ageYears,
			Permeability:       k,
			PermeabilityRatio:  kRatio,
			SeepageFlow:        simResult.TotalSeepageFlow,
			SeepageFlowRatio:   flowRatio,
			MaxPorePressure:    simResult.MaxPorePressure,
			DegreeOfAging:      agingDegree,
			FailureProbability: failureProb,
			RecommendedAction:  action,
		}

		dataPoints = append(dataPoints, dp)

		if criticalYear == 0 && failureProb > 0.5 {
			criticalYear = year
		}
	}

	agingRate := 0.0
	if len(dataPoints) > 1 {
		firstK := dataPoints[0].Permeability
		lastK := dataPoints[len(dataPoints)-1].Permeability
		agingRate = (lastK - firstK) / float64(predYears)
	}

	summary := generateAgingSummary(preset.DamName, initialAge, predYears, dataPoints, criticalYear)
	recommendations := generateRecommendations(preset.DamName, dataPoints, req.MaintenanceFrequency)

	calcTime := time.Since(startTime).Milliseconds()

	result := &models.AgingPredictionResult{
		DamKey:            req.DamKey,
		DamName:           preset.DamName,
		InitialAge:        initialAge,
		PredictionYears:   predYears,
		DataPoints:        dataPoints,
		AgingRate:         agingRate,
		CriticalYear:      criticalYear,
		Summary:           summary,
		Recommendations:   recommendations,
		UpstreamWL:        upWL,
		DownstreamWL:      downWL,
		CalculationTimeMs: calcTime,
	}

	return result, nil
}

func calculateFailureProbability(kRatio, flowRatio float64, ageYears int) float64 {
	kFactor := math.Min(1.0, (kRatio-1.0)/10.0)
	flowFactor := math.Min(1.0, (flowRatio-1.0)/5.0)
	ageFactor := math.Min(1.0, float64(ageYears)/1000.0)

	prob := 0.1*kFactor + 0.4*flowFactor + 0.2*ageFactor
	prob = math.Max(0.0, math.Min(1.0, prob))

	if kRatio > 100 {
		prob = math.Max(prob, 0.8)
	}
	if flowRatio > 3 {
		prob = math.Max(prob, 0.7)
	}

	return prob
}

func getRecommendedAction(agingDegree, failureProb float64) string {
	switch {
	case failureProb > 0.8:
		return "紧急加固：立即进行防渗加固工程，建议停止高水位运行"
	case failureProb > 0.5:
		return "重点维修：实施防渗灌浆，增设防渗铺盖，加强监测频率"
	case failureProb > 0.2:
		return "定期检查：每季度进行渗流监测，每年进行安全评估"
	case agingDegree > 30:
		return "预防性维护：进行表面防渗处理，清理排水系统"
	default:
		return "正常监测：按常规频率进行监测和维护"
	}
}

func generateAgingSummary(damName string, initialAge, predYears int, dataPoints []models.AgingDataPoint, criticalYear int) string {
	if len(dataPoints) == 0 {
		return "无数据"
	}

	first := dataPoints[0]
	last := dataPoints[len(dataPoints)-1]

	kIncrease := (last.Permeability - first.Permeability) / first.Permeability * 100
	flowIncrease := (last.SeepageFlow - first.SeepageFlow) / first.SeepageFlow * 100

	summary := fmt.Sprintf("%s现已建成%d年，预测未来%d年的老化演变趋势：", damName, initialAge, predYears)
	summary += fmt.Sprintf("渗透系数将从%.2e m/s增加到%.2e m/s（增长%.1f%%），",
		first.Permeability, last.Permeability, kIncrease)
	summary += fmt.Sprintf("渗流量将从%.4f L/s增加到%.4f L/s（增长%.1f%%）。",
		first.SeepageFlow*1000, last.SeepageFlow*1000, flowIncrease)

	if criticalYear > 0 {
		summary += fmt.Sprintf("预计在%d年达到临界失效风险（>50%%），建议提前进行加固处理。", criticalYear)
	} else {
		summary += "在预测期内失效风险较低，维持正常维护即可。"
	}

	return summary
}

func generateRecommendations(damName string, dataPoints []models.AgingDataPoint, maintenanceFreq string) []string {
	if len(dataPoints) == 0 {
		return []string{}
	}

	recommendations := []string{}

	last := dataPoints[len(dataPoints)-1]

	if last.DegreeOfAging > 50 {
		recommendations = append(recommendations,
			fmt.Sprintf("建议对%s进行全面防渗性能评估，考虑增设防渗帷幕或防渗铺盖", damName))
	}

	if last.FailureProbability > 0.3 {
		recommendations = append(recommendations,
			"建议加密渗流监测频率，从每月1次增加到每两周1次")
	}

	if maintenanceFreq == "none" || maintenanceFreq == "low" {
		recommendations = append(recommendations,
			"建议提高维护频率，定期进行坝体表面检查和排水系统清理")
	}

	recommendations = append(recommendations,
		"建立坝体健康档案，记录每年的渗流监测数据，建立长期演变趋势分析")

	recommendations = append(recommendations,
		"建议安装自动化监测系统，实现扬压力、渗流量的实时监测和预警")

	if last.DegreeOfAging > 30 {
		recommendations = append(recommendations,
			"考虑采用现代防渗技术进行加固，如高压喷射灌浆、土工膜防渗等")
	}

	return recommendations
}

func CompareAgingScenarios(damKey string, baseRequest *models.AgingPredictionRequest) (map[string]*models.AgingPredictionResult, error) {
	scenarios := map[string]*models.AgingPredictionResult{}

	scenarios["baseline"], _ = PredictAging(baseRequest)

	highMaintenanceReq := *baseRequest
	highMaintenanceReq.MaintenanceFrequency = "high"
	scenarios["high_maintenance"], _ = PredictAging(&highMaintenanceReq)

	noMaintenanceReq := *baseRequest
	noMaintenanceReq.MaintenanceFrequency = "none"
	scenarios["no_maintenance"], _ = PredictAging(&noMaintenanceReq)

	climateReq := *baseRequest
	climateReq.ConsiderClimate = true
	climateReq.ConsiderMaintenance = true
	scenarios["with_climate"], _ = PredictAging(&climateReq)

	return scenarios, nil
}
