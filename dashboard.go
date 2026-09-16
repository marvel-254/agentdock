package main

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OrionOS — Control Plane</title>
    <script src="https://unpkg.com/htmx.org@1.9.12"></script>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Fira+Code:wght@400;500;600;700&family=Fira+Sans:wght@300;400;500;600;700&family=Caveat:wght@400;700&display=swap" rel="stylesheet">
    <script src="https://unpkg.com/dagre@0.8.5/dist/dagre.min.js"></script>
    <style>
        :root {
            --bg: #0f0f14;
            --surface: #16161e;
            --border: #1e1e2e;
            --primary: #7aa2f7;
            --success: #9ece6a;
            --warning: #e0af68;
            --danger: #f7768e;
            --muted: #565f89;
            --text: #e9e9f0;
            --secondary: #a9b1d6;
            --annotation: #bb9af7;
            --nav-width: 240px;
            --header-height: 64px;
            --panel-width: 400px;
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'Fira Sans', system-ui, sans-serif;
            background: var(--bg);
            color: var(--text);
            line-height: 1.5;
            -webkit-font-smoothing: antialiased;
            overflow: hidden;
        }
        .monospace { font-family: 'Fira Code', monospace; }
        .handwritten { font-family: 'Caveat', cursive; color: var(--annotation); }

        /* Layout */
        .app-container {
            display: grid;
            grid-template-columns: var(--nav-width) 1fr;
            height: 100vh;
            width: 100vw;
        }

        /* Sidebar */
        .sidebar {
            background: var(--surface);
            border-right: 1px solid var(--border);
            display: flex;
            flex-direction: column;
            z-index: 100;
        }
        .sidebar-header { padding: 1.5rem; border-bottom: 1px solid var(--border); }
        .logo { font-family: 'Fira Code', monospace; font-weight: 700; font-size: 1.25rem; color: var(--text); }
        .logo-subtitle { font-size: 0.65rem; color: var(--muted); letter-spacing: 0.05em; margin-top: 0.25rem; }
        
        .nav-list { list-style: none; padding: 1rem 0; flex: 1; overflow-y: auto; }
        .nav-item {
            padding: 0.75rem 1.5rem;
            display: flex;
            align-items: center;
            gap: 0.75rem;
            color: var(--secondary);
            text-decoration: none;
            cursor: pointer;
            transition: all 0.2s ease;
            font-size: 0.9rem;
        }
        .nav-item:hover { color: var(--text); background: rgba(122, 162, 247, 0.05); }
        .nav-item.active { color: var(--primary); background: rgba(122, 162, 247, 0.08); border-right: 2px solid var(--primary); }

        /* Main Workspace */
        .main {
            display: flex;
            flex-direction: column;
            position: relative;
            overflow: hidden;
        }
        .top-bar {
            height: var(--header-height);
            border-bottom: 1px solid var(--border);
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 0 1.5rem;
            background: var(--bg);
            z-index: 90;
        }

        /* View Containers */
        .view-port { flex: 1; position: relative; overflow: auto; }
        .view-container { display: none; min-height: 100%; width: 100%; position: absolute; top: 0; left: 0; }
        .view-container.active { display: block; }

        /* Graph View */
        #graph-toolbar {
            position: sticky; top: 0; z-index: 50;
            display: flex; gap: 0.5rem; align-items: center;
            padding: 0.75rem 1rem; background: var(--surface);
            border-bottom: 1px solid var(--border);
            flex-wrap: wrap;
        }
        #graph-toolbar input, #graph-toolbar select {
            background: var(--bg); border: 1px solid var(--border);
            color: var(--text); padding: 0.4rem 0.75rem;
            border-radius: 4px; font-size: 0.8rem; font-family: 'Fira Sans', sans-serif;
        }
        #graph-toolbar input:focus, #graph-toolbar select:focus {
            outline: none; border-color: var(--primary);
        }
        #graph-toolbar .btn {
            background: var(--bg); border: 1px solid var(--border);
            color: var(--text); padding: 0.4rem 0.75rem;
            border-radius: 4px; font-size: 0.75rem; cursor: pointer;
            font-family: 'Fira Code', monospace;
            transition: all 0.15s;
        }
        #graph-toolbar .btn:hover { background: var(--primary); border-color: var(--primary); color: var(--bg); }
        #graph-toolbar .btn.active { background: var(--primary); border-color: var(--primary); color: var(--bg); }
        #graph-toolbar .spacer { flex: 1; }
        
        #graph-workspace {
            width: 3000px; height: 3000px;
            background-image: radial-gradient(var(--border) 1px, transparent 1px);
            background-size: 40px 40px;
            position: relative;
            transform-origin: 0 0;
            transition: transform 0.1s linear;
        }
        #graph-svg {
            position: absolute; top: 0; left: 0; width: 100%; height: 100%;
            pointer-events: none;
        }
        .node {
            position: absolute;
            width: 200px;
            background: var(--surface);
            border: 1px solid var(--border);
            border-radius: 4px;
            padding: 1rem;
            z-index: 10;
            transition: border-color 0.2s, transform 0.1s;
            cursor: pointer;
            user-select: none;
        }
        .node:hover { border-color: var(--primary); transform: translateY(-2px); }
        .node.selected { border-color: var(--primary); box-shadow: 0 0 15px rgba(122, 162, 247, 0.2); }
        .node-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 0.5rem; }
        .node-status { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
        .node-icon { width: 24px; height: 24px; margin-right: 0.5rem; flex-shrink: 0; }
        .node-title { display: flex; align-items: center; }

        /* Office View */
        #office-workspace {
            display: flex; justify-content: center; align-items: center;
            width: 100%; height: 100%;
            perspective: 1200px;
            background: radial-gradient(circle at center, #1a1b26 0%, var(--bg) 100%);
        }
        #office-toolbar {
            position: sticky; top: 0; z-index: 50;
            display: flex; gap: 0.5rem; align-items: center;
            padding: 0.75rem 1rem; background: var(--surface);
            border-bottom: 1px solid var(--border);
            flex-wrap: wrap;
        }
        #office-toolbar select {
            background: var(--bg); border: 1px solid var(--border);
            color: var(--text); padding: 0.4rem 0.75rem;
            border-radius: 4px; font-size: 0.8rem; font-family: 'Fira Sans', sans-serif;
        }
        #office-toolbar .btn {
            background: var(--bg); border: 1px solid var(--border);
            color: var(--text); padding: 0.4rem 0.75rem;
            border-radius: 4px; font-size: 0.75rem; cursor: pointer;
            font-family: 'Fira Code', monospace;
            transition: all 0.15s;
        }
        #office-toolbar .btn:hover { background: var(--primary); border-color: var(--primary); color: var(--bg); }
        #office-toolbar .spacer { flex: 1; }
        .isometric-grid {
            width: 800px; height: 800px;
            transform: rotateX(60deg) rotateZ(-45deg);
            background-image: 
                linear-gradient(var(--border) 1px, transparent 1px),
                linear-gradient(90deg, var(--border) 1px, transparent 1px);
            background-size: 100px 100px;
            position: relative;
            transform-style: preserve-3d;
        }
        .station {
            position: absolute;
            width: 100px; height: 100px;
            background: rgba(122, 162, 247, 0.1);
            border: 1px solid var(--primary);
            transform-style: preserve-3d;
            cursor: pointer;
            transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
        }
        .station::before {
            content: ""; position: absolute; inset: 0; background: var(--primary); opacity: 0.1;
        }
        .station-avatar {
            position: absolute; bottom: 100%; left: 50%;
            width: 50px; height: 70px;
            background: var(--primary);
            transform: rotateZ(45deg) rotateX(-60deg) translateX(-50%);
            transform-origin: bottom center;
            border-radius: 6px 6px 0 0;
            box-shadow: 0 8px 20px rgba(0,0,0,0.5);
            display: flex; flex-direction: column; align-items: center; justify-content: center;
            padding-top: 8px;
        }
        .station.working .station-avatar { background: var(--success); animation: pulse 2s infinite; }
        .station.active .station-avatar { background: var(--primary); }
        .station.waiting .station-avatar { background: var(--muted); }
        .station .avatar-icon { width: 28px; height: 28px; color: rgba(0,0,0,0.6); }
        .station-label {
            position: absolute; top: 110%; left: 50%;
            width: 120px; margin-left: -60px;
            text-align: center; color: var(--secondary); font-size: 0.65rem; font-weight: 600;
            transform: rotateX(-60deg) rotateZ(45deg);
            white-space: nowrap;
        }
        .station-status {
            position: absolute; top: -8px; right: -8px;
            width: 14px; height: 14px; border-radius: 50%;
            border: 2px solid var(--bg);
        }
        .station.working .station-status { background: var(--success); animation: pulse 2s infinite; }
        .station.active .station-status { background: var(--primary); }
        .station.waiting .station-status { background: var(--muted); }
        @keyframes pulse { 0% { opacity: 1; } 50% { opacity: 0.7; } 100% { opacity: 1; } }
        .console-area {
            position: absolute; bottom: 20px; left: 50%; transform: translateX(-50%) rotateX(-60deg) rotateZ(45deg);
            width: 400px; padding: 1rem; background: var(--surface);
            border: 1px solid var(--border); border-radius: 4px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.3);
        }
        .console-title { font-size: 0.7rem; color: var(--muted); text-transform: uppercase; margin-bottom: 0.5rem; font-family: 'Fira Code', monospace; }
        .console-line { font-size: 0.65rem; color: var(--text); font-family: 'Fira Code', monospace; margin: 0.25rem 0; }
        .office-list { display: none; padding: 1rem; }
        .office-list .station-card { 
            display: flex; align-items: center; gap: 1rem; padding: 1rem; 
            background: var(--surface); border: 1px solid var(--border); 
            border-radius: 4px; margin-bottom: 0.5rem; cursor: pointer;
            transition: all 0.2s;
        }
        .office-list .station-card:hover { border-color: var(--primary); }
        .office-list .station-avatar-sm { width: 40px; height: 40px; border-radius: 4px; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
        .office-list .station-info { flex: 1; }
        .office-list .station-name { font-weight: 600; margin-bottom: 0.25rem; }
        .office-list .station-meta { font-size: 0.75rem; color: var(--muted); font-family: 'Fira Code', monospace; }

        /* Inspector Stack */
        .inspector-stack {
            position: fixed; top: 0; right: 0; bottom: 0;
            display: flex; align-items: flex-start;
            pointer-events: none; z-index: 200;
        }
        .panel {
            width: var(--panel-width);
            height: 100vh;
            background: var(--surface);
            border-left: 1px solid var(--border);
            box-shadow: -10px 0 30px rgba(0,0,0,0.5);
            pointer-events: auto;
            display: flex; flex-direction: column;
            transform: translateX(100%);
            transition: transform 0.3s cubic-bezier(0.16, 1, 0.3, 1);
        }
        .panel.active { transform: translateX(0); }
        .panel-header { padding: 1.5rem; border-bottom: 1px solid var(--border); display: flex; justify-content: space-between; align-items: center; }
        .panel-content { padding: 1.5rem; flex: 1; overflow-y: auto; }

        /* Annotations */
        .annotation-layer {
            position: absolute; inset: 0; pointer-events: none; z-index: 50;
        }
        .note {
            position: absolute;
            font-family: 'Caveat', cursive;
            color: var(--annotation);
            font-size: 1.1rem;
            white-space: nowrap;
            transform: rotate(-2deg);
        }

        /* Responsive */
        .mobile-nav { display: none; position: fixed; bottom: 0; left: 0; right: 0; height: 60px; background: var(--surface); border-top: 1px solid var(--border); z-index: 100; grid-template-columns: repeat(4, 1fr); }
        @media (max-width: 768px) {
            .app-container { grid-template-columns: 1fr; }
            .sidebar { display: none; }
            .mobile-nav { display: grid; }
            .panel { width: 100vw; }
        }

        /* Generic UI */
        .badge { padding: 0.2rem 0.5rem; border-radius: 4px; font-size: 0.7rem; font-weight: 600; text-transform: uppercase; font-family: 'Fira Code', monospace; }
        .badge.working { background: rgba(158, 206, 106, 0.1); color: var(--success); }
        .badge.active { background: rgba(122, 162, 247, 0.1); color: var(--primary); }
        .card { background: var(--surface); border: 1px solid var(--border); border-radius: 4px; padding: 1.25rem; margin-bottom: 1rem; }
    </style>
</head>
<body>
    <div class="app-container">
        <aside class="sidebar">
            <div class="sidebar-header">
                <div class="logo">ORIONOS</div>
                <div class="logo-subtitle">LOCAL AGENT CONTROL PLANE</div>
            </div>
            <nav class="nav-list">
                <a class="nav-item" onclick="engine.switchView('overview')">Overview</a>
                <a class="nav-item" onclick="engine.switchView('agents')">Agents</a>
                <a class="nav-item" onclick="engine.switchView('graph')">Graph</a>
                <a class="nav-item" onclick="engine.switchView('office')">Office</a>
                <a class="nav-item" onclick="engine.switchView('activity')">Activity</a>
                <a class="nav-item" onclick="engine.switchView('alerts')">Alerts</a>
            </nav>
            <div class="sidebar-footer" style="padding: 1.5rem; border-top: 1px solid var(--border); font-size: 0.8rem;">
                <div style="color: var(--muted); margin-bottom: 0.5rem; text-transform: uppercase;">System</div>
                <div id="sys-mini-stats" class="monospace">
                    CPU <span id="mini-cpu">0%</span><br>
                    MEM <span id="mini-mem">0MB</span>
                </div>
            </div>
        </aside>

        <main class="main">
            <header class="top-bar">
                <div id="view-title" class="monospace" style="text-transform: uppercase; font-size: 0.8rem; letter-spacing: 0.1em;">OVERVIEW</div>
                <div style="display: flex; gap: 1.5rem; font-family: 'Fira Code', monospace; font-size: 0.8rem;">
                    <div id="status-tag" style="color: var(--success);">● LIVE</div>
                    <div id="clock" style="color: var(--muted);">00:00:00</div>
                </div>
            </header>

            <div class="view-port">
                <!-- Overview -->
                <div id="view-overview" class="view-container active" style="padding: 2rem; max-width: 1200px; margin: 0 auto;">
                    <div style="display: grid; grid-template-columns: repeat(4, 1fr); gap: 1rem; margin-bottom: 2rem;">
                        <div class="card">
                            <div style="font-size: 0.7rem; color: var(--muted); text-transform: uppercase;">Total Agents</div>
                            <div id="total-agents-val" style="font-size: 2rem; font-weight: 700;">0</div>
                        </div>
                        <div class="card">
                            <div style="font-size: 0.7rem; color: var(--muted); text-transform: uppercase;">Active Load</div>
                            <div id="active-cpu-val" style="font-size: 2rem; font-weight: 700; color: var(--primary);">0%</div>
                        </div>
                        <div class="card">
                            <div style="font-size: 0.7rem; color: var(--muted); text-transform: uppercase;">Memory Usage</div>
                            <div id="active-ram-val" style="font-size: 2rem; font-weight: 700;">0MB</div>
                        </div>
                        <div class="card">
                            <div style="font-size: 0.7rem; color: var(--muted); text-transform: uppercase;">Notifications</div>
                            <div id="ntfy-status-val" style="font-size: 2rem; font-weight: 700;">...</div>
                        </div>
                    </div>
                    
                    <div style="display: grid; grid-template-columns: 2fr 1fr; gap: 2rem;">
                        <div>
                            <h3 style="font-size: 0.8rem; color: var(--muted); margin-bottom: 1rem; text-transform: uppercase;">Recent Events</h3>
                            <div id="overview-events" class="card" style="padding: 0;"></div>
                        </div>
                        <div>
                            <h3 style="font-size: 0.8rem; color: var(--muted); margin-bottom: 1rem; text-transform: uppercase;">Attention</h3>
                            <div id="overview-alerts"></div>
                        </div>
                    </div>
                </div>

                <!-- Graph -->
                <div id="view-graph" class="view-container" style="display: flex; flex-direction: column; height: 100%;">
                    <div id="graph-toolbar">
                        <input type="search" id="graph-search" placeholder="Search nodes..." style="width: 200px;">
                        <select id="graph-filter-type">
                            <option value="">All Types</option>
                            <option value="core">Core</option>
                            <option value="agent">Agent</option>
                        </select>
                        <select id="graph-filter-status">
                            <option value="">All Status</option>
                            <option value="working">Working</option>
                            <option value="active">Active</option>
                            <option value="waiting">Waiting</option>
                        </select>
                        <div class="spacer"></div>
                        <button class="btn" id="btn-zoom-in" title="Zoom In (+)">+</button>
                        <button class="btn" id="btn-zoom-out" title="Zoom Out (-)">-</button>
                        <button class="btn" id="btn-zoom-fit" title="Fit to Screen (0)">0</button>
                        <button class="btn" id="btn-layout-reset" title="Reset Layout (R)">⟳</button>
                        <button class="btn" id="btn-live-toggle" title="Live Updates">●</button>
                    </div>
                    <div id="graph-workspace-wrapper" style="flex: 1; overflow: hidden; position: relative;">
                        <div id="graph-workspace">
                            <svg id="graph-svg"></svg>
                            <div id="graph-nodes"></div>
                        </div>
                    </div>
                    <div class="annotation-layer" id="graph-annotations"></div>
                </div>

                <!-- Office -->
                <div id="view-office" class="view-container" style="display: flex; flex-direction: column; height: 100%;">
                    <div id="office-toolbar">
                        <select id="office-filter-status">
                            <option value="">All Status</option>
                            <option value="working">Working</option>
                            <option value="active">Active</option>
                            <option value="waiting">Waiting</option>
                        </select>
                        <div class="spacer"></div>
                        <button class="btn" id="btn-office-grid" title="Grid View">⊞</button>
                        <button class="btn active" id="btn-office-list" title="List View">☰</button>
                        <button class="btn" id="btn-office-fit" title="Fit View">⛶</button>
                        <button class="btn" id="btn-office-live" title="Live Updates">●</button>
                    </div>
                    <div id="office-workspace-wrapper" style="flex: 1; overflow: hidden; position: relative;">
                        <div id="office-workspace">
                            <div class="isometric-grid" id="office-grid"></div>
                            <div class="console-area" id="office-console"></div>
                        </div>
                        <div class="office-list" id="office-list"></div>
                    </div>
                    <div class="annotation-layer" id="office-annotations"></div>
                </div>

                <!-- Agents (List) -->
                <div id="view-agents" class="view-container" style="padding: 2rem; max-width: 1200px; margin: 0 auto;">
                    <h2 style="margin-bottom: 1.5rem;">Monitored Processes</h2>
                    <div id="agents-list-full"></div>
                </div>
                
                <!-- Activity -->
                <div id="view-activity" class="view-container" style="padding: 2rem; max-width: 1200px; margin: 0 auto;">
                    <h2 style="margin-bottom: 1.5rem;">Activity Stream</h2>
                    <div id="activity-timeline-full"></div>
                </div>

                <!-- Alerts -->
                <div id="view-alerts" class="view-container" style="padding: 2rem; max-width: 1200px; margin: 0 auto;">
                    <h2 style="margin-bottom: 1.5rem;">Alert Center</h2>
                    <div id="alerts-list-full"></div>
                </div>
            </div>
        </main>

        <div class="inspector-stack" id="inspector-stack"></div>

        <nav class="mobile-nav">
            <a class="mobile-nav-item" onclick="engine.switchView('overview')">Home</a>
            <a class="nav-item" onclick="engine.switchView('graph')">Graph</a>
            <a class="nav-item" onclick="engine.switchView('office')">Office</a>
            <a class="nav-item" onclick="engine.switchView('agents')">Agents</a>
        </nav>
    </div>

    <script>
        const engine = {
            view: 'overview',
            agents: [],
            events: [],
            health: {},
            panels: [],
            // Graph zoom/pan state
            graphScale: 1,
            graphTranslateX: 0,
            graphTranslateY: 0,
            graphLive: true,

            init() {
                this.switchView('overview');
                setInterval(() => this.update(), 5000);
                setInterval(() => {
                    document.getElementById('clock').textContent = new Date().toLocaleTimeString('en-GB');
                }, 1000);
                this.update();
                
                // Graph toolbar events
                this.setupGraphToolbar();
                // Keyboard shortcuts
                this.setupKeyboardShortcuts();
            },
            
            setupGraphToolbar() {
                const search = document.getElementById('graph-search');
                const filterType = document.getElementById('graph-filter-type');
                const filterStatus = document.getElementById('graph-filter-status');
                if (search) search.addEventListener('input', () => this.applyGraphFilters());
                if (filterType) filterType.addEventListener('change', () => this.applyGraphFilters());
                if (filterStatus) filterStatus.addEventListener('change', () => this.applyGraphFilters());
                
                const zoomIn = document.getElementById('btn-zoom-in');
                const zoomOut = document.getElementById('btn-zoom-out');
                const zoomFit = document.getElementById('btn-zoom-fit');
                const layoutReset = document.getElementById('btn-layout-reset');
                const liveToggle = document.getElementById('btn-live-toggle');
                
                if (zoomIn) zoomIn.addEventListener('click', () => this.graphZoom(1.2));
                if (zoomOut) zoomOut.addEventListener('click', () => this.graphZoom(0.833));
                if (zoomFit) zoomFit.addEventListener('click', () => this.graphFit());
                if (layoutReset) layoutReset.addEventListener('click', () => this.graphResetLayout());
                if (liveToggle) liveToggle.addEventListener('click', () => this.toggleLive());
            },
            
            setupKeyboardShortcuts() {
                document.addEventListener('keydown', (e) => {
                    if (e.target.tagName === 'INPUT' || e.target.tagName === 'SELECT') return;
                    
                    if (this.view === 'graph') {
                        switch (e.key) {
                            case '+': case '=': e.preventDefault(); this.graphZoom(1.2); break;
                            case '-': case '_': e.preventDefault(); this.graphZoom(0.833); break;
                            case '0': e.preventDefault(); this.graphFit(); break;
                            case 'r': case 'R': e.preventDefault(); this.graphResetLayout(); break;
                            case 'Escape': this.closeTopPanel(); break;
                        }
                    }
                    if (e.key === 'Escape') this.closeTopPanel();
                });
            },
            
            closeTopPanel() {
                const stack = document.getElementById('inspector-stack');
                const panels = stack.querySelectorAll('.panel');
                if (panels.length > 0) {
                    const top = panels[panels.length - 1];
                    top.classList.remove('active');
                    setTimeout(() => top.remove(), 300);
                }
            },
            
            graphZoom(factor) {
                this.graphScale = Math.max(0.25, Math.min(3, this.graphScale * factor));
                this.applyGraphTransform();
            },
            
            graphFit() {
                const wrapper = document.getElementById('graph-workspace-wrapper');
                const workspace = document.getElementById('graph-workspace');
                if (!wrapper || !workspace) return;
                
                const nodes = workspace.querySelectorAll('.node:not([style*="display: none"])');
                if (nodes.length === 0) {
                    this.graphScale = 1;
                    this.graphTranslateX = 0;
                    this.graphTranslateY = 0;
                } else {
                    let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
                    nodes.forEach(n => {
                        const rect = n.getBoundingClientRect();
                        const wRect = workspace.getBoundingClientRect();
                        const x = (rect.left - wRect.left) / this.graphScale;
                        const y = (rect.top - wRect.top) / this.graphScale;
                        minX = Math.min(minX, x);
                        minY = Math.min(minY, y);
                        maxX = Math.max(maxX, x + n.offsetWidth);
                        maxY = Math.max(maxY, y + n.offsetHeight);
                    });
                    const padding = 100;
                    const contentWidth = maxX - minX + padding * 2;
                    const contentHeight = maxY - minY + padding * 2;
                    const wrapperRect = wrapper.getBoundingClientRect();
                    const scaleX = wrapperRect.width / contentWidth;
                    const scaleY = wrapperRect.height / contentHeight;
                    this.graphScale = Math.min(scaleX, scaleY, 1);
                    this.graphTranslateX = -minX * this.graphScale + padding * this.graphScale;
                    this.graphTranslateY = -minY * this.graphScale + padding * this.graphScale;
                }
                this.applyGraphTransform();
            },
            
            graphResetLayout() {
                this.graphScale = 1;
                this.graphTranslateX = 0;
                this.graphTranslateY = 0;
                this.renderGraph();
            },
            
            toggleLive() {
                this.graphLive = !this.graphLive;
                const btn = document.getElementById('btn-live-toggle');
                if (btn) {
                    btn.classList.toggle('active', this.graphLive);
                    btn.textContent = this.graphLive ? '●' : '○';
                }
            },
            
            applyGraphTransform() {
                const workspace = document.getElementById('graph-workspace');
                if (workspace) {
                    workspace.style.transform = 'translate(' + this.graphTranslateX + 'px, ' + this.graphTranslateY + 'px) scale(' + this.graphScale + ')';
                }
            },

            async update() {
                if (!this.graphLive) return;
                try {
                    const [a, e, h] = await Promise.all([
                        fetch('/api/agents').then(r => r.json()),
                        fetch('/api/events?limit=50').then(r => r.json()),
                        fetch('/api/health').then(r => r.json())
                    ]);
                    this.agents = a;
                    this.events = e;
                    this.health = h;
                    this.render();
                } catch (err) {
                    console.error('Update failed', err);
                    document.getElementById('status-tag').textContent = '○ OFFLINE';
                    document.getElementById('status-tag').style.color = 'var(--danger)';
                }
            },

            switchView(view) {
                this.view = view;
                document.querySelectorAll('.view-container').forEach(v => v.classList.remove('active'));
                document.getElementById('view-' + view).classList.add('active');
                document.getElementById('view-title').textContent = view;
                
                document.querySelectorAll('.nav-item').forEach(n => {
                    n.classList.toggle('active', n.textContent.toLowerCase() === view);
                });

                this.render();
            },

            render() {
                this.renderStats();
                if (this.view === 'overview') this.renderOverview();
                if (this.view === 'graph') this.renderGraph();
                if (this.view === 'office') this.renderOffice();
                if (this.view === 'agents') this.renderAgents();
                if (this.view === 'activity') this.renderActivity();
                if (this.view === 'alerts') this.renderAlerts();
            },

            renderStats() {
                const totalCpu = this.agents.reduce((s, a) => s + a.cpu, 0);
                const totalRam = this.agents.reduce((s, a) => s + a.ram, 0);
                const ramMb = Math.round(totalRam / 1024 / 1024);

                document.getElementById('total-agents-val').textContent = this.agents.length;
                document.getElementById('active-cpu-val').textContent = totalCpu.toFixed(1) + '%';
                document.getElementById('active-ram-val').textContent = ramMb + 'MB';
                document.getElementById('ntfy-status-val').textContent = this.health.ntfy_enabled ? 'ON' : 'OFF';
                
                document.getElementById('mini-cpu').textContent = totalCpu.toFixed(1) + '%';
                document.getElementById('mini-mem').textContent = ramMb + 'MB';
            },

            renderOverview() {
                const evContainer = document.getElementById('overview-events');
                evContainer.innerHTML = this.events.slice(0, 10).map(function(e) {
    return '<div style="padding: 0.75rem 1rem; border-bottom: 1px solid var(--border); display: flex; gap: 1rem; font-size: 0.85rem;">' +
           '<span style="color: var(--muted);" class="monospace">' + e.created_at.slice(11, 19) + '</span>' +
           '<span style="color: var(--primary); font-weight: 600;">' + e.agent_name + '</span>' +
           '<span>' + e.message + '</span>' +
           '</div>'; 
}).join('');

                const alertContainer = document.getElementById('overview-alerts');
                const highCpu = this.agents.filter(a => a.cpu > 80);
                if (highCpu.length === 0) {
                    alertContainer.innerHTML = '<div class="card" style="color: var(--muted); font-size: 0.85rem;">✓ No critical issues</div>';
                } else {
                    alertContainer.innerHTML = highCpu.map(function(a) {
    return '<div class="card" style="border-left: 3px solid var(--danger);">' +
           '<div style="font-weight: 600; color: var(--danger); font-size: 0.7rem; text-transform: uppercase;">High CPU</div>' +
           '<div style="font-size: 0.85rem;">' + a.name + ' is using ' + a.cpu.toFixed(1) + '%</div>' +
           '</div>'; 
}).join('');
                }
            },

            renderGraph() {
                const container = document.getElementById('graph-nodes');
                const svg = document.getElementById('graph-svg');
                const annotations = document.getElementById('graph-annotations');
                
                // Agent SVG icons
                const agentIcons = {
                    'OpenCode': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="node-icon"><rect x="2" y="3" width="20" height="14" rx="2"/><path d="M8 10h8M8 14h5"/></svg>',
                    'Hermes': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="node-icon"><path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/></svg>',
                    'Goose': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="node-icon"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h11"/><polyline points="15 3 21 3 21 9"/></svg>',
                    'Claude Code': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="node-icon"><circle cx="12" cy="12" r="10"/><path d="M8 14s1.5 2 4 2 4-2 4-2"/><line x1="9" y1="9" x2="9.01" y2="9"/><line x1="15" y1="9" x2="15.01" y2="9"/></svg>',
                    'Kilo': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="node-icon"><polygon points="12 2 22 8.5 22 15.5 12 22 2 15.5 2 8.5 12 2"/><line x1="12" y1="22" x2="12" y2="15.5"/><line x1="22" y1="8.5" x2="12" y2="15.5"/><line x1="2" y1="8.5" x2="12" y2="15.5"/></svg>',
                    'Codex': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="node-icon"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>',
                    'Cursor Agent': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="node-icon"><path d="M3 3h18v18H3z"/><path d="M12 8v8M8 12h8"/></svg>',
                    'Cline': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="node-icon"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>',
                    'Kiro': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="node-icon"><circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/></svg>',
                    'Droid': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="node-icon"><rect x="5" y="2" width="14" height="20" rx="2"/><line x1="12" y1="18" x2="12.01" y2="18"/><path d="M9 6h6"/></svg>',
                    'Amp': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="node-icon"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>',
                    'Grok': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="node-icon"><path d="M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2z"/><path d="M12 6v6l4 2"/></svg>',
                    'Qwen Code': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="node-icon"><path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/></svg>',
                    'Hermes Agent': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="node-icon"><path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/></svg>',
                    'ORIONOS CORE': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="node-icon"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>'
                };
                
                const getIcon = (name) => agentIcons[name] || agentIcons['OpenCode'];
                
                // Use dagre for layout
                const g = new dagre.graphlib.Graph();
                g.setGraph({ rankdir: 'TB', nodesep: 120, ranksep: 150, marginx: 100, marginy: 100 });
                g.setDefaultEdgeLabel(() => ({}));
                
                // Add root node
                g.setNode('root', { width: 200, height: 80, label: 'ORIONOS CORE' });
                
                // Add agent nodes
                this.agents.forEach(a => {
                    g.setNode(a.id, { 
                        width: 200, 
                        height: 100, 
                        label: a.name + '|' + a.status + '|CPU ' + a.cpu.toFixed(1) + '%|RAM ' + Math.round(a.ram/1024/1024) + 'MB' 
                    });
                    g.setEdge('root', a.id);
                });
                
                dagre.layout(g);
                
                // Render nodes
                let html = '';
                g.nodes().forEach(nodeId => {
                    const node = g.node(nodeId);
                    const x = node.x - node.width / 2;
                    const y = node.y - node.height / 2;
                    
                    if (nodeId === 'root') {
                        html += '<div class="node" style="left: ' + x + 'px; top: ' + y + 'px;" id="node-root">' +
                            '<div class="node-header">' +
                                '<span class="node-title">' + getIcon('ORIONOS CORE') + '<strong>ORIONOS CORE</strong></span>' +
                                '<span class="node-status" style="background: var(--primary);"></span>' +
                            '</div>' +
                            '<div style="font-size: 0.75rem; color: var(--muted);" class="monospace">SYSTEM MONITOR</div>' +
                        '</div>';
                    } else {
                        const agent = this.agents.find(a => a.id === nodeId);
                        if (!agent) return;
                        const statusColor = agent.status === 'working' ? 'success' : (agent.status === 'active' ? 'primary' : 'muted');
                        html += '<div class="node" style="left: ' + x + 'px; top: ' + y + 'px;" onclick="engine.inspect(\'agent\', \'" + agent.id + "\')">' +
                            '<div class="node-header">' +
                                '<span class="node-title">' + getIcon(agent.name) + '<span style="font-weight: 600;">' + agent.name + '</span></span>' +
                                '<span class="node-status" style="background: var(--' + statusColor + ');"></span>' +
                            '</div>' +
                            '<div style="font-size: 0.75rem; color: var(--muted);" class="monospace">' +
                                'CPU ' + agent.cpu.toFixed(1) + '%<br>' +
                                'RAM ' + Math.round(agent.ram/1024/1024) + 'MB' +
                            '</div>' +
                        '</div>';
                    }
                });
                
                container.innerHTML = html;
                
                // Apply search/filter
                this.applyGraphFilters();
                
                // Draw edges using SVG paths
                let svgHtml = '';
                g.edges().forEach(e => {
                    const source = g.node(e.v);
                    const target = g.node(e.w);
                    if (!source || !target) return;
                    const sx = source.x;
                    const sy = source.y + source.height / 2;
                    const tx = target.x;
                    const ty = target.y - target.height / 2;
                    // Curved path
                    const cx = sx;
                    const cy = (sy + ty) / 2;
                    svgHtml += '<path d="M ' + sx + ' ' + sy + ' C ' + cx + ' ' + cy + ' ' + cx + ' ' + cy + ' ' + tx + ' ' + ty + '" stroke="var(--border)" stroke-width="1.5" fill="none" marker-end="url(#arrowhead)"/>';
                });
                
                // Add arrowhead marker definition
                svg.innerHTML = '<defs><marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto"><polygon points="0 0, 10 3.5, 0 7" fill="var(--border)"/></marker></defs>' + svgHtml;
                
                annotations.innerHTML = '<div class="note" style="left: 20px; top: 20px;">↳ ' + this.agents.length + ' agents connected to core</div>';
                
                // Setup pan (drag to pan)
                this.setupGraphPan();
            },
            
            setupGraphPan() {
                const workspace = document.getElementById('graph-workspace');
                const wrapper = document.getElementById('graph-workspace-wrapper');
                if (!workspace || !wrapper) return;
                
                // Remove old listeners
                workspace._panHandler = workspace._panHandler || {
                    down: (e) => {
                        if (e.target.closest('.node')) return; // Don't pan when clicking nodes
                        this.isPanning = true;
                        this.panStartX = e.clientX - this.graphTranslateX;
                        this.panStartY = e.clientY - this.graphTranslateY;
                        wrapper.style.cursor = 'grabbing';
                        e.preventDefault();
                    },
                    move: (e) => {
                        if (!this.isPanning) return;
                        this.graphTranslateX = e.clientX - this.panStartX;
                        this.graphTranslateY = e.clientY - this.panStartY;
                        this.applyGraphTransform();
                    },
                    up: () => {
                        this.isPanning = false;
                        wrapper.style.cursor = 'grab';
                    }
                };
                
                wrapper.style.cursor = 'grab';
                wrapper.removeEventListener('mousedown', workspace._panHandler.down);
                wrapper.removeEventListener('mousemove', workspace._panHandler.move);
                wrapper.removeEventListener('mouseup', workspace._panHandler.up);
                wrapper.removeEventListener('mouseleave', workspace._panHandler.up);
                
                wrapper.addEventListener('mousedown', workspace._panHandler.down);
                wrapper.addEventListener('mousemove', workspace._panHandler.move);
                wrapper.addEventListener('mouseup', workspace._panHandler.up);
                wrapper.addEventListener('mouseleave', workspace._panHandler.up);
                
                // Wheel zoom
                wrapper.onwheel = (e) => {
                    e.preventDefault();
                    const factor = e.deltaY > 0 ? 0.9 : 1.1;
                    const rect = wrapper.getBoundingClientRect();
                    const mouseX = e.clientX - rect.left;
                    const mouseY = e.clientY - rect.top;
                    
                    // Zoom towards mouse position
                    const newScale = Math.max(0.25, Math.min(3, this.graphScale * factor));
                    const scaleRatio = newScale / this.graphScale;
                    this.graphTranslateX = mouseX - (mouseX - this.graphTranslateX) * scaleRatio;
                    this.graphTranslateY = mouseY - (mouseY - this.graphTranslateY) * scaleRatio;
                    this.graphScale = newScale;
                    this.applyGraphTransform();
                };
            },
            
            applyGraphFilters() {
                const search = document.getElementById('graph-search').value.toLowerCase();
                const typeFilter = document.getElementById('graph-filter-type').value;
                const statusFilter = document.getElementById('graph-filter-status').value;
                
                document.querySelectorAll('#graph-nodes .node').forEach(node => {
                    const name = node.querySelector('.node-title')?.textContent?.toLowerCase() || '';
                    const isRoot = node.id === 'node-root';
                    const nodeType = isRoot ? 'core' : 'agent';
                    const status = node.querySelector('.node-status')?.style.backgroundColor?.replace('var(--', '')?.replace(')', '') || '';
                    
                    let show = true;
                    if (search && !name.includes(search)) show = false;
                    if (typeFilter && nodeType !== typeFilter) show = false;
                    if (statusFilter && status !== statusFilter) show = false;
                    
                    node.style.display = show ? 'block' : 'none';
                });
            },

            renderOffice() {
                const grid = document.getElementById('office-grid');
                const list = document.getElementById('office-list');
                const console = document.getElementById('office-console');
                const annotations = document.getElementById('office-annotations');
                
                // Agent SVG icons (reuse from graph)
                const agentIcons = {
                    'OpenCode': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="avatar-icon"><rect x="2" y="3" width="20" height="14" rx="2"/><path d="M8 10h8M8 14h5"/></svg>',
                    'Hermes': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="avatar-icon"><path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/></svg>',
                    'Goose': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="avatar-icon"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h11"/><polyline points="15 3 21 3 21 9"/></svg>',
                    'Claude Code': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="avatar-icon"><circle cx="12" cy="12" r="10"/><path d="M8 14s1.5 2 4 2 4-2 4-2"/><line x1="9" y1="9" x2="9.01" y2="9"/><line x1="15" y1="9" x2="15.01" y2="9"/></svg>',
                    'Kilo': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="avatar-icon"><polygon points="12 2 22 8.5 22 15.5 12 22 2 15.5 2 8.5 12 2"/><line x1="12" y1="22" x2="12" y2="15.5"/><line x1="22" y1="8.5" x2="12" y2="15.5"/><line x1="2" y1="8.5" x2="12" y2="15.5"/></svg>',
                    'Codex': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="avatar-icon"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>',
                    'Cursor Agent': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="avatar-icon"><path d="M3 3h18v18H3z"/><path d="M12 8v8M8 12h8"/></svg>',
                    'Cline': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="avatar-icon"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>',
                    'Kiro': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="avatar-icon"><circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/></svg>',
                    'Droid': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="avatar-icon"><rect x="5" y="2" width="14" height="20" rx="2"/><line x1="12" y1="18" x2="12.01" y2="18"/><path d="M9 6h6"/></svg>',
                    'Amp': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="avatar-icon"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>',
                    'Grok': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="avatar-icon"><path d="M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2z"/><path d="M12 6v6l4 2"/></svg>',
                    'Qwen Code': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="avatar-icon"><path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/></svg>',
                    'Hermes Agent': '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="avatar-icon"><path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/></svg>'
                };
                
                const getIcon = (name) => agentIcons[name] || agentIcons['OpenCode'];
                
                // Grid view
                let html = '';
                this.agents.forEach((a, i) => {
                    const row = Math.floor(i / 3);
                    const col = i % 3;
                    const x = col * 200 + 100;
                    const y = row * 200 + 100;
                    
                    const statusColor = a.status === 'working' ? 'success' : (a.status === 'active' ? 'primary' : 'muted');
                    html += '<div class="station ' + a.status + '" style="left: ' + x + 'px; top: ' + y + 'px;" onclick="engine.inspect(\'agent\', \'" + a.id + "\')">' +
                        '<div class="station-avatar">' + getIcon(a.name) + '</div>' +
                        '<div class="station-label">' + a.name + '</div>' +
                        '<div class="station-status" style="background: var(--' + statusColor + ');"></div>' +
                    '</div>';
                });
                grid.innerHTML = html;
                
                // List view
                let listHtml = '';
                this.agents.forEach(a => {
                    const statusColor = a.status === 'working' ? 'success' : (a.status === 'active' ? 'primary' : 'muted');
                    listHtml += '<div class="station-card" onclick="engine.inspect(\'agent\', \'" + a.id + "\')">' +
                        '<div class="station-avatar-sm" style="background: var(--' + statusColor + ');">' + getIcon(a.name) + '</div>' +
                        '<div class="station-info">' +
                            '<div class="station-name">' + a.name + '</div>' +
                            '<div class="station-meta">' + a.status.toUpperCase() + ' · CPU ' + a.cpu.toFixed(1) + '% · RAM ' + Math.round(a.ram/1024/1024) + 'MB · PID ' + a.pid + '</div>' +
                        '</div>' +
                    '</div>';
                });
                list.innerHTML = listHtml;
                
                // Console - show recent events
                if (console) {
                    const recentEvents = this.events.slice(0, 8);
                    console.innerHTML = '<div class="console-title">SYSTEM CONSOLE · EVENT STREAM</div>' +
                        recentEvents.map(e => '<div class="console-line"><span style="color: var(--muted);">[' + e.created_at.slice(11, 19) + ']</span> <span style="color: var(--primary);">' + e.agent_name + '</span> ' + e.message + '</div>').join('');
                }
                
                annotations.innerHTML = '<div class="note" style="left: 60%; top: 20%;">"' + (this.agents.find(a => a.name === 'Goose') ? 'Goose seems busy today' : 'All agents operational') + '"</div>';
                
                // Setup office toolbar
                this.setupOfficeToolbar();
            },
            
            setupOfficeToolbar() {
                const filterStatus = document.getElementById('office-filter-status');
                const btnGrid = document.getElementById('btn-office-grid');
                const btnList = document.getElementById('btn-office-list');
                const btnFit = document.getElementById('btn-office-fit');
                const btnLive = document.getElementById('btn-office-live');
                
                if (filterStatus) filterStatus.addEventListener('change', () => this.applyOfficeFilters());
                if (btnGrid) btnGrid.addEventListener('click', () => this.setOfficeView('grid'));
                if (btnList) btnList.addEventListener('click', () => this.setOfficeView('list'));
                if (btnFit) btnFit.addEventListener('click', () => this.officeFitView());
                if (btnLive) btnLive.addEventListener('click', () => this.toggleOfficeLive());
            },
            
            applyOfficeFilters() {
                const statusFilter = document.getElementById('office-filter-status').value;
                
                document.querySelectorAll('#office-grid .station, #office-list .station-card').forEach(el => {
                    const status = el.classList.contains('working') ? 'working' : 
                                  el.classList.contains('active') ? 'active' : 'waiting';
                    el.style.display = (!statusFilter || status === statusFilter) ? '' : 'none';
                });
            },
            
            setOfficeView(mode) {
                const grid = document.getElementById('office-grid');
                const list = document.getElementById('office-list');
                const btnGrid = document.getElementById('btn-office-grid');
                const btnList = document.getElementById('btn-office-list');
                
                if (mode === 'grid') {
                    if (grid) grid.style.display = '';
                    if (list) list.style.display = 'none';
                    if (btnGrid) btnGrid.classList.add('active');
                    if (btnList) btnList.classList.remove('active');
                } else {
                    if (grid) grid.style.display = 'none';
                    if (list) list.style.display = 'block';
                    if (btnGrid) btnGrid.classList.remove('active');
                    if (btnList) btnList.classList.add('active');
                }
            },
            
            officeFitView() {
                // Reset grid view positions
                this.renderOffice();
            },
            
            toggleOfficeLive() {
                this.graphLive = !this.graphLive; // reuse same live toggle
                const btn = document.getElementById('btn-office-live');
                if (btn) {
                    btn.classList.toggle('active', this.graphLive);
                    btn.textContent = this.graphLive ? '●' : '○';
                }
            },

            renderAgents() {
                const list = document.getElementById('agents-list-full');
                list.innerHTML = this.agents.map(function(a) {
    return '<div class="card" style="display: grid; grid-template-columns: 1.5fr 1fr 1fr 1fr; align-items: center; cursor: pointer;" onclick="engine.inspect(\'agent\', \'" + a.id + "\')">" +
        '<div>' +
            '<div style="font-weight: 600;">' + a.name + '</div>' +
            '<div class="monospace" style="font-size: 0.75rem; color: var(--muted);">PID ' + a.pid + '</div>' +
        '</div>' +
        '<div><span class="badge ' + a.status + '">' + a.status + '</span></div>' +
        '<div class="monospace">' + a.cpu.toFixed(1) + '% CPU</div>' +
        '<div class="monospace">' + Math.round(a.ram/1024/1024) + 'MB RAM</div>' +
    '</div>'; 
}).join('');
            },

            renderActivity() {
                const list = document.getElementById('activity-timeline-full');
                list.innerHTML = this.events.map(function(e) {
    return '<div style="padding: 1rem; border-bottom: 1px solid var(--border); display: flex; gap: 2rem;">' +
        '<div class="monospace" style="color: var(--muted); font-size: 0.8rem;">' + e.created_at.replace('T', ' ').slice(0, 19) + '</div>' +
        '<div>' +
            '<div style="font-weight: 600; color: var(--primary); margin-bottom: 0.25rem;">' + e.agent_name + '</div>' +
            '<div style="font-size: 0.9rem;">' + e.message + '</div>' +
        '</div>' +
    '</div>'; 
}).join('');
            },

            renderAlerts() {
                const list = document.getElementById('alerts-list-full');
                const highCpu = this.agents.filter(a => a.cpu > 80);
                if (highCpu.length === 0) {
                    list.innerHTML = '<div class="empty-state" style="padding: 5rem; text-align: center; color: var(--muted);">✓ All systems operational. No active alerts.</div>';
                } else {
                    list.innerHTML = highCpu.map(function(a) {
    return '<div class="card" style="border-left: 4px solid var(--danger);">' +
        '<div style="font-weight: 700; color: var(--danger); text-transform: uppercase; font-size: 0.75rem; margin-bottom: 0.5rem;">CRITICAL: HIGH CPU</div>' +
        '<div>Process <strong>' + a.name + '</strong> (PID ' + a.pid + ') is consuming ' + a.cpu.toFixed(1) + '% CPU.</div>' +
    '</div>'; 
}).join('');
                }
            },

            inspect(type, id) {
                const agent = this.agents.find(a => a.id === id);
                if (!agent) return;

                const panelHtml = '<div class="panel" id="panel-' + id + '">' +
    '<div class="panel-header">' +
        '<div style="font-weight: 700; text-transform: uppercase; font-size: 0.8rem;">Agent Detail</div>' +
        '<button onclick="engine.closePanel(\'' + id + '\')" style="background: none; border: none; color: var(--muted); cursor: pointer;">✕</button>' +
    '</div>' +
    '<div class="panel-content">' +
        '<div style="margin-bottom: 2rem;">' +
            '<h2 style="font-size: 1.5rem; margin-bottom: 0.25rem;">' + agent.name + '</h2>' +
            '<div class="badge ' + agent.status + '">' + agent.status + '</div>' +
        '</div>' +
        '<div style="display: grid; gap: 1rem; margin-bottom: 2rem;">' +
            '<div class="card" style="margin: 0;">' +
                '<div style="font-size: 0.7rem; color: var(--muted); text-transform: uppercase; margin-bottom: 0.5rem;">Performance</div>' +
                '<div class="monospace">CPU: ' + agent.cpu.toFixed(1) + '%</div>' +
                '<div class="monospace">RAM: ' + Math.round(agent.ram/1024/1024) + 'MB</div>' +
            '</div>' +
            '<div class="card" style="margin: 0;">' +
                '<div style="font-size: 0.7rem; color: var(--muted); text-transform: uppercase; margin-bottom: 0.5rem;">Process</div>' +
                '<div class="monospace">PID: ' + agent.pid + '</div>' +
                '<div class="monospace">Started: ' + agent.started_at.slice(11, 19) + '</div>' +
            '</div>' +
        '</div>' +
        '<h3 style="font-size: 0.8rem; color: var(--muted); text-transform: uppercase; margin-bottom: 1rem;">Recent History</h3>' +
        '<div class="monospace" style="font-size: 0.75rem;">' +
            this.events.filter(function(e) { return e.agent_name === agent.name; }).slice(0,5).map(function(e) { return '<div style="margin-bottom: 0.5rem; padding-bottom: 0.5rem; border-bottom: 1px solid var(--border);"><span style="color: var(--muted);">' + e.created_at.slice(11,19) + '</span><br>' + e.message + '</div>'; }).join('') +
        '</div>' +
    '</div>' +
'</div>';

                const stack = document.getElementById('inspector-stack');
                stack.innerHTML = panelHtml;
                setTimeout(() => {
                    document.getElementById('panel-' + id).classList.add('active');
                }, 10);
            },

            closePanel(id) {
                const p = document.getElementById('panel-' + id);
                if (p) {
                    p.classList.remove('active');
                    setTimeout(() => p.remove(), 300);
                }
            }
        };

        window.onload = () => engine.init();
    </script>
</body>
</html>`
