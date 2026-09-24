(function () {
  'use strict';

  const CATALOG_URL = './data/catalog.json';

  const els = {
    search: document.getElementById('search'),
    categoryFilters: document.getElementById('category-filters'),
    featureFilters: document.getElementById('feature-filters'),
    hosted: document.getElementById('filter-hosted'),
    selfHosted: document.getElementById('filter-selfhosted'),
    openSource: document.getElementById('filter-opensource'),
    matchMode: document.getElementsByName('match-mode'),
    hideIrrelevant: document.getElementById('hide-irrelevant'),
    sort: document.getElementById('sort'),
    reset: document.getElementById('reset-filters'),
    resultCount: document.getElementById('result-count'),
    table: document.getElementById('tools-table'),
    thead: document.getElementById('tools-head'),
    tbody: document.getElementById('tools-body'),
  };

  let catalog = null;
  let hoverState = null;

  const categoryDisplay = (c) =>
    ({ monitoring: 'Monitoring', incident_management: 'Incident mgmt', on_call: 'On-call', status_page: 'Status pages' }[c] || c);

  const groupDisplay = (g) => g.replace(/_/g, ' ');

  function parseState(s) {
    return String(s || 'unknown');
  }

  function stateClass(state) {
    return {
      supported: 'state-supported',
      unsupported: 'state-unsupported',
      unknown: 'state-unknown',
      not_applicable: 'state-na',
    }[state] || 'state-unknown';
  }

  function stateLabel(state) {
    return {
      supported: 'Supported',
      unsupported: 'Unsupported',
      unknown: 'Unknown / unverified',
      not_applicable: 'Not applicable',
    }[state] || state;
  }

  function isStale(dateStr) {
    if (!dateStr) return true;
    const then = new Date(dateStr);
    if (isNaN(then)) return true;
    const yearAgo = new Date();
    yearAgo.setFullYear(yearAgo.getFullYear() - 1);
    return then < yearAgo;
  }

  function formatDate(dateStr) {
    if (!dateStr) return '—';
    const d = new Date(dateStr);
    return isNaN(d) ? dateStr : d.toLocaleDateString();
  }

  function buildCategoryFilters() {
    const container = els.categoryFilters;
    container.innerHTML = '';
    catalog.features
      .map(f => f.group)
      .filter((g, i, arr) => arr.indexOf(g) === i && ['monitoring', 'incident_management', 'on_call', 'status_page'].includes(g))
      .forEach(group => {
        const label = document.createElement('label');
        const input = document.createElement('input');
        input.type = 'checkbox';
        input.value = group;
        input.className = 'category-filter';
        label.appendChild(input);
        label.appendChild(document.createTextNode(' ' + categoryDisplay(group)));
        container.appendChild(label);
      });
  }

  function buildFeatureFilters() {
    const groups = {};
    catalog.features.forEach(f => {
      if (!groups[f.group]) groups[f.group] = [];
      groups[f.group].push(f);
    });

    els.featureFilters.innerHTML = '';
    Object.keys(groups).sort().forEach(group => {
      const div = document.createElement('div');
      div.className = 'feature-group';
      const h3 = document.createElement('h3');
      h3.textContent = groupDisplay(group);
      div.appendChild(h3);
      groups[group].sort((a, b) => a.name.localeCompare(b.name)).forEach(f => {
        const label = document.createElement('label');
        const input = document.createElement('input');
        input.type = 'checkbox';
        input.value = f.id;
        input.className = 'feature-filter';
        input.title = f.description;
        label.appendChild(input);
        label.appendChild(document.createTextNode(' ' + f.name));
        div.appendChild(label);
      });
      els.featureFilters.appendChild(div);
    });
  }

  function getState() {
    const categories = Array.from(document.querySelectorAll('.category-filter:checked')).map(i => i.value);
    const features = Array.from(document.querySelectorAll('.feature-filter:checked')).map(i => i.value);
    const matchMode = Array.from(els.matchMode).find(r => r.checked)?.value || 'all';
    return {
      q: els.search.value.trim().toLowerCase(),
      categories,
      hosted: els.hosted.checked,
      selfHosted: els.selfHosted.checked,
      openSource: els.openSource.checked,
      features,
      matchMode,
      hideIrrelevant: els.hideIrrelevant.checked,
      sort: els.sort.value,
    };
  }

  function setState(state) {
    els.search.value = state.q || '';
    document.querySelectorAll('.category-filter').forEach(i => {
      i.checked = state.categories.includes(i.value);
    });
    document.querySelectorAll('.feature-filter').forEach(i => {
      i.checked = state.features.includes(i.value);
    });
    els.hosted.checked = !!state.hosted;
    els.selfHosted.checked = !!state.selfHosted;
    els.openSource.checked = !!state.openSource;
    Array.from(els.matchMode).forEach(r => {
      r.checked = r.value === (state.matchMode || 'all');
    });
    els.hideIrrelevant.checked = !!state.hideIrrelevant;
    els.sort.value = state.sort || 'name-asc';
  }

  function stateToParams(state) {
    const params = new URLSearchParams();
    if (state.q) params.set('q', state.q);
    if (state.categories.length) params.set('categories', state.categories.join(','));
    if (state.hosted) params.set('hosted', '1');
    if (state.selfHosted) params.set('self_hosted', '1');
    if (state.openSource) params.set('open_source', '1');
    if (state.features.length) params.set('features', state.features.join(','));
    if (state.matchMode !== 'all') params.set('match', state.matchMode);
    if (state.hideIrrelevant) params.set('hide', '1');
    if (state.sort && state.sort !== 'name-asc') params.set('sort', state.sort);
    return params;
  }

  function paramsToState(params) {
    return {
      q: params.get('q') || '',
      categories: (params.get('categories') || '').split(',').filter(Boolean),
      hosted: params.get('hosted') === '1',
      selfHosted: params.get('self_hosted') === '1',
      openSource: params.get('open_source') === '1',
      features: (params.get('features') || '').split(',').filter(Boolean),
      matchMode: params.get('match') || 'all',
      hideIrrelevant: params.get('hide') === '1',
      sort: params.get('sort') || 'name-asc',
    };
  }

  function sanitizeState(state) {
    const knownCategories = new Set(['monitoring', 'incident_management', 'on_call', 'status_page']);
    const knownFeatures = new Set(catalog.features.map(f => f.id));
    const knownSorts = new Set(['name-asc', 'name-desc', 'verified-desc']);
    return {
      q: state.q,
      categories: state.categories.filter(c => knownCategories.has(c)),
      hosted: !!state.hosted,
      selfHosted: !!state.selfHosted,
      openSource: !!state.openSource,
      features: state.features.filter(f => knownFeatures.has(f)),
      matchMode: state.matchMode === 'any' ? 'any' : 'all',
      hideIrrelevant: !!state.hideIrrelevant,
      sort: knownSorts.has(state.sort) ? state.sort : 'name-asc',
    };
  }

  function updateURL(state) {
    const params = stateToParams(state);
    const qs = params.toString();
    const url = qs ? '?' + qs : window.location.pathname;
    if (window.location.search !== '?' + qs && !(window.location.search === '' && qs === '')) {
      window.history.replaceState({}, '', url);
    }
  }

  function filterTools(state) {
    return catalog.tools.filter(tool => {
      if (state.q && !tool.name.toLowerCase().includes(state.q)) return false;

      if (state.categories.length && !state.categories.some(c => tool.categories.includes(c))) return false;

      // Deployment options are alternatives, not requirements: selecting
      // "SaaS" and "Open source" means "either", not "both".
      const deploymentSelected = state.hosted || state.selfHosted || state.openSource;
      if (deploymentSelected) {
        const matchesDeployment =
          (state.hosted && tool.deployment.hosted) ||
          (state.selfHosted && tool.deployment.self_hosted) ||
          (state.openSource && tool.open_source);
        if (!matchesDeployment) return false;
      }

      if (state.features.length) {
        const supported = state.features.filter(fid => {
          const claim = tool.features[fid];
          return claim && claim.value === 'supported';
        });
        if (state.matchMode === 'all') {
          if (supported.length !== state.features.length) return false;
        } else {
          if (supported.length === 0) return false;
        }
      }

      return true;
    });
  }

  function sortTools(tools, sortKey) {
    const sorted = [...tools];
    switch (sortKey) {
      case 'name-asc':
        sorted.sort((a, b) => a.name.localeCompare(b.name));
        break;
      case 'name-desc':
        sorted.sort((a, b) => b.name.localeCompare(a.name));
        break;
      case 'verified-desc':
        sorted.sort((a, b) => {
          const da = a.last_verified || '';
          const db = b.last_verified || '';
          if (da === db) return a.name.localeCompare(b.name);
          return db.localeCompare(da);
        });
        break;
    }
    return sorted;
  }

  function determineColumns(tools, state) {
    const cols = catalog.features.filter(f => {
      if (state.hideIrrelevant) {
        const allNA = tools.every(t => (t.features[f.id]?.value || 'unknown') === 'not_applicable');
        if (allNA) return false;
      }
      return true;
    });
    return cols;
  }

  function renderCell(claim) {
    const state = parseState(claim?.value);
    const span = document.createElement('span');
    span.className = 'state ' + stateClass(state);
    span.setAttribute('aria-label', stateLabel(state));
    span.textContent = state === 'supported' ? 'yes' : state === 'unsupported' ? 'no' : state === 'not_applicable' ? 'n/a' : '?';

    if ((state === 'supported' || state === 'unsupported') && claim.sources && claim.sources.length) {
      const a = document.createElement('a');
      a.href = claim.sources[0];
      a.target = '_blank';
      a.rel = 'noopener noreferrer';
      a.title = claim.notes || stateLabel(state);
      a.appendChild(span);
      return a;
    }
    if (claim?.notes) {
      span.title = claim.notes;
    }
    return span;
  }

  // Small favicon next to the tool name. Falls back to a letter tile when no
  // local asset exists, so a row never renders a broken image.
  function createToolIcon(tool) {
    const fallback = () => {
      const tile = document.createElement('span');
      tile.className = 'tool-icon tool-icon-fallback';
      tile.setAttribute('aria-hidden', 'true');
      tile.textContent = (tool.name || '?').trim().charAt(0).toUpperCase();
      return tile;
    };
    if (!tool.favicon) return fallback();
    const img = document.createElement('img');
    img.className = 'tool-icon';
    img.src = './assets/favicons/' + tool.favicon;
    img.alt = '';
    img.width = 16;
    img.height = 16;
    img.loading = 'lazy';
    img.decoding = 'async';
    img.addEventListener('error', () => img.replaceWith(fallback()));
    return img;
  }

  // Highlight the hovered cell's row and column (crosshair). Sticky header and
  // tool column keep their own classes, so the highlight is applied on top.
  function clearHover() {
    if (!hoverState) return;
    els.table.querySelectorAll('.is-hover-row, .is-hover-col, .is-hover-cell').forEach(el => {
      el.classList.remove('is-hover-row', 'is-hover-col', 'is-hover-cell');
    });
    hoverState = null;
  }

  function applyHover(cell) {
    const row = cell.closest('tr');
    if (!row || !els.table.contains(row)) return;
    const col = cell.cellIndex;
    if (hoverState && hoverState.row === row && hoverState.col === col) return;
    clearHover();
    hoverState = { row, col };
    row.classList.add('is-hover-row');
    cell.classList.add('is-hover-cell');
    els.table.querySelectorAll('tr').forEach(r => {
      const c = r.cells[col];
      if (c) c.classList.add('is-hover-col');
    });
  }

  function renderTable(tools) {
    const state = getState();
    const cols = determineColumns(tools, state);

    // Rows are rebuilt below, so any previous hover targets are gone.
    hoverState = null;
    els.thead.innerHTML = '';
    const headerRow = document.createElement('tr');
    const thTool = document.createElement('th');
    thTool.className = 'col-tool';
    thTool.scope = 'col';
    thTool.textContent = 'Tool';
    headerRow.appendChild(thTool);
    cols.forEach(f => {
      const th = document.createElement('th');
      th.scope = 'col';
      th.textContent = f.name;
      th.title = f.description;
      headerRow.appendChild(th);
    });
    const thVerified = document.createElement('th');
    thVerified.scope = 'col';
    thVerified.textContent = 'Last verified';
    headerRow.appendChild(thVerified);
    els.thead.appendChild(headerRow);

    els.tbody.innerHTML = '';
    tools.forEach(tool => {
      const tr = document.createElement('tr');

      const tdTool = document.createElement('th');
      tdTool.className = 'col-tool';
      tdTool.scope = 'row';
      const head = document.createElement('div');
      head.className = 'tool-head';
      head.appendChild(createToolIcon(tool));
      const nameLink = document.createElement('a');
      nameLink.href = tool.homepage;
      nameLink.className = 'tool-name';
      nameLink.textContent = tool.name;
      nameLink.target = '_blank';
      nameLink.rel = 'noopener noreferrer';
      head.appendChild(nameLink);
      tdTool.appendChild(head);

      const meta = document.createElement('div');
      meta.className = 'tool-meta';
      if (tool.categories && tool.categories.length) {
        tool.categories.forEach(c => {
          const badge = document.createElement('span');
          badge.className = 'badge';
          badge.textContent = categoryDisplay(c);
          meta.appendChild(badge);
        });
      }
      if (tool.status && tool.status !== 'active') {
        const badge = document.createElement('span');
        badge.className = 'badge badge-status';
        badge.textContent = tool.status;
        meta.appendChild(badge);
      }
      if (tool.open_source) {
        const badge = document.createElement('span');
        badge.className = 'badge';
        badge.textContent = tool.license || 'Open source';
        meta.appendChild(badge);
      }
      if (tool.repository) {
        const repoLink = document.createElement('a');
        repoLink.href = tool.repository;
        repoLink.textContent = 'Repo';
        repoLink.target = '_blank';
        repoLink.rel = 'noopener noreferrer';
        meta.appendChild(document.createTextNode(' · '));
        meta.appendChild(repoLink);
      }
      tdTool.appendChild(meta);
      tr.appendChild(tdTool);

      cols.forEach(f => {
        const td = document.createElement('td');
        td.className = 'cell-state';
        td.appendChild(renderCell(tool.features[f.id]));
        tr.appendChild(td);
      });

      const tdVerified = document.createElement('td');
      const dateStr = tool.last_verified;
      tdVerified.textContent = formatDate(dateStr);
      if (isStale(dateStr)) {
        tdVerified.className = 'stale';
        tdVerified.title = 'Verification is more than one year old';
      }
      tr.appendChild(tdVerified);

      els.tbody.appendChild(tr);
    });

    els.resultCount.textContent = `${tools.length} tool${tools.length === 1 ? '' : 's'} match${tools.length === 1 ? 'es' : ''}`;
  }

  function render() {
    const state = getState();
    updateURL(state);
    let tools = filterTools(state);
    tools = sortTools(tools, state.sort);
    renderTable(tools);
  }

  function bindEvents() {
    const inputs = [
      els.search,
      els.hosted,
      els.selfHosted,
      els.openSource,
      els.hideIrrelevant,
      els.sort,
      ...els.matchMode,
    ];
    inputs.forEach(el => el.addEventListener('input', render));

    els.categoryFilters.addEventListener('change', render);
    els.featureFilters.addEventListener('change', render);

    // Crosshair highlight: follows the pointer, and also follows keyboard
    // focus so the row/column context is visible when tabbing into cells.
    els.table.addEventListener('mouseover', event => {
      const cell = event.target.closest('td, th');
      if (cell) applyHover(cell);
    });
    els.table.addEventListener('mouseleave', clearHover);
    els.table.addEventListener('focusin', event => {
      const cell = event.target.closest('td, th');
      if (cell) applyHover(cell);
    });
    els.table.addEventListener('focusout', clearHover);

    els.reset.addEventListener('click', () => {
      setState({ q: '', categories: [], hosted: false, selfHosted: false, openSource: false, features: [], matchMode: 'all', hideIrrelevant: false, sort: 'name-asc' });
      render();
    });
  }

  async function loadCatalog() {
    const embedded = document.getElementById('catalog-data');
    if (embedded && embedded.textContent.trim()) {
      try {
        return JSON.parse(embedded.textContent);
      } catch (err) {
        throw new Error(`embedded catalog is invalid JSON: ${err.message}`);
      }
    }
    const res = await fetch(CATALOG_URL);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json();
  }

  // Compact, derived metadata line for the header. No data is invented: counts
  // come from the catalog and the date is the newest verified_at in the dataset.
  function renderHeaderMeta() {
    const el = document.getElementById('header-meta');
    if (!el || !catalog) return;
    const dates = catalog.tools.map(t => t.last_verified).filter(Boolean).sort();
    const latest = dates.length ? dates[dates.length - 1] : '';
    const parts = [
      `${catalog.tools.length} tools`,
      `${catalog.features.length} features`,
    ];
    if (latest) parts.push(`last verified ${latest}`);
    el.textContent = parts.join('  ·  ');
  }

  async function init() {
    try {
      catalog = await loadCatalog();
    } catch (err) {
      els.resultCount.textContent = `Error loading catalog: ${err.message}`;
      return;
    }

    renderHeaderMeta();
    buildCategoryFilters();
    buildFeatureFilters();

    const params = new URLSearchParams(window.location.search);
    setState(sanitizeState(paramsToState(params)));
    bindEvents();
    render();
  }

  init();
})();
