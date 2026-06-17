const DamComparison = (function() {
    let allDams = [];
    let selectedDams = new Set();
    let seepageCompareChart = null;
    let pressureCompareChart = null;
    let efficiencyCompareChart = null;

    async function init() {
        await loadDams();
        renderDamSelector();
        bindEvents();
        initCharts();
    }

    async function loadDams() {
        try {
            const data = await SeepagePanel.apiGet('/dams');
            allDams = data.dams || [];
            allDams.forEach(dam => {
                if (dam.dam_key === 'tashan_weir' || dam.dam_key === 'mulan_bei') {
                    selectedDams.add(dam.dam_key);
                }
            });
        } catch (e) {
            console.error('Failed to load dams:', e);
        }
    }

    function renderDamSelector() {
        const container = document.getElementById('damSelector');
        if (!container) return;

        container.innerHTML = allDams.map(dam => `
            <label class="dam-checkbox ${selectedDams.has(dam.dam_key) ? 'selected' : ''}">
                <input type="checkbox" value="${dam.dam_key}" 
                    ${selectedDams.has(dam.dam_key) ? 'checked' : ''}>
                <div class="dam-info">
                    <span class="dam-name">${dam.dam_name}</span>
                    <span class="dam-dynasty">${dam.build_dynasty}</span>
                </div>
            </label>
        `).join('');

        container.querySelectorAll('input[type="checkbox"]').forEach(cb => {
            cb.addEventListener('change', (e) => {
                const key = e.target.value;
                if (e.target.checked) {
                    selectedDams.add(key);
                    cb.closest('.dam-checkbox').classList.add('selected');
                } else {
                    selectedDams.delete(key);
                    cb.closest('.dam-checkbox').classList.remove('selected');
                }
            });
        });
    }

    function bindEvents() {
        const btnCompare = document.getElementById('btnCompareDams');
        if (btnCompare) {
            btnCompare.addEventListener('click', runComparison);
        }

        const btnCrossEra = document.getElementById('btnCrossEra');
        if (btnCrossEra) {
            btnCrossEra.addEventListener('click', runCrossEraComparison);
        }
    }

    function initCharts() {
        const commonOptions = {
            responsive: true,
            maintainAspectRatio: false,
            plugins: { legend: { position: 'bottom' } },
            scales: { y: { beginAtZero: true } }
        };

        const seepageCtx = document.getElementById('seepageCompareChart');
        if (seepageCtx) {
            seepageCompareChart = new Chart(seepageCtx, {
                type: 'bar',
                data: { labels: [], datasets: [] },
                options: commonOptions
            });
        }

        const pressureCtx = document.getElementById('pressureCompareChart');
        if (pressureCtx) {
            pressureCompareChart = new Chart(pressureCtx, {
                type: 'bar',
                data: { labels: [], datasets: [] },
                options: commonOptions
            });
        }

        const efficiencyCtx = document.getElementById('efficiencyCompareChart');
        if (efficiencyCtx) {
            efficiencyCompareChart = new Chart(efficiencyCtx, {
                type: 'bar',
                data: { labels: [], datasets: [] },
                options: commonOptions
            });
        }
    }

    async function runComparison() {
        if (selectedDams.size < 2) {
            SeepagePanel.showNotification('请至少选择2座堰坝进行对比', 'warning');
            return;
        }

        const upWL = parseFloat(document.getElementById('cmpUpstream').value);
        const downWL = parseFloat(document.getElementById('cmpDownstream').value);
        const resolution = parseInt(document.getElementById('cmpResolution').value);
        const includeAging = document.getElementById('cmpIncludeAging').checked;

        SeepagePanel.showLoading('正在对比分析多座堰坝...', 10);

        try {
            const result = await SeepagePanel.apiPost('/compare/dams', {
                dam_keys: Array.from(selectedDams),
                upstream_water_level: upWL,
                downstream_water_level: downWL,
                grid_resolution_x: resolution,
                grid_resolution_y: Math.round(resolution / 2),
                include_current_aging: includeAging
            });

            SeepagePanel.hideLoading();
            renderComparisonResult(result);
        } catch (e) {
            SeepagePanel.hideLoading();
            SeepagePanel.showNotification('对比分析失败: ' + e.message, 'danger');
        }
    }

    function renderComparisonResult(result) {
        const summaryEl = document.getElementById('comparisonSummary');
        const summaryContent = document.getElementById('summaryContent');
        const cardsContainer = document.getElementById('comparisonCards');

        summaryEl.style.display = 'block';
        summaryContent.innerHTML = `
            <div class="summary-grid">
                <div class="summary-item">
                    <span class="summary-label">对比坝数</span>
                    <span class="summary-value">${result.items.length} 座</span>
                </div>
                <div class="summary-item">
                    <span class="summary-label">水位差</span>
                    <span class="summary-value">${(result.upstream_water_level - result.downstream_water_level).toFixed(1)} m</span>
                </div>
                <div class="summary-item">
                    <span class="summary-label">渗流量范围</span>
                    <span class="summary-value">${(result.summary.min_seepage_flow_lps || 0).toFixed(2)} - ${(result.summary.max_seepage_flow_lps || 0).toFixed(2)} L/s</span>
                </div>
                <div class="summary-item">
                    <span class="summary-label">最优防渗坝</span>
                    <span class="summary-value highlight">${result.summary.best_anti_seepage_dam || '-'}</span>
                </div>
            </div>
            ${result.summary.key_insights ? `
                <div class="insights">
                    <h5>💡 关键洞察</h5>
                    <ul>${result.summary.key_insights.map(i => `<li>${i}</li>`).join('')}</ul>
                </div>
            ` : ''}
        `;

        const labels = result.items.map(item => item.dam_name);
        const colors = ['#00d4ff', '#00ff88', '#ffaa00', '#ff3366', '#aa66ff'];

        if (seepageCompareChart) {
            seepageCompareChart.data.labels = labels;
            seepageCompareChart.data.datasets = [{
                label: '渗流量 (L/s)',
                data: result.items.map(item => (item.total_seepage_flow * 1000).toFixed(2)),
                backgroundColor: colors.slice(0, result.items.length),
                borderColor: colors.slice(0, result.items.length),
                borderWidth: 2
            }];
            seepageCompareChart.update();
        }

        if (pressureCompareChart) {
            pressureCompareChart.data.labels = labels;
            pressureCompareChart.data.datasets = [{
                label: '最大扬压力 (kPa)',
                data: result.items.map(item => item.max_pore_pressure.toFixed(1)),
                backgroundColor: colors.slice(0, result.items.length),
                borderColor: colors.slice(0, result.items.length),
                borderWidth: 2
            }];
            pressureCompareChart.update();
        }

        if (efficiencyCompareChart) {
            efficiencyCompareChart.data.labels = labels;
            efficiencyCompareChart.data.datasets = [{
                label: '防渗效率 (%)',
                data: result.items.map(item => item.anti_seepage_efficiency.toFixed(1)),
                backgroundColor: colors.slice(0, result.items.length),
                borderColor: colors.slice(0, result.items.length),
                borderWidth: 2
            }];
            efficiencyCompareChart.update();
        }

        cardsContainer.innerHTML = result.items.map((item, idx) => `
            <div class="comparison-card" style="border-left: 4px solid ${colors[idx]}">
                <h4>${item.dam_name}</h4>
                <div class="card-meta">${item.build_dynasty} · ${dam_presets_GetDamTypeLabel(item.dam_type)}</div>
                <div class="card-stats">
                    <div class="card-stat">
                        <span class="stat-label">几何尺寸</span>
                        <span class="stat-value">长${item.geometry.length}m × 高${item.geometry.height}m</span>
                    </div>
                    <div class="card-stat">
                        <span class="stat-label">渗流量</span>
                        <span class="stat-value ${item.total_seepage_flow * 1000 > 10 ? 'danger' : ''}">${(item.total_seepage_flow * 1000).toFixed(2)} L/s</span>
                    </div>
                    <div class="card-stat">
                        <span class="stat-label">单位渗流量</span>
                        <span class="stat-value">${(item.seepage_flow_per_meter * 1000).toFixed(3)} L/s/m</span>
                    </div>
                    <div class="card-stat">
                        <span class="stat-label">最大扬压力</span>
                        <span class="stat-value">${item.max_pore_pressure.toFixed(1)} kPa</span>
                    </div>
                    <div class="card-stat">
                        <span class="stat-label">出口梯度</span>
                        <span class="stat-value">${item.exit_gradient.toFixed(3)}</span>
                    </div>
                    <div class="card-stat">
                        <span class="stat-label">防渗效率</span>
                        <span class="stat-value highlight">${item.anti_seepage_efficiency.toFixed(1)}%</span>
                    </div>
                </div>
            </div>
        `).join('');
    }

    async function runCrossEraComparison() {
        SeepagePanel.showLoading('正在进行跨时代对比分析...', 15);

        try {
            const result = await SeepagePanel.apiPost('/compare/cross-era', {
                ancient_dam_key: 'tashan_weir',
                modern_dam_key: 'modern_gravity',
                upstream_wl: parseFloat(document.getElementById('cmpUpstream').value),
                downstream_wl: parseFloat(document.getElementById('cmpDownstream').value),
                scale_to_same_size: true
            });

            SeepagePanel.hideLoading();
            renderCrossEraResult(result);
        } catch (e) {
            SeepagePanel.hideLoading();
            SeepagePanel.showNotification('跨时代对比失败: ' + e.message, 'danger');
        }
    }

    function renderCrossEraResult(result) {
        const crossEraResult = document.getElementById('crossEraResult');
        const ancientInfo = document.getElementById('ancientDamInfo');
        const modernInfo = document.getElementById('modernDamInfo');
        const insightsEl = document.getElementById('crossEraInsights');

        crossEraResult.style.display = 'block';

        ancientInfo.innerHTML = renderDamInfoCard(result.ancient_dam, result.ancient_metrics);
        modernInfo.innerHTML = renderDamInfoCard(result.modern_dam, result.modern_metrics);

        insightsEl.innerHTML = result.insights.map(i => `<li>${i}</li>`).join('');

        const comp = result.comparison;
        const cmpHtml = `
            <div class="comparison-metrics">
                <h5>📊 技术进步量化对比</h5>
                <div class="metric-row">
                    <span>渗透系数降低</span>
                    <span class="highlight">${Math.abs(comp.permeability_reduction_pct).toFixed(1)}%</span>
                </div>
                <div class="metric-row">
                    <span>渗流量减少</span>
                    <span class="highlight">${comp.seepage_flow_reduction_pct.toFixed(1)}%</span>
                </div>
                <div class="metric-row">
                    <span>扬压力降低</span>
                    <span class="highlight">${comp.pore_pressure_reduction_pct.toFixed(1)}%</span>
                </div>
                <div class="metric-row">
                    <span>坝高提升</span>
                    <span>${(comp.height_ratio - 1) * 100.toFixed(0)}%</span>
                </div>
                <div class="metric-row">
                    <span>技术跨越</span>
                    <span class="highlight">${comp.technology_gap_years} 年</span>
                </div>
            </div>
        `;
        insightsEl.insertAdjacentHTML('afterend', cmpHtml);
    }

    function renderDamInfoCard(item, metrics) {
        return `
            <div class="info-section">
                <div class="info-row"><span>建造时间</span><span>${item.build_dynasty}</span></div>
                <div class="info-row"><span>材料</span><span>${item.geometry ? metrics?.material || '-' : '-'}</span></div>
                <div class="info-row"><span>尺寸</span><span>${item.geometry?.length}m × ${item.geometry?.height}m</span></div>
                <div class="info-row highlight"><span>渗流量</span><span>${(item.total_seepage_flow * 1000).toFixed(2)} L/s</span></div>
                <div class="info-row highlight"><span>最大扬压力</span><span>${item.max_pore_pressure.toFixed(1)} kPa</span></div>
                <div class="info-row"><span>防渗效率</span><span>${item.anti_seepage_efficiency.toFixed(1)}%</span></div>
            </div>
            ${metrics ? `
                <div class="info-section">
                    <div class="info-row"><span>防渗工艺</span><span>${metrics.anti_seepage_method || '-'}</span></div>
                    ${metrics.cultural_value ? `<div class="info-row cultural"><span>文化价值</span><span>${metrics.cultural_value}</span></div>` : ''}
                </div>
            ` : ''}
        `;
    }

    function dam_presets_GetDamTypeLabel(type) {
        const labels = { 'ancient_stone': '古代条石坝', 'modern_concrete': '现代混凝土重力坝' };
        return labels[type] || type;
    }

    return { init, runComparison };
})();

const AgingPrediction = (function() {
    let permeabilityChart = null;
    let agingFlowChart = null;
    let failureProbChart = null;
    let agingDegreeChart = null;
    let currentResult = null;

    async function init() {
        await loadDamOptions();
        bindEvents();
        initCharts();
    }

    async function loadDamOptions() {
        const select = document.getElementById('agingDamSelect');
        if (!select) return;

        try {
            const data = await SeepagePanel.apiGet('/dams');
            select.innerHTML = data.dams.map(dam => 
                `<option value="${dam.dam_key}">${dam.dam_name} (${dam.build_dynasty})</option>`
            ).join('');
        } catch (e) {
            console.error('Failed to load dam options:', e);
        }
    }

    function bindEvents() {
        const btnPredict = document.getElementById('btnPredictAging');
        if (btnPredict) {
            btnPredict.addEventListener('click', runPrediction);
        }

        const btnScenarios = document.getElementById('btnCompareScenarios');
        if (btnScenarios) {
            btnScenarios.addEventListener('click', runScenarioComparison);
        }
    }

    function initCharts() {
        const commonOptions = {
            responsive: true,
            maintainAspectRatio: false,
            plugins: { legend: { position: 'top' } },
            scales: { y: { beginAtZero: true } },
            animation: { duration: 800 }
        };

        const permCtx = document.getElementById('permeabilityChart');
        if (permCtx) {
            permeabilityChart = new Chart(permCtx, {
                type: 'line',
                data: { labels: [], datasets: [] },
                options: {
                    ...commonOptions,
                    scales: { y: { type: 'logarithmic' } }
                }
            });
        }

        const flowCtx = document.getElementById('agingFlowChart');
        if (flowCtx) {
            agingFlowChart = new Chart(flowCtx, {
                type: 'line',
                data: { labels: [], datasets: [] },
                options: commonOptions
            });
        }

        const failCtx = document.getElementById('failureProbChart');
        if (failCtx) {
            failureProbChart = new Chart(failCtx, {
                type: 'line',
                data: { labels: [], datasets: [] },
                options: {
                    ...commonOptions,
                    scales: { y: { max: 100 } }
                }
            });
        }

        const degCtx = document.getElementById('agingDegreeChart');
        if (degCtx) {
            agingDegreeChart = new Chart(degCtx, {
                type: 'line',
                data: { labels: [], datasets: [] },
                options: {
                    ...commonOptions,
                    scales: { y: { max: 100 } }
                }
            });
        }
    }

    async function runPrediction() {
        const req = {
            dam_key: document.getElementById('agingDamSelect').value,
            prediction_years: parseInt(document.getElementById('agingYears').value),
            time_step_years: parseInt(document.getElementById('agingStep').value),
            initial_permeability: parseFloat(document.getElementById('agingInitialK').value),
            consider_climate: document.getElementById('agingConsiderClimate').checked,
            consider_maintenance: document.getElementById('agingConsiderMaintenance').checked,
            maintenance_frequency: document.getElementById('agingMaintenance').value
        };

        SeepagePanel.showLoading('正在预测未来老化趋势...', 20);

        try {
            currentResult = await SeepagePanel.apiPost('/aging/predict', req);
            SeepagePanel.hideLoading();
            renderPredictionResult(currentResult);
        } catch (e) {
            SeepagePanel.hideLoading();
            SeepagePanel.showNotification('预测失败: ' + e.message, 'danger');
        }
    }

    function renderPredictionResult(result) {
        const summaryEl = document.getElementById('agingSummary');
        const summaryContent = document.getElementById('agingSummaryContent');

        summaryEl.style.display = 'block';
        summaryContent.innerHTML = `
            <div class="summary-grid">
                <div class="summary-item">
                    <span class="summary-label">坝名</span>
                    <span class="summary-value">${result.dam_name}</span>
                </div>
                <div class="summary-item">
                    <span class="summary-label">当前坝龄</span>
                    <span class="summary-value">${result.initial_age} 年</span>
                </div>
                <div class="summary-item">
                    <span class="summary-label">预测年限</span>
                    <span class="summary-value">未来 ${result.prediction_years} 年</span>
                </div>
                ${result.critical_year ? `
                    <div class="summary-item danger">
                        <span class="summary-label">⚠️ 临界年份</span>
                        <span class="summary-value">${result.critical_year} 年</span>
                    </div>
                ` : ''}
                <div class="summary-item">
                    <span class="summary-label">年老化速率</span>
                    <span class="summary-value">${(result.aging_rate * 1e12).toFixed(2)} ×10⁻¹² m/s/年</span>
                </div>
            </div>
            <p class="summary-text">${result.summary}</p>
        `;

        const labels = result.data_points.map(dp => dp.year);

        const datasets = [
            {
                chart: permeabilityChart,
                label: '渗透系数 (m/s)',
                data: result.data_points.map(dp => dp.permeability),
                color: '#ff3366',
                logScale: true
            },
            {
                chart: agingFlowChart,
                label: '渗流量 (L/s)',
                data: result.data_points.map(dp => (dp.seepage_flow * 1000).toFixed(3)),
                color: '#00d4ff'
            },
            {
                chart: failureProbChart,
                label: '失效概率 (%)',
                data: result.data_points.map(dp => (dp.failure_probability * 100).toFixed(1)),
                color: '#ffaa00',
                warningZone: 50,
                dangerZone: 80
            },
            {
                chart: agingDegreeChart,
                label: '老化程度 (%)',
                data: result.data_points.map(dp => dp.degree_of_aging.toFixed(1)),
                color: '#aa66ff',
                warningZone: 30,
                dangerZone: 60
            }
        ];

        datasets.forEach(ds => {
            if (ds.chart) {
                ds.chart.data.labels = labels;
                ds.chart.data.datasets = [{
                    label: ds.label,
                    data: ds.data,
                    borderColor: ds.color,
                    backgroundColor: ds.color + '20',
                    fill: true,
                    tension: 0.3,
                    pointRadius: 4,
                    pointHoverRadius: 6
                }];
                ds.chart.update();
            }
        });

        const tableBody = document.getElementById('agingTableBody');
        tableBody.innerHTML = result.data_points.map(dp => `
            <tr class="${dp.failure_probability > 0.5 ? 'danger' : dp.failure_probability > 0.2 ? 'warning' : ''}">
                <td>${dp.year}</td>
                <td>${dp.age_years}</td>
                <td>${dp.permeability.toExponential(2)}</td>
                <td>${(dp.seepage_flow * 1000).toFixed(3)}</td>
                <td>${dp.max_pore_pressure.toFixed(1)}</td>
                <td>${dp.degree_of_aging.toFixed(1)}%</td>
                <td>
                    <span class="prob-badge ${dp.failure_probability > 0.5 ? 'danger' : dp.failure_probability > 0.2 ? 'warning' : 'success'}">
                        ${(dp.failure_probability * 100).toFixed(0)}%
                    </span>
                </td>
                <td>${dp.recommended_action}</td>
            </tr>
        `).join('');

        const recList = document.getElementById('recommendationList');
        recList.innerHTML = result.recommendations.map(r => `<li>${r}</li>`).join('');
        document.getElementById('agingRecommendations').style.display = 'block';
    }

    async function runScenarioComparison() {
        const req = {
            dam_key: document.getElementById('agingDamSelect').value,
            prediction_years: parseInt(document.getElementById('agingYears').value),
            time_step_years: parseInt(document.getElementById('agingStep').value),
            initial_permeability: parseFloat(document.getElementById('agingInitialK').value),
            consider_climate: document.getElementById('agingConsiderClimate').checked,
            consider_maintenance: true,
            maintenance_frequency: 'medium'
        };

        SeepagePanel.showLoading('正在对比多情景预测...', 25);

        try {
            const result = await SeepagePanel.apiPost('/aging/scenarios', req);
            SeepagePanel.hideLoading();
            renderScenarioComparison(result.scenarios);
        } catch (e) {
            SeepagePanel.hideLoading();
            SeepagePanel.showNotification('情景对比失败: ' + e.message, 'danger');
        }
    }

    function renderScenarioComparison(scenarios) {
        const labels = scenarios.baseline?.data_points?.map(dp => dp.year) || [];
        const scenarioColors = {
            baseline: { name: '基准情景', color: '#00d4ff' },
            high_maintenance: { name: '高维护', color: '#00ff88' },
            no_maintenance: { name: '无维护', color: '#ff3366' },
            with_climate: { name: '考虑气候', color: '#ffaa00' }
        };

        [permeabilityChart, agingFlowChart, failureProbChart, agingDegreeChart].forEach(chart => {
            if (!chart) return;
            chart.data.datasets = Object.keys(scenarios).map(key => {
                const sc = scenarios[key];
                const meta = scenarioColors[key] || { name: key, color: '#aa66ff' };
                let data = [];
                if (chart === permeabilityChart) data = sc.data_points.map(dp => dp.permeability);
                else if (chart === agingFlowChart) data = sc.data_points.map(dp => (dp.seepage_flow * 1000).toFixed(3));
                else if (chart === failureProbChart) data = sc.data_points.map(dp => (dp.failure_probability * 100).toFixed(1));
                else if (chart === agingDegreeChart) data = sc.data_points.map(dp => dp.degree_of_aging.toFixed(1));

                return {
                    label: meta.name,
                    data: data,
                    borderColor: meta.color,
                    backgroundColor: meta.color + '15',
                    fill: false,
                    tension: 0.3,
                    borderWidth: 2
                };
            });
            chart.update();
        });
    }

    return { init };
})();

const VirtualTour = (function() {
    let currentDamKey = 'tashan_weir';
    let currentSceneIndex = 0;
    let scenes = [];
    let dam3dViewer = null;
    let interactiveResult = null;

    async function init() {
        await loadScenes();
        bindEvents();
        renderSceneList();
    }

    async function loadScenes() {
        try {
            const data = await SeepagePanel.apiGet(`/dams/${currentDamKey}/virtual-tour`);
            scenes = data.scenes || [];
            if (scenes.length > 0) {
                showScene(0);
            }
        } catch (e) {
            console.error('Failed to load virtual tour scenes:', e);
        }
    }

    function bindEvents() {
        const prevBtn = document.getElementById('prevScene');
        if (prevBtn) {
            prevBtn.addEventListener('click', () => showScene(currentSceneIndex - 1));
        }

        const nextBtn = document.getElementById('nextScene');
        if (nextBtn) {
            nextBtn.addEventListener('click', () => showScene(currentSceneIndex + 1));
        }

        const tourUpstream = document.getElementById('tourUpstream');
        if (tourUpstream) {
            tourUpstream.addEventListener('input', (e) => {
                document.getElementById('tourUpstreamValue').textContent = e.target.value + ' m';
            });
        }

        const tourDownstream = document.getElementById('tourDownstream');
        if (tourDownstream) {
            tourDownstream.addEventListener('input', (e) => {
                document.getElementById('tourDownstreamValue').textContent = e.target.value + ' m';
            });
        }

        const btnApply = document.getElementById('btnApplyWaterLevel');
        if (btnApply) {
            btnApply.addEventListener('click', applyWaterLevel);
        }

        const showSeepage = document.getElementById('tourShowSeepage');
        if (showSeepage) {
            showSeepage.addEventListener('change', (e) => {
                if (dam3dViewer) {
                    dam3dViewer.setStreamlinesVisible(e.target.checked);
                }
            });
        }

        const showSensors = document.getElementById('tourShowSensors');
        if (showSensors) {
            showSensors.addEventListener('change', (e) => {
                if (dam3dViewer) {
                    dam3dViewer.setSensorsVisible(e.target.checked);
                }
            });
        }

        const vizMode = document.getElementById('tourVizMode');
        if (vizMode) {
            vizMode.addEventListener('change', (e) => {
                if (dam3dViewer) {
                    const mode = e.target.value;
                    dam3dViewer.setPressureCloudVisible(mode === 'both' || mode === 'pressure');
                    dam3dViewer.setStreamlinesVisible(mode === 'both' || mode === 'streamline');
                }
            });
        }

        const highlight = document.getElementById('tourHighlight');
        if (highlight) {
            highlight.addEventListener('change', (e) => {
                if (dam3dViewer) {
                    dam3dViewer.highlightArea(e.target.value);
                }
            });
        }
    }

    function renderSceneList() {
        const container = document.getElementById('sceneList');
        if (!container) return;

        container.innerHTML = scenes.map((scene, idx) => `
            <div class="scene-item ${idx === currentSceneIndex ? 'active' : ''}" data-index="${idx}">
                <div class="scene-number">${idx + 1}</div>
                <div class="scene-info">
                    <div class="scene-name">${scene.scene_name}</div>
                    <div class="scene-desc">${scene.description}</div>
                </div>
            </div>
        `).join('');

        container.querySelectorAll('.scene-item').forEach(item => {
            item.addEventListener('click', () => {
                showScene(parseInt(item.dataset.index));
            });
        });
    }

    function showScene(index) {
        if (index < 0 || index >= scenes.length) return;

        currentSceneIndex = index;
        const scene = scenes[index];

        document.getElementById('sceneIndicator').textContent = `${index + 1} / ${scenes.length}`;
        document.getElementById('sceneTitle').textContent = scene.scene_name;
        document.getElementById('sceneNarrative').textContent = scene.narrative;

        if (dam3dViewer && scene.camera_position) {
            dam3dViewer.setCamera(
                scene.camera_position.x,
                scene.camera_position.y,
                scene.camera_position.z,
                scene.camera_target.x,
                scene.camera_target.y,
                scene.camera_target.z
            );
        }

        renderHotspots(scene.hotspots || []);
        renderSceneList();
    }

    function renderHotspots(hotspots) {
        const container = document.getElementById('hotspotList');
        if (!container) return;

        if (hotspots.length === 0) {
            container.innerHTML = '<p class="text-center">当前场景无热点</p>';
            return;
        }

        container.innerHTML = hotspots.map(hs => `
            <div class="hotspot-item">
                <h5>${hs.title}</h5>
                <p>${hs.description}</p>
            </div>
        `).join('');
    }

    async function applyWaterLevel() {
        const upWL = parseFloat(document.getElementById('tourUpstream').value);
        const downWL = parseFloat(document.getElementById('tourDownstream').value);
        const highlight = document.getElementById('tourHighlight').value;
        const vizMode = document.getElementById('tourVizMode').value;

        SeepagePanel.showLoading('正在计算渗流场变化...', 15);

        try {
            interactiveResult = await SeepagePanel.apiPost('/interactive/adjust', {
                dam_key: currentDamKey,
                upstream_wl: upWL,
                downstream_wl: downWL,
                highlight_area: highlight,
                visualization_mode: vizMode
            });

            SeepagePanel.hideLoading();
            renderInteractiveResult(interactiveResult);
        } catch (e) {
            SeepagePanel.hideLoading();
            SeepagePanel.showNotification('计算失败: ' + e.message, 'danger');
        }
    }

    function renderInteractiveResult(result) {
        const statsPanel = document.getElementById('tourStats');
        const riskLevel = document.getElementById('tourRiskLevel');
        const explanation = document.getElementById('tourExplanation');

        statsPanel.style.display = 'block';

        riskLevel.className = `risk-level ${result.risk_level}`;
        const riskLabels = { low: '✅ 安全', medium: '⚠️ 警戒', high: '🔴 危险', critical: '🚨 严重' };
        riskLevel.textContent = riskLabels[result.risk_level] || result.risk_level;

        document.getElementById('tourStatFlow').textContent = 
            result.key_metrics.total_seepage_flow_lps.toFixed(2) + ' L/s';
        document.getElementById('tourStatPressure').textContent = 
            result.key_metrics.max_pore_pressure_kpa.toFixed(1) + ' kPa';
        document.getElementById('tourStatHead').textContent = 
            result.key_metrics.water_head_difference_m.toFixed(1) + ' m';

        explanation.textContent = result.explanation;
        SeepagePanel.showNotification(result.water_level_change, 
            result.risk_level === 'low' ? 'success' : 
            result.risk_level === 'medium' ? 'warning' : 'danger');

        if (result.simulation && result.grids && dam3dViewer) {
            dam3dViewer.updateSimulationData(result.grids, result.simulation);
            dam3dViewer.updateWaterLevels(
                result.key_metrics.upstream_wl_m,
                result.key_metrics.downstream_wl_m
            );
        }
    }

    function setDamViewer(viewer) {
        dam3dViewer = viewer;
    }

    function getCurrentDamKey() {
        return currentDamKey;
    }

    return { init, setDamViewer, getCurrentDamKey, applyWaterLevel };
})();

document.addEventListener('DOMContentLoaded', () => {
    if (typeof SeepagePanel !== 'undefined') {
        const originalInit = SeepagePanel.init;
        SeepagePanel.init = function() {
            if (originalInit) originalInit.call(this);
            
            setTimeout(() => {
                DamComparison.init().catch(console.error);
                AgingPrediction.init().catch(console.error);
                VirtualTour.init().catch(console.error);
            }, 500);
        };
    }

    const originalTabHandler = function(tabName) {
        document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
        document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
        event.target.classList.add('active');
        document.getElementById('tab-' + tabName).classList.add('active');

        if (tabName === 'tour' && !window.tourViewerInitialized) {
            setTimeout(() => {
                if (typeof TuoshanDam3D !== 'undefined') {
                    const container = document.getElementById('tourCanvasContainer');
                    if (container && !window.tourViewer) {
                        window.tourViewer = new TuoshanDam3D(container);
                        VirtualTour.setDamViewer(window.tourViewer);
                        window.tourViewerInitialized = true;
                    }
                }
            }, 100);
        }
    };

    setTimeout(() => {
        document.querySelectorAll('.tab-btn').forEach(btn => {
            const tabName = btn.dataset.tab;
            if (tabName) {
                btn.onclick = function() {
                    document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
                    document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
                    btn.classList.add('active');
                    document.getElementById('tab-' + tabName).classList.add('active');

                    if (tabName === 'tour' && !window.tourViewerInitialized) {
                        setTimeout(() => {
                            if (typeof TuoshanDam3D !== 'undefined') {
                                const container = document.getElementById('tourCanvasContainer');
                                if (container && !window.tourViewer) {
                                    window.tourViewer = new TuoshanDam3D(container);
                                    VirtualTour.setDamViewer(window.tourViewer);
                                    window.tourViewerInitialized = true;
                                }
                            }
                        }, 100);
                    }
                };
            }
        });
    }, 100);
});
