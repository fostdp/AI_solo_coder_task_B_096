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

    const DEBOUNCE_MS = 300;
    const COARSE_STEP = 0.10;
    const FINE_STEP = 0.01;
    let debounceTimer = null;
    let lastUpstreamValue = null;
    let lastDownstreamValue = null;
    let lastRequestTime = 0;
    const MIN_REQUEST_INTERVAL = 200;

    function debounce(fn, ms) {
        return function (...args) {
            if (debounceTimer) clearTimeout(debounceTimer);
            debounceTimer = setTimeout(() => fn.apply(this, args), ms);
        };
    }

    function throttle(fn, ms) {
        let last = 0;
        return function (...args) {
            const now = Date.now();
            if (now - last >= ms) {
                last = now;
                return fn.apply(this, args);
            }
        };
    }

    const throttledApplyWaterLevel = throttle(applyWaterLevel, MIN_REQUEST_INTERVAL);

    function hasSignificantChange(upWL, downWL, threshold = 0.02) {
        if (lastUpstreamValue === null || lastDownstreamValue === null) {
            return true;
        }
        const dUp = Math.abs(upWL - lastUpstreamValue);
        const dDown = Math.abs(downWL - lastDownstreamValue);
        return dUp >= threshold || dDown >= threshold;
    }

    const debouncedApplyFromSlider = debounce(function () {
        const upWL = parseFloat(document.getElementById('tourUpstream').value);
        const downWL = parseFloat(document.getElementById('tourDownstream').value);

        if (!hasSignificantChange(upWL, downWL, FINE_STEP)) {
            return;
        }

        lastUpstreamValue = upWL;
        lastDownstreamValue = downWL;
        throttledApplyWaterLevel();
    }, DEBOUNCE_MS);

        const tourUpstream = document.getElementById('tourUpstream');
        if (tourUpstream) {
            tourUpstream.addEventListener('input', (e) => {
                document.getElementById('tourUpstreamValue').textContent = e.target.value + ' m';
                debouncedApplyFromSlider();
            });
            tourUpstream.addEventListener('keydown', (e) => {
                const step = e.shiftKey ? FINE_STEP : COARSE_STEP;
                const el = e.target;
                let val = parseFloat(el.value);
                if (e.key === 'ArrowUp') {
                    e.preventDefault();
                    val = Math.min(parseFloat(el.max) || Infinity, val + step);
                    el.value = val.toFixed(2);
                    document.getElementById('tourUpstreamValue').textContent = val.toFixed(2) + ' m';
                    debouncedApplyFromSlider();
                } else if (e.key === 'ArrowDown') {
                    e.preventDefault();
                    val = Math.max(parseFloat(el.min) || 0, val - step);
                    el.value = val.toFixed(2);
                    document.getElementById('tourUpstreamValue').textContent = val.toFixed(2) + ' m';
                    debouncedApplyFromSlider();
                }
            });
            tourUpstream.setAttribute('step', COARSE_STEP.toString());
        }

        const tourDownstream = document.getElementById('tourDownstream');
        if (tourDownstream) {
            tourDownstream.addEventListener('input', (e) => {
                document.getElementById('tourDownstreamValue').textContent = e.target.value + ' m';
                debouncedApplyFromSlider();
            });
            tourDownstream.addEventListener('keydown', (e) => {
                const step = e.shiftKey ? FINE_STEP : COARSE_STEP;
                const el = e.target;
                let val = parseFloat(el.value);
                if (e.key === 'ArrowUp') {
                    e.preventDefault();
                    val = Math.min(parseFloat(el.max) || Infinity, val + step);
                    el.value = val.toFixed(2);
                    document.getElementById('tourDownstreamValue').textContent = val.toFixed(2) + ' m';
                    debouncedApplyFromSlider();
                } else if (e.key === 'ArrowDown') {
                    e.preventDefault();
                    val = Math.max(parseFloat(el.min) || 0, val - step);
                    el.value = val.toFixed(2);
                    document.getElementById('tourDownstreamValue').textContent = val.toFixed(2) + ' m';
                    debouncedApplyFromSlider();
                }
            });
            tourDownstream.setAttribute('step', COARSE_STEP.toString());
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

        lastUpstreamValue = upWL;
        lastDownstreamValue = downWL;
        lastRequestTime = Date.now();

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
                EraComparator.init().catch(console.error);
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
