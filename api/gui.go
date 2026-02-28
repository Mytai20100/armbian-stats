package api

const Gui = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Armbian Stats</title>
<script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
<style>
  :root {
    --bg:         {{.Background}};
    --surface:    {{.Surface}};
    --surface2:   {{.SurfaceAlt}};
    --primary:    {{.Primary}};
    --secondary:  {{.Secondary}};
    --accent:     {{.Accent}};
    --warning:    {{.Warning}};
    --text:       {{.Text}};
    --muted:      {{.TextMuted}};
    --border:     {{.Border}};
  }

  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

  body {
    background: var(--bg);
    color: var(--text);
    font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace, sans-serif;
    font-size: 13px;
    line-height: 1.6;
    min-height: 100vh;
  }

  header {
    background: var(--surface);
    border-bottom: 1px solid var(--border);
    padding: 0 24px;
    height: 52px;
    display: flex;
    align-items: center;
    gap: 24px;
    position: sticky;
    top: 0;
    z-index: 100;
  }
  .header-logo {
    font-size: 15px;
    font-weight: 700;
    color: var(--primary);
    letter-spacing: 1px;
    display: flex;
    align-items: center;
    gap: 8px;
    white-space: nowrap;
  }
  .header-logo svg { width: 20px; height: 20px; }
  .header-info {
    display: flex;
    gap: 16px;
    flex: 1;
    overflow: hidden;
  }
  .header-pill {
    display: flex;
    align-items: center;
    gap: 6px;
    background: var(--surface2);
    border: 1px solid var(--border);
    border-radius: 20px;
    padding: 3px 12px;
    font-size: 11px;
    color: var(--muted);
    white-space: nowrap;
  }
  .header-pill span { color: var(--text); font-weight: 600; }
  .status-dot {
    width: 6px; height: 6px;
    border-radius: 50%;
    background: var(--secondary);
    box-shadow: 0 0 6px var(--secondary);
    animation: pulse 2s infinite;
  }
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.4; }
  }
  .header-clock {
    margin-left: auto;
    font-size: 11px;
    color: var(--muted);
    white-space: nowrap;
  }

  main {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(440px, 1fr));
    gap: 16px;
    padding: 20px;
    max-width: 1600px;
    margin: 0 auto;
  }
  .card-full { grid-column: 1 / -1; }

  .card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 16px 20px 20px;
    display: flex;
    flex-direction: column;
    gap: 14px;
    transition: border-color 0.2s;
  }
  .card:hover { border-color: color-mix(in srgb, var(--primary) 40%, var(--border)); }

  .card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }
  .card-title {
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 1.5px;
    color: var(--muted);
    display: flex;
    align-items: center;
    gap: 7px;
  }
  .card-title svg { width: 14px; height: 14px; opacity: 0.6; }
  .card-value {
    font-size: 22px;
    font-weight: 700;
    color: var(--text);
  }

  .chart-wrap {
    position: relative;
    width: 100%;
    height: 140px;
  }
  .chart-wrap canvas { width: 100% !important; height: 100% !important; }

  .stats-row {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }
  .stat-item {
    flex: 1;
    min-width: 80px;
    background: var(--surface2);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 8px 10px;
  }
  .stat-label {
    font-size: 10px;
    color: var(--muted);
    text-transform: uppercase;
    letter-spacing: 0.8px;
    margin-bottom: 3px;
  }
  .stat-value { font-size: 14px; font-weight: 700; }

  .progress-wrap { display: flex; flex-direction: column; gap: 6px; }
  .progress-row { display: flex; align-items: center; gap: 8px; }
  .progress-label {
    font-size: 11px;
    color: var(--muted);
    width: 80px;
    flex-shrink: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .progress-bar-bg {
    flex: 1;
    height: 8px;
    background: var(--surface2);
    border-radius: 4px;
    overflow: hidden;
    border: 1px solid var(--border);
  }
  .progress-bar-fill { height: 100%; border-radius: 4px; transition: width 0.6s ease; }
  .progress-pct { font-size: 11px; font-weight: 600; width: 38px; text-align: right; flex-shrink: 0; }
  .progress-size { font-size: 10px; color: var(--muted); width: 90px; text-align: right; flex-shrink: 0; }

  .core-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(90px, 1fr));
    gap: 6px;
  }
  .core-item {
    background: var(--surface2);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 6px 8px;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .core-name { font-size: 9px; color: var(--muted); text-transform: uppercase; letter-spacing: 0.8px; }
  .core-pct { font-size: 16px; font-weight: 700; line-height: 1.2; }
  .core-mhz { font-size: 9px; color: var(--muted); }
  .core-mini-bar { height: 3px; background: var(--surface); border-radius: 2px; margin-top: 2px; overflow: hidden; }
  .core-mini-fill { height: 100%; border-radius: 2px; background: var(--primary); transition: width 0.5s ease; }

  .temp-grid { display: flex; gap: 8px; flex-wrap: wrap; }
  .temp-badge {
    display: flex;
    flex-direction: column;
    align-items: center;
    background: var(--surface2);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 8px 14px;
    gap: 2px;
    min-width: 80px;
  }
  .temp-name { font-size: 9px; color: var(--muted); text-transform: uppercase; letter-spacing: 0.8px; }
  .temp-val { font-size: 20px; font-weight: 700; }

  .log-box {
    background: var(--surface2);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 10px 14px;
    height: 130px;
    overflow-y: auto;
    font-size: 11px;
    line-height: 1.8;
    color: var(--muted);
    scrollbar-width: thin;
    scrollbar-color: var(--border) transparent;
  }
  .log-box::-webkit-scrollbar { width: 4px; }
  .log-box::-webkit-scrollbar-thumb { background: var(--border); border-radius: 4px; }
  .log-entry { display: flex; gap: 10px; }
  .log-time { color: var(--primary); flex-shrink: 0; font-size: 10px; }
  .log-msg { color: var(--text); }

  .col-primary   { color: var(--primary); }
  .col-secondary { color: var(--secondary); }
  .col-accent    { color: var(--accent); }
  .col-warning   { color: var(--warning); }

  .net-rates { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
  .net-rate-card {
    background: var(--surface2);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 10px 14px;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .net-rate-label {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 1px;
    color: var(--muted);
    display: flex;
    align-items: center;
    gap: 5px;
  }
  .net-rate-value { font-size: 18px; font-weight: 700; }
  .net-rate-total { font-size: 10px; color: var(--muted); margin-top: 2px; }

  @media (max-width: 600px) {
    main { grid-template-columns: 1fr; padding: 12px; gap: 12px; }
  }
</style>
</head>
<body>

<header>
  <div class="header-logo">
    </svg>
    ARMBIAN STATS
  </div>
  <div class="header-info">
    <div class="header-pill"><div class="status-dot"></div> <span id="hHostname">--</span></div>
    <div class="header-pill">UP <span id="hUptime">--</span></div>
    <div class="header-pill">CPU <span id="hCpuTotal">--</span></div>
    <div class="header-pill">RAM <span id="hRamPct">--</span></div>
  </div>
  <div class="header-clock" id="hClock">--</div>
</header>

<main>

  <div class="card">
    <div class="card-header">
      <div class="card-title">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M14 14.76V3.5a2.5 2.5 0 0 0-5 0v11.26a4.5 4.5 0 1 0 5 0z"/>
        </svg>
        Temperature
      </div>
      <div class="card-value col-warning" id="tempMain">--</div>
    </div>
    <div class="chart-wrap"><canvas id="chartTemp"></canvas></div>
    <div class="temp-grid" id="tempGrid"></div>
  </div>

  <div class="card">
    <div class="card-header">
      <div class="card-title">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="4" y="4" width="16" height="16" rx="2"/>
          <rect x="9" y="9" width="6" height="6"/>
          <line x1="9" y1="2" x2="9" y2="4"/><line x1="15" y1="2" x2="15" y2="4"/>
          <line x1="9" y1="20" x2="9" y2="22"/><line x1="15" y1="20" x2="15" y2="22"/>
          <line x1="2" y1="9" x2="4" y2="9"/><line x1="2" y1="15" x2="4" y2="15"/>
          <line x1="20" y1="9" x2="22" y2="9"/><line x1="20" y1="15" x2="22" y2="15"/>
        </svg>
        CPU
      </div>
      <div class="card-value col-primary" id="cpuTotal">--</div>
    </div>
    <div class="chart-wrap"><canvas id="chartCPU"></canvas></div>
    <div class="core-grid" id="coreGrid"></div>
  </div>

  <div class="card">
    <div class="card-header">
      <div class="card-title">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M6 19v-3m4 3v-3m4 3v-3m4 3v-3M3 10h18M5 5h14a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7a2 2 0 0 1 2-2z"/>
        </svg>
        Memory
      </div>
      <div class="card-value col-secondary" id="ramPct">--</div>
    </div>
    <div class="chart-wrap"><canvas id="chartRAM"></canvas></div>
    <div class="stats-row">
      <div class="stat-item">
        <div class="stat-label">Used</div>
        <div class="stat-value col-secondary" id="ramUsed">--</div>
      </div>
      <div class="stat-item">
        <div class="stat-label">Free</div>
        <div class="stat-value" id="ramFree">--</div>
      </div>
      <div class="stat-item">
        <div class="stat-label">Total</div>
        <div class="stat-value" id="ramTotal">--</div>
      </div>
      <div class="stat-item">
        <div class="stat-label">Swap</div>
        <div class="stat-value col-accent" id="swapInfo">--</div>
      </div>
    </div>
  </div>

  <div class="card">
    <div class="card-header">
      <div class="card-title">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <line x1="2" y1="12" x2="22" y2="12"/>
          <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>
        </svg>
        Network
      </div>
    </div>
    <div class="chart-wrap"><canvas id="chartNet"></canvas></div>
    <div class="net-rates">
      <div class="net-rate-card">
        <div class="net-rate-label col-secondary">RX / Download</div>
        <div class="net-rate-value col-secondary" id="netRx">--</div>
        <div class="net-rate-total" id="netRxTotal">total: --</div>
      </div>
      <div class="net-rate-card">
        <div class="net-rate-label col-accent">TX / Upload</div>
        <div class="net-rate-value col-accent" id="netTx">--</div>
        <div class="net-rate-total" id="netTxTotal">total: --</div>
      </div>
    </div>
  </div>

  <div class="card card-full">
    <div class="card-header">
      <div class="card-title">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <ellipse cx="12" cy="5" rx="9" ry="3"/>
          <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/>
          <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>
        </svg>
        Disk
      </div>
    </div>
    <div class="progress-wrap" id="diskList"></div>
  </div>

  <div class="card card-full">
    <div class="card-header">
      <div class="card-title">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
          <polyline points="14 2 14 8 20 8"/>
          <line x1="16" y1="13" x2="8" y2="13"/>
          <line x1="16" y1="17" x2="8" y2="17"/>
        </svg>
        Log
      </div>
      <button onclick="clearLog()" style="background:none;border:1px solid var(--border);color:var(--muted);padding:3px 10px;border-radius:4px;cursor:pointer;font-size:10px;font-family:inherit;">clear</button>
    </div>
    <div class="log-box" id="logBox"></div>
  </div>

</main>

<script>
const MAX_POINTS = 60;
const css = name => getComputedStyle(document.documentElement).getPropertyValue(name).trim();

function makeLabels() { return Array(MAX_POINTS).fill(''); }

function buildChart(ctx, datasets, yMax) {
  return new Chart(ctx, {
    type: 'line',
    data: { labels: makeLabels(), datasets },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: { duration: 300 },
      interaction: { mode: 'index', intersect: false },
      plugins: {
        legend: {
          display: datasets.length > 1,
          labels: { color: css('--muted'), boxWidth: 10, font: { size: 10, family: 'monospace' } }
        },
        tooltip: { enabled: false }
      },
      scales: {
        x: { display: false },
        y: {
          min: 0,
          max: yMax || undefined,
          grid: { color: css('--border') + '55' },
          ticks: { color: css('--muted'), font: { size: 10 }, maxTicksLimit: 5 },
          border: { display: false },
        }
      },
      elements: { point: { radius: 0 }, line: { tension: 0.3, borderWidth: 1.5 } },
    }
  });
}

function ds(label, color, fill) {
  return {
    label,
    data: Array(MAX_POINTS).fill(null),
    borderColor: color,
    backgroundColor: color + (fill ? '22' : '00'),
    fill: !!fill,
  };
}

function push(chart, idx, val) {
  const d = chart.data.datasets[idx].data;
  d.push(val);
  if (d.length > MAX_POINTS) d.shift();
  chart.data.labels = makeLabels();
}

const chartTemp = buildChart(document.getElementById('chartTemp').getContext('2d'), [ds('Temp', css('--warning'), true)], 100);
const chartCPU  = buildChart(document.getElementById('chartCPU').getContext('2d'),  [ds('CPU',  css('--primary'), true)], 100);
const chartRAM  = buildChart(document.getElementById('chartRAM').getContext('2d'),  [ds('RAM',  css('--secondary'), true)], 100);
const chartNet  = buildChart(document.getElementById('chartNet').getContext('2d'),  [ds('RX', css('--secondary'), false), ds('TX', css('--accent'), false)]);

function fmtBytes(b) {
  if (b == null) return '--';
  const u = ['B','KB','MB','GB','TB'];
  let i = 0;
  while (b >= 1024 && i < u.length - 1) { b /= 1024; i++; }
  return b.toFixed(i === 0 ? 0 : 1) + ' ' + u[i];
}

function fmtRate(bps) { return bps == null ? '--' : fmtBytes(bps) + '/s'; }

function pctColor(p) {
  if (p < 50) return css('--secondary');
  if (p < 80) return css('--accent');
  return css('--warning');
}

function tempColor(c) {
  if (c < 50) return css('--secondary');
  if (c < 70) return css('--accent');
  return css('--warning');
}

function updateClock() {
  document.getElementById('hClock').textContent = new Date().toLocaleTimeString();
}
setInterval(updateClock, 1000);
updateClock();

const logBox = document.getElementById('logBox');
let logLines = [];

function addLog(msg) {
  const t = new Date().toTimeString().substring(0, 8);
  logLines.push({ t, msg });
  if (logLines.length > 100) logLines.shift();
  logBox.innerHTML = logLines.map(l =>
    '<div class="log-entry"><span class="log-time">' + l.t + '</span><span class="log-msg">' + l.msg + '</span></div>'
  ).join('');
  logBox.scrollTop = logBox.scrollHeight;
}

function clearLog() { logLines = []; logBox.innerHTML = ''; }

let coreCount = 0;

function renderCores(pcts, mhzs) {
  const grid = document.getElementById('coreGrid');
  if (pcts.length !== coreCount) {
    coreCount = pcts.length;
    grid.innerHTML = pcts.map((_, i) =>
      '<div class="core-item">' +
        '<div class="core-name">cpu' + i + '</div>' +
        '<div class="core-pct" id="cp' + i + '">0%</div>' +
        '<div class="core-mhz" id="cm' + i + '">-- MHz</div>' +
        '<div class="core-mini-bar"><div class="core-mini-fill" id="cb' + i + '" style="width:0%"></div></div>' +
      '</div>'
    ).join('');
  }
  pcts.forEach((p, i) => {
    const el = document.getElementById('cp' + i);
    const bar = document.getElementById('cb' + i);
    const mhzEl = document.getElementById('cm' + i);
    if (!el) return;
    el.textContent = p.toFixed(1) + '%';
    el.style.color = pctColor(p);
    bar.style.width = p.toFixed(1) + '%';
    bar.style.background = pctColor(p);
    if (mhzs && mhzs[i] != null) mhzEl.textContent = mhzs[i].toFixed(0) + ' MHz';
  });
}

function renderDisks(disks) {
  document.getElementById('diskList').innerHTML = disks.map(d => {
    const pct = d.percent.toFixed(1);
    const c = pctColor(d.percent);
    return '<div class="progress-row">' +
      '<div class="progress-label" title="' + d.device + ' -> ' + d.mount + '">' + d.mount + '</div>' +
      '<div class="progress-bar-bg"><div class="progress-bar-fill" style="width:' + pct + '%;background:' + c + '"></div></div>' +
      '<div class="progress-pct" style="color:' + c + '">' + pct + '%</div>' +
      '<div class="progress-size">' + fmtBytes(d.used) + ' / ' + fmtBytes(d.total) + '</div>' +
    '</div>';
  }).join('');
}

function renderTemps(temps) {
  const grid = document.getElementById('tempGrid');
  if (!temps || temps.length === 0) {
    grid.innerHTML = '<span style="color:var(--muted);font-size:11px">no thermal zones detected</span>';
    return;
  }
  grid.innerHTML = temps.map(t =>
    '<div class="temp-badge">' +
      '<div class="temp-name">' + t.name + '</div>' +
      '<div class="temp-val" style="color:' + tempColor(t.temp) + '">' + t.temp.toFixed(1) + 'C</div>' +
    '</div>'
  ).join('');
}

let firstData = true;
let retries = 0;

function connect() {
  const es = new EventSource('/api/stream');

  es.onopen = function() {
    if (!firstData) addLog('reconnected');
  };

  es.onmessage = function(e) {
    retries = 0;
    try { handle(JSON.parse(e.data)); }
    catch(err) { console.error('parse error', err); }
  };

  es.onerror = function() {
    es.close();
    retries++;
    const delay = Math.min(1000 * retries, 10000);
    addLog('connection lost, retry in ' + (delay / 1000).toFixed(0) + 's');
    setTimeout(connect, delay);
  };
}

function handle(s) {
  if (firstData) {
    firstData = false;
    addLog('connected  host=' + s.hostname);
  }

  document.getElementById('hHostname').textContent = s.hostname || '--';
  document.getElementById('hUptime').textContent   = s.uptime   || '--';
  document.getElementById('hCpuTotal').textContent = s.cpu_total.toFixed(1) + '%';
  document.getElementById('hRamPct').textContent   = s.ram_percent.toFixed(1) + '%';

  const maxTemp = s.temps && s.temps.length > 0 ? Math.max(...s.temps.map(t => t.temp)) : null;
  if (maxTemp !== null) {
    document.getElementById('tempMain').textContent  = maxTemp.toFixed(1) + 'C';
    document.getElementById('tempMain').style.color  = tempColor(maxTemp);
    push(chartTemp, 0, maxTemp);
  } else {
    document.getElementById('tempMain').textContent = 'N/A';
    push(chartTemp, 0, null);
  }
  chartTemp.update();
  renderTemps(s.temps);

  document.getElementById('cpuTotal').textContent    = s.cpu_total.toFixed(1) + '%';
  document.getElementById('cpuTotal').style.color    = pctColor(s.cpu_total);
  push(chartCPU, 0, s.cpu_total);
  chartCPU.update();
  if (s.cpu_percent && s.cpu_percent.length > 0) renderCores(s.cpu_percent, s.cpu_mhz);

  document.getElementById('ramPct').textContent      = s.ram_percent.toFixed(1) + '%';
  document.getElementById('ramPct').style.color      = pctColor(s.ram_percent);
  document.getElementById('ramUsed').textContent     = fmtBytes(s.ram_used);
  document.getElementById('ramFree').textContent     = fmtBytes(s.ram_total - s.ram_used);
  document.getElementById('ramTotal').textContent    = fmtBytes(s.ram_total);
  document.getElementById('swapInfo').textContent    = s.swap_total > 0
    ? fmtBytes(s.swap_used) + ' / ' + fmtBytes(s.swap_total)
    : 'disabled';
  push(chartRAM, 0, s.ram_percent);
  chartRAM.update();

  document.getElementById('netRx').textContent       = fmtRate(s.network.rx_rate);
  document.getElementById('netTx').textContent       = fmtRate(s.network.tx_rate);
  document.getElementById('netRxTotal').textContent  = 'total: ' + fmtBytes(s.network.rx_bytes);
  document.getElementById('netTxTotal').textContent  = 'total: ' + fmtBytes(s.network.tx_bytes);
  push(chartNet, 0, s.network.rx_rate);
  push(chartNet, 1, s.network.tx_rate);
  chartNet.update();

  if (s.disks && s.disks.length > 0) renderDisks(s.disks);
}

addLog('starting...');
connect();
</script>
</body>
</html>
`
