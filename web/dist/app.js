async function loadStatus() {
  const res = await fetch('/api/v1/status');
  const data = await res.json();
  document.getElementById('summary').textContent =
    `${data.device.name} on ${data.serial.port} at ${data.serial.baudRate} baud, ${data.packs.length} pack(s) visible`;

  const services = document.getElementById('services');
  services.innerHTML = '';
  Object.values(data.service).sort((a, b) => a.name.localeCompare(b.name)).forEach(svc => {
    const div = document.createElement('div');
    div.className = 'card';
    div.innerHTML = `<strong>${svc.name}</strong><span class="${svc.connected ? 'ok' : 'bad'}">${svc.status}</span>${svc.lastError ? `<br><small>${svc.lastError}</small>` : ''}`;
    services.appendChild(div);
  });

  const packs = document.getElementById('packs');
  packs.innerHTML = '';
  data.packs.forEach(pack => {
    const div = document.createElement('div');
    div.className = 'card';
    const cells = pack.cellsMv || [];
    const min = cells.length ? Math.min(...cells) / 1000 : null;
    const max = cells.length ? Math.max(...cells) / 1000 : null;
    const diff = min !== null && max !== null ? max - min : null;
    const cellSummary = diff !== null
      ? `<div class="cell-summary"><span>Min ${Math.round(min * 1000)} mV</span><span>Max ${Math.round(max * 1000)} mV</span><span>Diff ${Math.round(diff * 1000)} mV</span></div>`
      : '';
    div.innerHTML = `<strong>Pack ${String(pack.address).padStart(2, '0')}</strong>${pack.voltageV} V<br>${pack.currentA} A<br>${pack.soc} % SOC<br>${cells.length} cells${cellSummary}`;
    packs.appendChild(div);
  });

  const telemetry = document.getElementById('telemetry');
  telemetry.innerHTML = '';
  data.telemetry.forEach(item => {
    const tr = document.createElement('tr');
    const updated = item.updatedAt ? new Date(item.updatedAt).toLocaleString() : '';
    tr.innerHTML = `<td>${item.name}</td><td>${item.rendered}${item.unit ? ` ${item.unit}` : ''}</td><td>${updated}</td>`;
    telemetry.appendChild(tr);
  });
}

loadStatus().catch(err => {
  document.getElementById('summary').textContent = err.message;
});
setInterval(loadStatus, 5000);
