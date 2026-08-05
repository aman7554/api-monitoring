document.addEventListener('DOMContentLoaded', () => {
  const API_BASE = '/api/v1';
  let jwtToken = null;
  let chartInstance = null;
  let historyData = {
    labels: [],
    datasets: []
  };

  // UI Elements
  const monitorsListEl = document.getElementById('monitors-list');
  const totalMonitorsEl = document.getElementById('total-monitors-count');
  const upMonitorsEl = document.getElementById('up-monitors-count');
  const uptimeValEl = document.getElementById('uptime-val');
  const avgLatencyEl = document.getElementById('avg-latency-val');
  const activeIncidentsEl = document.getElementById('active-incidents-val');
  const totalChecksEl = document.getElementById('total-checks-val');
  const monitorBadgeCountEl = document.getElementById('monitor-badge-count');
  const logStreamEl = document.getElementById('log-stream');
  
  // Modal & Button Elements
  const modalAdd = document.getElementById('modal-add-monitor');
  const btnAddMonitor = document.getElementById('btn-add-monitor');
  const btnCloseModal = document.getElementById('btn-close-modal');
  const btnCancelModal = document.getElementById('btn-cancel-modal');
  const formCreateMonitor = document.getElementById('form-create-monitor');
  const btnRefresh = document.getElementById('btn-refresh');

  // Initialize Chart.js
  initChart();

  // Initial Fetch & Login
  autoLoginAndLoad();

  // Set 3-second live polling interval
  setInterval(fetchLiveStatus, 3000);

  // Event Listeners
  btnAddMonitor.addEventListener('click', () => modalAdd.classList.add('active'));
  btnCloseModal.addEventListener('click', () => modalAdd.classList.remove('active'));
  btnCancelModal.addEventListener('click', () => modalAdd.classList.remove('active'));
  btnRefresh.addEventListener('click', fetchLiveStatus);

  formCreateMonitor.addEventListener('submit', async (e) => {
    e.preventDefault();
    await createNewMonitor();
  });

  async function autoLoginAndLoad() {
    try {
      const res = await fetch(`${API_BASE}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: 'admin@pulsewatch.io',
          password: 'Password123!'
        })
      });
      if (res.ok) {
        const data = await res.json();
        jwtToken = data.access_token;
        addLog('success', 'Authenticated successfully as System Admin (JWT Token acquired)');
      }
    } catch (err) {
      addLog('error', `Login failed: ${err.message}`);
    }
    await fetchLiveStatus();
  }

  async function fetchLiveStatus() {
    try {
      const res = await fetch(`${API_BASE}/public/status/acme-prod-status`);
      if (!res.ok) throw new Error(`HTTP error ${res.status}`);
      
      const data = await res.json();
      renderStatusOverview(data);
    } catch (err) {
      addLog('error', `Error polling live status: ${err.message}`);
    }
  }

  function renderStatusOverview(data) {
    const monitors = data.monitors || [];
    const activeIncidents = data.active_incidents || [];

    // Summary Counts
    const totalCount = monitors.length;
    const upCount = monitors.filter(m => m.status === 'up').length;
    const avgLatency = 45; // ms avg

    totalMonitorsEl.textContent = `${totalCount} Monitors`;
    upMonitorsEl.textContent = `${upCount} Operational`;
    monitorBadgeCountEl.textContent = `${totalCount} Active`;
    uptimeValEl.textContent = `${data.uptime_90d || 100}%`;
    avgLatencyEl.textContent = `${avgLatency} ms`;
    activeIncidentsEl.textContent = activeIncidents.length;

    let totalSuccesses = 0;
    monitors.forEach(m => totalSuccesses += (m.consecutive_successes || 0));
    totalChecksEl.textContent = totalSuccesses;

    // Render Table
    if (monitors.length === 0) {
      monitorsListEl.innerHTML = `<tr><td colspan="8" class="text-center py-4 text-muted">No monitors found.</td></tr>`;
      return;
    }

    monitorsListEl.innerHTML = monitors.map(m => {
      const lastCheck = m.last_check_at ? new Date(m.last_check_at).toLocaleTimeString() : 'Never';
      const statusBadge = m.status === 'up' 
        ? `<span class="badge up"><span class="pulse-dot green"></span> UP</span>`
        : `<span class="badge down">DOWN</span>`;
      
      return `
        <tr>
          <td>${statusBadge}</td>
          <td><strong>${escapeHtml(m.name)}</strong></td>
          <td><span class="type-pill">${m.type}</span></td>
          <td><span class="url-code">${escapeHtml(m.url)}</span></td>
          <td>Every ${m.interval_seconds}s</td>
          <td><strong class="text-success">${m.consecutive_successes || 0}</strong></td>
          <td>${lastCheck}</td>
          <td>
            <button class="btn btn-secondary text-sm btn-test-check" data-id="${m.id}" style="padding: 0.25rem 0.6rem; font-size: 0.75rem;">
              ⚡ Test
            </button>
          </td>
        </tr>
      `;
    }).join('');

    // Update Chart with new time slice
    updateChartData(monitors);
  }

  async function createNewMonitor() {
    if (!jwtToken) {
      alert('Authentication token not available. Please wait...');
      return;
    }

    const name = document.getElementById('input-name').value;
    const type = document.getElementById('input-type').value;
    const method = document.getElementById('input-method').value;
    const url = document.getElementById('input-url').value;
    const interval = parseInt(document.getElementById('input-interval').value);
    const expectedCode = parseInt(document.getElementById('input-expected-code').value);

    try {
      const res = await fetch(`${API_BASE}/monitors`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${jwtToken}`
        },
        body: JSON.stringify({
          project_id: 'd0000000-0000-0000-0000-000000000001',
          name,
          type,
          method,
          url,
          interval_seconds: interval,
          expected_status_code: expectedCode
        })
      });

      if (!res.ok) {
        const errData = await res.json();
        throw new Error(errData.error || 'Failed to create monitor');
      }

      addLog('success', `Created new live monitor '${name}' (${url}) polling every ${interval}s`);
      modalAdd.classList.remove('active');
      formCreateMonitor.reset();
      await fetchLiveStatus();
    } catch (err) {
      alert(`Error: ${err.message}`);
      addLog('error', `Failed to create monitor: ${err.message}`);
    }
  }

  function initChart() {
    const ctx = document.getElementById('latencyChart').getContext('2d');
    chartInstance = new Chart(ctx, {
      type: 'line',
      data: {
        labels: ['10s ago', '8s ago', '6s ago', '4s ago', '2s ago', 'Just now'],
        datasets: [
          {
            label: 'Google Search Gateway (http)',
            data: [42, 45, 52, 48, 44, 46],
            borderColor: '#3b82f6',
            backgroundColor: 'rgba(59, 130, 246, 0.1)',
            tension: 0.3,
            fill: true
          },
          {
            label: 'Cloudflare DNS (dns)',
            data: [12, 14, 11, 15, 10, 12],
            borderColor: '#10b981',
            backgroundColor: 'rgba(16, 185, 129, 0.1)',
            tension: 0.3,
            fill: true
          },
          {
            label: 'GitHub SSL (ssl)',
            data: [120, 115, 125, 118, 122, 120],
            borderColor: '#f59e0b',
            backgroundColor: 'rgba(245, 158, 11, 0.1)',
            tension: 0.3,
            fill: true
          }
        ]
      },
      options: {
        responsive: true,
        plugins: {
          legend: { labels: { color: '#9ca3af', font: { family: 'Inter' } } }
        },
        scales: {
          x: { ticks: { color: '#6b7280' }, grid: { color: '#1f2937' } },
          y: { ticks: { color: '#6b7280' }, grid: { color: '#1f2937' }, title: { display: true, text: 'Latency (ms)', color: '#9ca3af' } }
        }
      }
    });
  }

  function updateChartData(monitors) {
    if (!chartInstance) return;
    const nowLabel = new Date().toLocaleTimeString();
    
    // Push new time label
    if (chartInstance.data.labels.length > 8) {
      chartInstance.data.labels.shift();
    }
    chartInstance.data.labels.push(nowLabel);

    // Update dataset points
    chartInstance.data.datasets.forEach((dataset, idx) => {
      if (dataset.data.length > 8) {
        dataset.data.shift();
      }
      const randomVariance = Math.floor(Math.random() * 6) - 3;
      const base = idx === 0 ? 45 : idx === 1 ? 12 : 120;
      dataset.data.push(Math.max(5, base + randomVariance));
    });

    chartInstance.update();
  }

  function addLog(type, msg) {
    const entry = document.createElement('div');
    entry.className = `log-entry ${type}`;
    const time = new Date().toLocaleTimeString();
    entry.innerHTML = `<span class="log-time">[${time}]</span> <span class="log-msg">${escapeHtml(msg)}</span>`;
    logStreamEl.prepend(entry);
  }

  function escapeHtml(str) {
    return str ? str.replace(/[&<>"']/g, m => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#039;' })[m]) : '';
  }
});
