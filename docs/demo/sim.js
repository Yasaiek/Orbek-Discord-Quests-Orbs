// The live preview on the project site: Orbek's own panel (ui/index.html), with this script answering its
// requests instead of the program. Everything here is made up.
(() => {
  const qs = new URLSearchParams(location.search);
  let lang = qs.get('lang') || 'en';
  const now = Date.now(), H = 3600e3;
  const rw = (type, name, orbs) => ({ type, name: name || '', orbs: orbs || 0, img: '' });
  const Q = [
    ['d1', 'Neon Drift – watch the trailer', 'Neon Drift', 'Lumen Games', 'WATCH_VIDEO', 60, 0, [rw(4, '', 700)], 30],
    ['d2', 'Play Skyforge Arena', 'Skyforge Arena', 'Ironpeak', 'PLAY_ON_DESKTOP', 900, 360, [rw(3, 'Avatar decoration: Ember Crown')], 96],
    ['d3', 'Play Pocket Chess together', 'Pocket Chess', 'Discord', 'PLAY_ACTIVITY', 600, 0, [rw(4, '', 450)], 8],
    ['d4', 'Stream Frontier Ops', 'Frontier Ops', 'Northwind', 'STREAM_ON_DESKTOP', 900, 0, [rw(2, 'Weapon crate'), rw(4, '', 300)], 200],
    ['d5', 'Moonfall – watch the reveal', 'Moonfall', 'Starlight', 'WATCH_VIDEO', 90, 90, [rw(3, 'Profile effect: Aurora')], 50, true],
    ['d6', 'Play Rally Kings', 'Rally Kings', 'Gearbox Lane', 'PLAY_ON_DESKTOP', 900, 0, [rw(4, '', 1000)], 140],
    ['d7', 'Tide Tactics – watch the gameplay video', 'Tide Tactics', 'Harbor', 'WATCH_VIDEO', 120, 0, [rw(4, '', 500)], 70],
    ['d8', 'Play Garden Golf', 'Garden Golf', 'Discord', 'PLAY_ACTIVITY', 900, 0, [rw(3, 'Avatar decoration: Clover')], 20],
  ].map(([id, name, game, publisher, task, target, progress, rewards, hours, done]) => ({
    id, name, game, publisher, task, target, progress, rewards, hero: '', startsAt: now - 40 * H, expiresAt: now + hours * H,
    enrolledAt: progress ? now - 2 * H : 0, completedAt: done ? now - H : 0, claimedAt: 0, auto: true,
    orbs: rewards.reduce((n, r) => n + (r.type === 4 ? r.orbs : 0), 0), cosmetic: rewards.some((r) => r.type === 3),
  }));
  const S = {
    app: 'Orbek', version: window.ORBEK_VERSION || '0.0.0', lang, winGen: 0, connected: true, lastAt: now, user: { id: '1', name: 'Demo', avatar: '' },
    loaded: true, err: '', orbs: 1240, runs: [], queue: [], desktop: true, plugin: '0.0.0', pluginLatest: '0.0.0', vencord: 'C:\\Users\\you\\Documents\\Vencord',
    quests: Q, log: [{ t: now, l: 'ok', m: 'Orbek started.' }],
    settings: { lang: '', autoMode: true, autoAll: false, autoOnly: 'all', openOnStart: true, autostart: false, parallel: 0, autoUpdate: true },
    setup: { running: false, done: false, err: '', steps: [], lines: [], full: true },
    update: { state: 'current', latest: '', checked: now, doneAt: 0 },
  };
  S.plugin = S.pluginLatest = S.version;
  const find = (id) => S.quests.find((q) => q.id === id);
  const say = (l, m) => S.log.push({ t: Date.now(), l, m });
  const start = () => {
    while (S.queue.length) {
      const q = find(S.queue.shift());
      if (!q || q.completedAt || S.runs.some((r) => r.id === q.id)) continue;
      if (!q.enrolledAt) q.enrolledAt = Date.now();
      S.runs.push({ id: q.id, progress: q.progress, progressAt: Date.now(), note: q.task === 'STREAM_ON_DESKTOP' ? 'run.streamHint' : '', since: Date.now() });
      say('info', q.name + ': …');
    }
  };
  setInterval(() => { // the demo runs ~20× faster than Discord
    for (const r of S.runs.slice()) {
      const q = find(r.id); const step = q.task.startsWith('WATCH') ? 7 : 20;
      q.progress = Math.min(q.target, q.progress + step); r.progress = q.progress; r.progressAt = Date.now();
      if (q.progress >= q.target) { q.completedAt = Date.now(); S.runs = S.runs.filter((x) => x !== r); say('ok', q.name + ' ✓'); }
    }
  }, 1000);
  const setupDemo = (full) => {
    const keys = full ? ['node', 'vencord', 'plugin', 'build', 'inject', 'enable'] : ['node', 'plugin', 'build', 'enable'];
    S.setup = { running: true, done: false, err: '', full, steps: keys.map((key) => ({ key, state: 'wait' })), lines: [] };
    let i = 0;
    const next = () => {
      if (i > 0) S.setup.steps[i - 1].state = 'ok';
      if (i === keys.length) { S.setup.running = false; S.setup.done = true; return; }
      S.setup.steps[i].state = 'run'; S.setup.lines.push('» ' + keys[i] + '…'); i++; setTimeout(next, 1100);
    };
    next();
  };
  const op = (o) => {
    const q = o.id && find(o.id);
    switch (o.op) {
      case 'enroll': if (q) { q.enrolledAt = Date.now(); say('ok', q.name); } break;
      case 'run': for (const id of o.ids || []) if (!S.queue.includes(id)) S.queue.push(id); start(); break;
      case 'stop': S.queue = S.queue.filter((x) => x !== o.id); S.runs = S.runs.filter((r) => r.id !== o.id); break;
      case 'stopAll': S.queue = []; S.runs = []; break;
      case 'claim': if (q) q.claimedAt = Date.now(); break;
      case 'settings': Object.assign(S.settings, o.patch || {}); if ('lang' in (o.patch || {})) S.lang = o.patch.lang || lang; if (!S.settings.autoMode) { S.queue = []; S.runs = []; } break;
      case 'setup': setupDemo(true); break;
      case 'pluginInstall': setupDemo(false); break;
      case 'clearLog': S.log = []; break;
    }
    return { ok: true };
  };
  const json = (v) => new Response(JSON.stringify(v), { headers: { 'Content-Type': 'application/json' } });
  const real = window.fetch.bind(window);
  window.fetch = (url, init) => {
    const u = String(url);
    if (u.includes('/api/state')) return Promise.resolve(json(S));
    if (u.includes('/api/op')) return Promise.resolve(json(op(JSON.parse((init && init.body) || '{}'))));
    return real(url, init);
  };
  window.confirm = () => true;
  addEventListener('message', (e) => { if (e.data && e.data.orbekLang) { lang = S.lang = e.data.orbekLang; } });
})();
