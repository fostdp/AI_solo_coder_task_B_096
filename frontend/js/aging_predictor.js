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
