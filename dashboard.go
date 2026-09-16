package main

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AgentDock</title>
    <script src="https://unpkg.com/htmx.org@1.9.12"></script>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'JetBrains Mono', 'Fira Code', monospace; background: #0f0f14; color: #e9e9f0; }
        
        .header { display: flex; justify-content: space-between; align-items: center; padding: 1rem 2rem; border-bottom: 1px solid #1e1e2e; }
        .header h1 { font-size: 1.5rem; color: #7aa2f7; }
        .header .status { color: #9ece6a; font-size: 0.875rem; }
        
        .stats { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; padding: 1rem 2rem; }
        .stat-card { background: #16161e; border: 1px solid #1e1e2e; border-radius: 8px; padding: 1.25rem; }
        .stat-card .label { color: #565f89; font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.05em; }
        .stat-card .value { color: #e9e9f0; font-size: 2rem; font-weight: 700; margin-top: 0.25rem; }
        .stat-card .unit { color: #565f89; font-size: 0.875rem; margin-left: 0.25rem; }
        
        .section { padding: 1rem 2rem; }
        .section h2 { font-size: 1rem; color: #7aa2f7; margin-bottom: 0.75rem; text-transform: uppercase; letter-spacing: 0.05em; }
        
        .agent-list { display: grid; gap: 0.5rem; }
        .agent-row { display: grid; grid-template-columns: 2fr 1fr 1fr 1fr; gap: 1rem; align-items: center; padding: 0.75rem 1rem; background: #16161e; border: 1px solid #1e1e2e; border-radius: 6px; }
        .agent-row .name { font-weight: 600; color: #e9e9f0; }
        .agent-row .pid { color: #565f89; font-size: 0.75rem; }
        .agent-row .cpu { color: #f7768e; }
        .agent-row .ram { color: #e9e9f0; }
        .agent-row .status { display: inline-block; padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.75rem; text-transform: uppercase; }
        .agent-row .status.working { background: #1a1b26; color: #9ece6a; border: 1px solid #9ece6a; }
        .agent-row .status.idle { background: #1a1b26; color: #565f89; border: 1px solid #565f89; }
        
        .event-list { display: grid; gap: 0.5rem; }
        .event-row { display: grid; grid-template-columns: 1fr 2fr 1fr; gap: 1rem; padding: 0.5rem 1rem; background: #16161e; border: 1px solid #1e1e2e; border-radius: 4px; font-size: 0.875rem; }
        .event-row .type { color: #7aa2f7; font-weight: 600; }
        .event-row .message { color: #a9b1d6; }
        .event-row .time { color: #565f89; font-size: 0.75rem; text-align: right; }
        
        .empty { text-align: center; padding: 2rem; color: #565f89; }
        
        .footer { text-align: center; padding: 1rem; color: #3b4261; font-size: 0.75rem; border-top: 1px solid #1e1e2e; }
    </style>
</head>
<body>
    <div class="header">
        <h1>AgentDock</h1>
        <div class="status" hx-get="/api/health" hx-trigger="every 10s">● Online</div>
    </div>
    
    <div class="stats" hx-get="/api/agents" hx-trigger="every 5s">
        <div class="stat-card">
            <div class="label">Agents</div>
            <div class="value" id="agent-count">0</div>
        </div>
        <div class="stat-card">
            <div class="label">CPU Total</div>
            <div class="value" id="cpu-total">0<span class="unit">%</span></div>
        </div>
        <div class="stat-card">
            <div class="label">RAM Total</div>
            <div class="value" id="ram-total">0<span class="unit">MB</span></div>
        </div>
    </div>
    
    <div class="section">
        <h2>Running Agents</h2>
        <div class="agent-list" hx-get="/api/agents" hx-trigger="every 5s">
            <div class="empty">No agents detected</div>
        </div>
    </div>
    
    <div class="section">
        <h2>Recent Events</h2>
        <div class="event-list" hx-get="/api/events?limit=20" hx-trigger="every 10s">
            <div class="empty">No events yet</div>
        </div>
    </div>
    
    <div class="footer">AgentDock v0.1.0 — Lightweight AI Agent Monitor</div>
    
    <script>
        document.body.onhtmxAfterSwap = function(evt) {
            if (evt.detail.target.classList.contains('agent-list')) {
                updateStats(evt.detail.xhr.responseText);
            }
        };
        
        function updateStats(responseText) {
            try {
                const agents = JSON.parse(responseText);
                document.getElementById('agent-count').textContent = agents.length;
                let totalCPU = 0;
                let totalRAM = 0;
                agents.forEach(a => {
                    totalCPU += a.cpu;
                    totalRAM += (a.ram / 1024 / 1024);
                });
                document.getElementById('cpu-total').innerHTML = totalCPU.toFixed(1) + '<span class="unit">%</span>';
                document.getElementById('ram-total').innerHTML = Math.round(totalRAM) + '<span class="unit">MB</span>';
            } catch(e) {}
        }
    </script>
</body>
</html>`
