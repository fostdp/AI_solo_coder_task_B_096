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

    function dam_presets_GetDamTypeLabel(type) {
        const labels = { 'ancient_stone': '古代条石坝', 'modern_concrete': '现代混凝土重力坝' };
        return labels[type] || type;
    }

    return { init, runComparison };
})();
