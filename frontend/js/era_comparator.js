const EraComparator = (function() {
    async function init() {
        bindEvents();
    }

    function bindEvents() {
        const btnCrossEra = document.getElementById('btnCrossEra');
        if (btnCrossEra) {
            btnCrossEra.addEventListener('click', runCrossEraComparison);
        }
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

    return { init };
})();
