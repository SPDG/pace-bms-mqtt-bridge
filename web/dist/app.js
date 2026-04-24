const fmt = new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 });

async function loadStatus() {
  const res = await fetch('/api/v1/status');
  const data = await res.json();
  const telemetryById = Object.fromEntries(data.telemetry.map(item => [item.id, item]));

  renderHeader(data);
  renderServices(data.service);
  renderPower(telemetryById, data.packs);
  renderPacks(data.packs);
  renderTelemetry(data.telemetry);
}

function renderHeader(data) {
  document.getElementById('device-name').textContent = data.device.name;
  document.getElementById('summary').textContent =
    `${data.serial.port} · ${data.serial.baudRate} baud · ${data.packs.length} pack(s)`;

  const topStatus = document.getElementById('top-status');
  topStatus.innerHTML = '';
  Object.values(data.service).sort((a, b) => a.name.localeCompare(b.name)).forEach(svc => {
    const pill = document.createElement('span');
    pill.className = `status-pill ${svc.connected ? 'ok' : 'bad'}`;
    pill.textContent = `${svc.name}: ${svc.status}`;
    topStatus.appendChild(pill);
  });
}

function renderServices(services) {
  const servicesEl = document.getElementById('services');
  servicesEl.innerHTML = '';
  Object.values(services).sort((a, b) => a.name.localeCompare(b.name)).forEach(svc => {
    const div = document.createElement('div');
    div.className = 'metric-tile';
    div.innerHTML = `
      <div class="tile-top">
        <span class="tile-label">${svc.name}</span>
        <span class="dot ${svc.connected ? 'ok' : 'bad'}"></span>
      </div>
      <strong class="${svc.connected ? 'ok-text' : 'bad-text'}">${svc.status}</strong>
      ${svc.lastError ? `<small>${svc.lastError}</small>` : '<small>online</small>'}
    `;
    servicesEl.appendChild(div);
  });
}

function renderPower(telemetryById, packs) {
  const overview = document.getElementById('power-overview');
  const batteryPower = telemetryById.battery_power;
  const dischargePower = telemetryById.battery_discharge_power;
  const chargePower = telemetryById.battery_charge_power;
  const powerValue = Number(batteryPower?.value ?? batteryPower?.rendered ?? 0);
  const flow = powerValue > 5 ? 'discharging' : powerValue < -5 ? 'charging' : 'idle';
  const flowLabel = flow === 'discharging' ? 'Discharging' : flow === 'charging' ? 'Charging' : 'Idle';
  const soc = averageSOC(packs);
  overview.innerHTML = batteryPower
    ? `
      <div class="power-tile ${flow}">
        <div class="tile-top">
          <span class="tile-label">Battery Power</span>
          <span class="speed-chip flow-chip">${flowLabel}</span>
        </div>
        <div class="battery-gauge" style="--soc:${soc ?? 0}">
          <span>${soc !== null ? `${fmt.format(soc)}%` : '-'}</span>
        </div>
        <div class="power-value">${batteryPower.rendered}<span>${batteryPower.unit}</span></div>
        <div class="power-split">
          <span>Discharge ${dischargePower?.rendered ?? '0'} W</span>
          <span>Charge ${chargePower?.rendered ?? '0'} W</span>
        </div>
      </div>
    `
    : '';
}

function renderPacks(packs) {
  const packsEl = document.getElementById('packs');
  const subtitle = document.getElementById('pack-subtitle');
  packsEl.innerHTML = '';
  subtitle.textContent = `${packs.length} pack(s) visible`;

  packs.forEach(pack => {
    const cells = pack.cellsMv || [];
    const min = cells.length ? Math.min(...cells) : null;
    const max = cells.length ? Math.max(...cells) : null;
    const diff = min !== null && max !== null ? max - min : null;
    const cellBlocks = cells.map(mv => {
      const pct = min !== null && max !== null && max > min ? (mv - min) / (max - min) : 0.5;
      const level = pct > 0.66 ? 'high' : pct < 0.33 ? 'low' : 'mid';
      return `<span class="cell ${level}" title="${mv} mV">${mv}</span>`;
    }).join('');

    const div = document.createElement('article');
    div.className = 'pack-card';
    div.innerHTML = `
      <div class="pack-head">
        <div>
          <h3>Pack ${String(pack.address).padStart(2, '0')}</h3>
          <p>${cells.length} cells · ${pack.temperaturesC?.length ?? 0} temps</p>
        </div>
        <span class="soc">${fmt.format(pack.soc)}%</span>
      </div>
      <div class="pack-stats">
        <span><strong>${fmt.format(pack.voltageV)}</strong> V</span>
        <span><strong>${fmt.format(pack.currentA)}</strong> A</span>
        <span><strong>${fmt.format(pack.powerKw * 1000)}</strong> W</span>
      </div>
      <div class="cell-map">${cellBlocks}</div>
      <div class="cell-summary">
        <span>Min ${min ?? '-'} mV</span>
        <span>Max ${max ?? '-'} mV</span>
        <span>Diff ${diff ?? '-'} mV</span>
      </div>
    `;
    packsEl.appendChild(div);
  });
}

function averageSOC(packs) {
  const values = packs
    .filter(pack => Number.isFinite(Number(pack.soc)))
    .map(pack => Number(pack.soc));
  if (!values.length) {
    return null;
  }
  return values.reduce((sum, value) => sum + value, 0) / values.length;
}

function renderTelemetry(telemetry) {
  const telemetryEl = document.getElementById('telemetry');
  const subtitle = document.getElementById('telemetry-subtitle');
  telemetryEl.innerHTML = '';
  subtitle.textContent = `${telemetry.length} sensors`;

  telemetry.forEach(item => {
    const tr = document.createElement('tr');
    const updated = item.updatedAt ? new Date(item.updatedAt).toLocaleTimeString() : '';
    tr.innerHTML = `<td>${item.name}</td><td>${item.rendered}${item.unit ? ` ${item.unit}` : ''}</td><td>${updated}</td>`;
    telemetryEl.appendChild(tr);
  });
}

loadStatus().catch(err => {
  document.getElementById('summary').textContent = err.message;
});
setInterval(loadStatus, 5000);
