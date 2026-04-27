/* global window.__telescope injected by the Go template */
const _cfg     = window.__telescope || {};
const HTTP_API  = _cfg.httpAPI  || '/telescope/api/requests';
const CACHE_API = _cfg.cacheAPI || '/telescope/api/cache';

let httpEntries  = [];
let cacheEntries = [];
let activeTab    = 'http';
let selectedId   = null;
let lastFetch    = 0;
let paused       = false;

const httpBody     = document.getElementById('http-body');
const cacheBody    = document.getElementById('cache-body');
const detailPanel  = document.getElementById('detail-panel');
const detailBody   = document.getElementById('detail-body');
const detailTitle  = document.getElementById('detail-title');
const countBadge   = document.getElementById('count-badge');
const statusDot    = document.getElementById('status-dot');
const httpPill     = document.getElementById('http-pill');
const cachePill    = document.getElementById('cache-pill');
const resizeHandle = document.getElementById('resize-handle');
const mainEl       = document.getElementById('main');
const pauseBtn     = document.getElementById('pause-btn');
const iconPause    = document.getElementById('icon-pause');
const iconPlay     = document.getElementById('icon-play');

// ── pause / resume ────────────────────────────────────────────
pauseBtn.addEventListener('click', () => {
  paused = !paused;
  iconPause.style.display = paused ? 'none' : '';
  iconPlay.style.display  = paused ? ''     : 'none';
  pauseBtn.classList.toggle('active', paused);
  pauseBtn.title = paused ? 'Resume auto-refresh' : 'Pause auto-refresh';
  if (paused) {
    statusDot.classList.add('paused');
    statusDot.classList.remove('stale');
    statusDot.title = 'Paused';
  } else {
    statusDot.classList.remove('paused');
    statusDot.title = 'Live';
    fetchAll();
  }
});

// ── resize ────────────────────────────────────────────────────
let isResizing   = false;
let resizeStartX = 0;
let resizeStartW = 0;
const MIN_DETAIL_W = 200;

resizeHandle.addEventListener('mousedown', e => {
  isResizing   = true;
  resizeStartX = e.clientX;
  resizeStartW = detailPanel.offsetWidth;
  resizeHandle.classList.add('dragging');
  document.body.style.cursor     = 'col-resize';
  document.body.style.userSelect = 'none';
  e.preventDefault();
});

document.addEventListener('mousemove', e => {
  if (!isResizing) return;
  const newW = Math.min(
    Math.max(MIN_DETAIL_W, resizeStartW + (resizeStartX - e.clientX)),
    mainEl.offsetWidth - 300
  );
  detailPanel.style.flexBasis = newW + 'px';
  detailPanel.style.width     = newW + 'px';
});

document.addEventListener('mouseup', () => {
  if (!isResizing) return;
  isResizing = false;
  resizeHandle.classList.remove('dragging');
  document.body.style.cursor     = '';
  document.body.style.userSelect = '';
});

// ── detail panel ─────────────────────────────────────────────
function openDetail() {
  detailPanel.classList.remove('hidden');
  resizeHandle.classList.remove('hidden');
}

function closeDetail() {
  if (selectedId) {
    const prev = document.querySelector(`tbody tr[data-id="${CSS.escape(selectedId)}"]`);
    if (prev) prev.classList.remove('selected');
  }
  selectedId = null;
  detailPanel.classList.add('hidden');
  resizeHandle.classList.add('hidden');
}

function setSelected(tbody, id) {
  tbody.querySelectorAll('tr.selected').forEach(r => r.classList.remove('selected'));
  if (id) {
    const tr = tbody.querySelector(`tr[data-id="${CSS.escape(id)}"]`);
    if (tr) tr.classList.add('selected');
  }
}

// ── helpers ───────────────────────────────────────────────────
function methodClass(m) {
  const k = (m || '').toUpperCase();
  return ['GET','POST','PUT','PATCH','DELETE','HEAD'].includes(k) ? k : 'OTHER';
}

function statusClass(code) {
  if (code < 0)   return 'status-neg';
  if (code < 200) return 'status-1xx';
  if (code < 300) return 'status-2xx';
  if (code < 400) return 'status-3xx';
  if (code < 500) return 'status-4xx';
  return 'status-5xx';
}

const WRITE_CMDS  = new Set(['SET','SETEX','PSETEX','SETNX','MSET','MSETNX','HMSET','HSET',
  'LPUSH','RPUSH','SADD','ZADD','INCR','DECR','INCRBY','DECRBY','APPEND','GETSET','SETRANGE']);
const DELETE_CMDS = new Set(['DEL','UNLINK','HDEL','LREM','SREM','ZREM',
  'EXPIRE','PERSIST','PEXPIRE','EXPIREAT','PEXPIREAT']);

function cmdClass(cmd, errStr) {
  if (errStr) return 'err';
  const k = (cmd || '').toUpperCase();
  if (WRITE_CMDS.has(k))  return 'write';
  if (DELETE_CMDS.has(k)) return 'delete';
  return '';
}

function durClass(ms) {
  if (ms > 2000) return 'slow';
  if (ms > 500)  return 'medium';
  return '';
}

function fmtTime(ts) {
  const d = new Date(ts);
  const p = n => String(n).padStart(2, '0');
  return `${p(d.getHours())}:${p(d.getMinutes())}:${p(d.getSeconds())}.${String(d.getMilliseconds()).padStart(3, '0')}`;
}

function fmtCode(body) {
  if (!body) return '(empty)';
  try { return JSON.stringify(JSON.parse(body), null, 2); }
  catch (_) { return body; }
}

function escHtml(s) {
  return String(s || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}

function detailRow(label, valueHtml) {
  return `<div class="detail-row">
    <span class="detail-label">${label}</span>
    <span class="detail-value">${valueHtml}</span>
  </div>`;
}

function detailCode(label, text, extraClass = '') {
  return `<div class="detail-row">
    <span class="detail-label">${label}</span>
    <pre class="detail-code ${extraClass}">${escHtml(fmtCode(text))}</pre>
  </div>`;
}

// ── tabs ──────────────────────────────────────────────────────
document.querySelectorAll('.tab-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    activeTab = btn.dataset.tab;
    document.querySelectorAll('.tab-btn').forEach(b => b.classList.toggle('active', b === btn));
    document.getElementById('tab-http').style.display  = activeTab === 'http'  ? '' : 'none';
    document.getElementById('tab-cache').style.display = activeTab === 'cache' ? '' : 'none';
    closeDetail();
    updateBadge();
  });
});

function updateBadge() {
  const n = activeTab === 'http' ? httpEntries.length : cacheEntries.length;
  countBadge.textContent = `${n} ${activeTab === 'http' ? 'request' : 'command'}${n !== 1 ? 's' : ''}`;
}

// ── HTTP table ────────────────────────────────────────────────
function renderHttp() {
  const frag = document.createDocumentFragment();
  httpEntries.forEach(e => {
    const tr = document.createElement('tr');
    tr.dataset.id = e.id;
    if (e.id === selectedId) tr.classList.add('selected');
    tr.innerHTML = `
      <td class="ts">${fmtTime(e.timestamp)}</td>
      <td><span class="method-badge method-${methodClass(e.method)}">${escHtml(e.method || '?')}</span></td>
      <td class="url-cell" title="${escHtml(e.url)}">${escHtml(e.url)}</td>
      <td><span class="status-badge ${statusClass(e.status_code)}">${e.status_code < 0 ? 'ERR' : e.status_code}</span></td>
      <td class="dur ${durClass(e.duration_ms)}">${e.duration_ms} ms</td>
    `;
    tr.addEventListener('click', () => selectHttpEntry(e.id));
    frag.appendChild(tr);
  });
  httpBody.innerHTML = '';
  httpBody.appendChild(frag);
  const n = httpEntries.length;
  httpPill.textContent = n;
  document.getElementById('http-table').style.display = n === 0 ? 'none' : '';
  document.getElementById('http-empty').style.display = n === 0 ? 'flex'  : 'none';
}

function selectHttpEntry(id) {
  selectedId = id;
  setSelected(httpBody, id);
  const e = httpEntries.find(x => x.id === id);
  if (!e) { closeDetail(); return; }

  detailTitle.textContent = 'HTTP Request';
  detailBody.innerHTML =
    detailRow('ID',        escHtml(e.id)) +
    detailRow('Timestamp', new Date(e.timestamp).toISOString()) +
    detailRow('Method',    `<span class="method-badge method-${methodClass(e.method)}">${escHtml(e.method)}</span>`) +
    detailRow('URL',       `<span class="url-cell">${escHtml(e.url)}</span>`) +
    detailRow('Status',    `<span class="status-badge ${statusClass(e.status_code)}">${e.status_code < 0 ? 'Error' : e.status_code}</span>`) +
    detailRow('Duration',  `<span class="dur ${durClass(e.duration_ms)}">${e.duration_ms} ms</span>`) +
    detailCode('Request Body',  e.body) +
    detailCode('Response Body', e.response_body, 'response');
  openDetail();
}

// ── Cache table ───────────────────────────────────────────────
function renderCache() {
  const frag = document.createDocumentFragment();
  cacheEntries.forEach(e => {
    const tr = document.createElement('tr');
    tr.dataset.id = e.id;
    if (e.id === selectedId) tr.classList.add('selected');
    const cc = cmdClass(e.command, e.error);
    tr.innerHTML = `
      <td class="ts">${fmtTime(e.timestamp)}</td>
      <td><span class="cmd-badge ${cc}">${escHtml(e.command || '?')}</span></td>
      <td class="key-cell" title="${escHtml(e.key)}">${escHtml(e.key)}</td>
      <td class="dur ${durClass(e.duration_ms)}">${e.duration_ms} ms</td>
      <td class="err-cell">${escHtml(e.error)}</td>
    `;
    tr.addEventListener('click', () => selectCacheEntry(e.id));
    frag.appendChild(tr);
  });
  cacheBody.innerHTML = '';
  cacheBody.appendChild(frag);
  const n = cacheEntries.length;
  cachePill.textContent = n;
  document.getElementById('cache-table').style.display = n === 0 ? 'none' : '';
  document.getElementById('cache-empty').style.display = n === 0 ? 'flex'  : 'none';
}

function selectCacheEntry(id) {
  selectedId = id;
  setSelected(cacheBody, id);
  const e = cacheEntries.find(x => x.id === id);
  if (!e) { closeDetail(); return; }

  const cc = cmdClass(e.command, e.error);
  detailTitle.textContent = 'Cache Command';
  detailBody.innerHTML =
    detailRow('ID',        escHtml(e.id)) +
    detailRow('Timestamp', new Date(e.timestamp).toISOString()) +
    detailRow('Command',   `<span class="cmd-badge ${cc}">${escHtml(e.command)}</span>`) +
    detailRow('Key',       `<span class="key-cell">${escHtml(e.key)}</span>`) +
    detailRow('Duration',  `<span class="dur ${durClass(e.duration_ms)}">${e.duration_ms} ms</span>`) +
    (e.error ? detailRow('Error', `<span class="err-cell">${escHtml(e.error)}</span>`) : '');
  openDetail();
}

// ── data fetching ─────────────────────────────────────────────
async function fetchAll() {
  if (paused) return;
  try {
    const [rHttp, rCache] = await Promise.all([fetch(HTTP_API), fetch(CACHE_API)]);
    if (!rHttp.ok || !rCache.ok) throw new Error('non-200');
    httpEntries  = await rHttp.json()  || [];
    cacheEntries = await rCache.json() || [];
    lastFetch = Date.now();
    statusDot.classList.remove('stale');
    renderHttp();
    renderCache();
    updateBadge();
    if (selectedId && activeTab === 'http'  && !httpEntries.find(e => e.id === selectedId))  closeDetail();
    if (selectedId && activeTab === 'cache' && !cacheEntries.find(e => e.id === selectedId)) closeDetail();
  } catch (_) {
    if (Date.now() - lastFetch > 6000) statusDot.classList.add('stale');
  }
}

document.getElementById('clear-btn').addEventListener('click', async () => {
  const api = activeTab === 'http' ? HTTP_API : CACHE_API;
  await fetch(api, { method: 'DELETE' });
  if (activeTab === 'http')  { httpEntries  = []; renderHttp(); }
  if (activeTab === 'cache') { cacheEntries = []; renderCache(); }
  closeDetail();
  updateBadge();
});

document.getElementById('close-detail').addEventListener('click', closeDetail);

fetchAll();
setInterval(fetchAll, 2000);
