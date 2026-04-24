const fmt = new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 });
let latestStatus = null;
let haYamlMode = localStorage.getItem('haYamlMode') || 'dashboard';

async function loadStatus() {
  const res = await fetch('/api/v1/status');
  const data = await res.json();
  latestStatus = data;
  const telemetryById = Object.fromEntries(data.telemetry.map(item => [item.id, item]));

  renderHeader(data);
  renderServices(data.service);
  renderPower(telemetryById, data.packs);
  renderPacks(data.packs);
  renderTelemetry(data.telemetry);
  renderHAConfig(data);
  renderSettings(data);
}

function initNavigation() {
  const rail = document.getElementById('rail');
  const expanded = localStorage.getItem('railExpanded') === 'true';
  document.body.classList.toggle('rail-expanded', expanded);
  rail.classList.toggle('expanded', expanded);

  document.getElementById('rail-toggle').addEventListener('click', () => {
    const next = !document.body.classList.contains('rail-expanded');
    document.body.classList.toggle('rail-expanded', next);
    rail.classList.toggle('expanded', next);
    localStorage.setItem('railExpanded', String(next));
  });

  document.querySelectorAll('[data-tab]').forEach(button => {
    button.addEventListener('click', () => activateTab(button.dataset.tab));
  });

  document.querySelectorAll('[data-copy-target]').forEach(button => {
    button.addEventListener('click', async () => {
      const target = document.getElementById(button.dataset.copyTarget);
      await copyText(target.textContent);
      const previous = button.textContent;
      button.textContent = 'Copied';
      window.setTimeout(() => {
        button.textContent = previous;
      }, 1200);
    });
  });

  document.querySelectorAll('input[name="ha-yaml-mode"]').forEach(input => {
    input.checked = input.value === haYamlMode;
    input.addEventListener('change', () => {
      haYamlMode = input.value;
      localStorage.setItem('haYamlMode', haYamlMode);
      if (latestStatus) {
        renderHAConfig(latestStatus);
      }
    });
  });
}

function activateTab(tab) {
  document.querySelectorAll('[data-tab]').forEach(button => {
    button.classList.toggle('active', button.dataset.tab === tab);
  });
  document.querySelectorAll('.tab-view').forEach(view => {
    view.classList.toggle('active', view.id === `view-${tab}`);
  });
  if (latestStatus) {
    renderHAConfig(latestStatus);
    renderSettings(latestStatus);
  }
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

async function copyText(text) {
  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text);
      return;
    } catch {
      // Fall back to a temporary textarea for non-secure HTTP contexts.
    }
  }
  const textarea = document.createElement('textarea');
  textarea.value = text;
  textarea.style.position = 'fixed';
  textarea.style.opacity = '0';
  document.body.appendChild(textarea);
  textarea.focus();
  textarea.select();
  document.execCommand('copy');
  textarea.remove();
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

function renderHAConfig(data) {
  const target = document.getElementById('ha-yaml');
  if (!target) {
    return;
  }
  target.textContent = generateDashboardYAML(data, haYamlMode);
}

function renderSettings(data) {
  setDefinitionList('settings-runtime', [
    ['Version', data.build?.version ?? '-'],
    ['Commit', data.build?.commit ?? '-'],
    ['Build date', data.build?.buildDate ?? '-'],
    ['Config path', data.runtime?.configPath ?? '-'],
    ['Uptime', data.runtime?.uptime ?? '-'],
    ['HTTP listen', data.http?.listen ?? '-'],
    ['Log level', data.logging?.level ?? '-'],
  ]);
  setDefinitionList('settings-serial', [
    ['Port', data.serial?.port ?? '-'],
    ['Protocol', data.device?.protocol ?? '-'],
    ['Baud rate', data.serial?.baudRate ?? '-'],
    ['Frame', `${data.serial?.dataBits ?? '-'}${data.serial?.parity ?? '-'}${data.serial?.stopBits ?? '-'}`],
    ['Timeout', data.serial?.timeout ?? '-'],
    ['Poll interval', data.polling?.interval ?? '-'],
    ['Reconnect delay', data.polling?.reconnectDelay ?? '-'],
  ]);
  setDefinitionList('settings-mqtt', [
    ['Broker', data.mqtt?.broker ?? '-'],
    ['Username', data.mqtt?.username || '-'],
    ['Password', data.mqtt?.passwordConfigured ? 'configured' : 'empty'],
    ['Client ID', data.mqtt?.clientId ?? '-'],
    ['Topic prefix', data.mqtt?.topicPrefix ?? '-'],
    ['Discovery prefix', data.mqtt?.discoveryPrefix ?? '-'],
    ['Retain', data.mqtt?.retain ? 'true' : 'false'],
  ]);
  setDefinitionList('settings-device', [
    ['Name', data.device?.name ?? '-'],
    ['Manufacturer', data.device?.manufacturer ?? '-'],
    ['Model', data.device?.model ?? '-'],
    ['First pack address', data.device?.firstPackAddress ?? '-'],
    ['Max parallel packs', data.device?.maxParallelPacks ?? '-'],
    ['Discover on startup', data.device?.discoverOnStartup ? 'true' : 'false'],
    ['Visible packs', data.packs?.length ?? 0],
  ]);
}

function setDefinitionList(id, rows) {
  const el = document.getElementById(id);
  if (!el) {
    return;
  }
  el.innerHTML = rows.map(([label, value]) => `
    <dt>${escapeHTML(label)}</dt>
    <dd>${escapeHTML(String(value))}</dd>
  `).join('');
}

function generateDashboardYAML(data, mode) {
  const view = generateSectionsViewYAML(data);
  if (mode === 'tab') {
    return `${view.join('\n')}\n`;
  }
  return `title: PACE BMS\nviews:\n${viewAsListItem(view)}\n`;
}

function generateSectionsViewYAML(data) {
  const packs = data.packs || [];
  const telemetry = data.telemetry || [];
  const entity = id => `sensor.${sanitizeEntity(data.device?.name || 'pace_main')}_${sanitizeEntity(id)}`;
  const existing = new Set(telemetry.map(item => item.id));
  const lines = [
    'type: sections',
    'title: PACE BMS',
    'path: pace-bms',
    'icon: mdi:battery',
    'max_columns: 4',
    'sections:',
    '  - type: grid',
    '    cards:',
    '      - type: heading',
    '        heading: Battery system',
    '      - type: tile',
    `        entity: ${entity('battery_power')}`,
    '        name: Battery Power',
    '        vertical: false',
    '      - type: tile',
    `        entity: ${entity('battery_discharge_power')}`,
    '        name: Discharge Power',
    '        vertical: false',
    '      - type: tile',
    `        entity: ${entity('battery_charge_power')}`,
    '        name: Charge Power',
    '        vertical: false',
  ];

  const historyEntities = ['battery_power'];
  for (const pack of packs) {
    historyEntities.push(`pack_${pad2(pack.address)}_soc`);
  }
  lines.push(
    '      - type: history-graph',
    '        title: Power and SOC',
    '        hours_to_show: 24',
    '        entities:',
    ...historyEntities.map(id => `          - entity: ${entity(id)}`),
    '  - type: grid',
    '    cards:',
    '      - type: heading',
    '        heading: Battery packs'
  );

  if (!packs.length) {
    lines.push(
      '      - type: markdown',
      '        content: >',
      '          No packs were visible when this YAML was generated.'
    );
  }

  for (const pack of packs) {
    const prefix = `pack_${pad2(pack.address)}`;
    const entities = [
      `${prefix}_soc`,
      `${prefix}_voltage`,
      `${prefix}_current`,
      `${prefix}_power`,
      `${prefix}_remaining_capacity`,
      `${prefix}_full_capacity`,
      `${prefix}_cell_voltage_min`,
      `${prefix}_cell_voltage_max`,
      `${prefix}_cell_voltage_diff`,
      ...Array.from({ length: pack.temperaturesC?.length ?? 0 }, (_, index) => `${prefix}_temperature_${pad2(index + 1)}`),
    ].filter(id => existing.has(id));
    lines.push(
      '      - type: entities',
      `        title: Pack ${pad2(pack.address)}`,
      '        show_header_toggle: false',
      '        entities:',
      ...entities.map(id => `          - entity: ${entity(id)}`)
    );
  }

  lines.push(
    '  - type: grid',
    '    column_span: 2',
    '    cards:',
    '      - type: heading',
    '        heading: Cell voltage spread'
  );
  if (!packs.length) {
    lines.push(
      '      - type: markdown',
      '        content: >',
      '          No cell voltage entities were visible when this YAML was generated.'
    );
  }
  for (const pack of packs) {
    const prefix = `pack_${pad2(pack.address)}`;
    const cellEntities = Array.from({ length: pack.cellsMv?.length ?? 0 }, (_, index) => `${prefix}_cell_${pad2(index + 1)}_voltage`)
      .filter(id => existing.has(id));
    lines.push(
      '      - type: entities',
      `        title: Pack ${pad2(pack.address)} cells`,
      '        show_header_toggle: false',
      '        entities:',
      ...cellEntities.map(id => `          - entity: ${entity(id)}`)
    );
  }

  return lines;
}

function viewAsListItem(lines) {
  return lines.map((line, index) => {
    if (index === 0) {
      return `  - ${line}`;
    }
    return `    ${line}`;
  }).join('\n');
}

function sanitizeEntity(value) {
  return String(value)
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '_')
    .replace(/^_+|_+$/g, '');
}

function pad2(value) {
  return String(value).padStart(2, '0');
}

function escapeHTML(value) {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;');
}

initNavigation();
loadStatus().catch(err => {
  document.getElementById('summary').textContent = err.message;
});
setInterval(loadStatus, 5000);
