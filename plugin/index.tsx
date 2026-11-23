/*
 * Orbek by Sówka – wtyczka do Discorda dla programu Orbek
 *
 *  Sam plugin nie ma okna. Łączy Discorda z programem Orbek (osobne okno na tym komputerze):
 *    - co kilka sekund wysyła do Orbka listę zadań Discorda (z nagrodami i postępem),
 *    - odbiera to, co klikniesz w Orbku (zapisz się, wykonaj, zatrzymaj, odbierz) i robi to w Discordzie.
 *  Do Discorda idą te same zapytania, które wysyła sam klient przy zadaniach. Token konta nigdzie nie wychodzi.
 *
 *  Automat (udawanie gry / streamu / filmu) działa tylko, gdy włączysz go w ustawieniach Orbka.
 *  Discord tego zabrania – może odebrać nagrody albo zablokować konto.
 *
 *  Vencord Toolbox (prawy górny róg) → „Otwórz Orbka”.
 */

import definePlugin, { PluginNative } from "@utils/types";
import { findStore } from "@webpack";
import { ChannelStore, FluxDispatcher, GuildChannelStore, NavigationRouter, RestAPI, UserStore } from "@webpack/common";

const Native = VencordNative.pluginHelpers.Orbek as PluginNative<typeof import("./native")>;

const PLUGIN_VERSION = "0.2.1";
const TAG = "[Orbek]";

// ---------- zadania ----------

const VIDEO = ["WATCH_VIDEO", "WATCH_VIDEO_ON_MOBILE"];
const TASK_ORDER = ["WATCH_VIDEO", "PLAY_ON_DESKTOP", "STREAM_ON_DESKTOP", "PLAY_ACTIVITY", "WATCH_VIDEO_ON_MOBILE"];

interface Reward { type: number; name: string; orbs: number; img: string; }
interface Quest {
    id: string; name: string; game: string; publisher: string; appId: string;
    task: string; target: number; progress: number; configVersion: number;
    rewards: Reward[]; hero: string;
    startsAt: number; expiresAt: number; enrolledAt: number; completedAt: number; claimedAt: number;
    auto: boolean;
}

const CDN = "https://cdn.discordapp.com/";

function asset(questId: string, a: any): string {
    if (typeof a !== "string" || !a) return "";
    if (/^https?:/.test(a)) return a;
    if (a.startsWith("quests/")) return CDN + a;
    return `${CDN}quests/${questId}/${a}`;
}

// API daje snake_case, pamięć Discorda camelCase – wszystko sprowadzamy do camelCase (klucze typu WATCH_VIDEO zostają)
function camel(v: any): any {
    if (Array.isArray(v)) return v.map(camel);
    if (!v || typeof v !== "object" || v instanceof Date) return v;
    const out: any = {};
    for (const [k, val] of Object.entries(v)) {
        const key = /^[a-z0-9]+(_[a-z0-9]+)+$/.test(k) ? k.replace(/_([a-z0-9])/g, (_, c) => c.toUpperCase()) : k;
        out[key] = camel(val);
    }
    return out;
}

const time = (s: any) => {
    const t = s ? new Date(s).getTime() : 0;
    return isNaN(t) ? 0 : t;
};

function normalize(raw: any): Quest | null {
    const q = camel(raw);
    if (!q?.id || !q.config) return null;
    const c = q.config;
    const tasks = c.taskConfigV2?.tasks ?? c.taskConfig?.tasks ?? {};
    const task = TASK_ORDER.find(t => tasks[t]) ?? Object.keys(tasks)[0] ?? "";
    const us = q.userStatus ?? {};
    let progress = Number(us.progress?.[task]?.value ?? 0);
    if (!progress && c.configVersion === 1) progress = Number(us.streamProgressSeconds ?? 0);
    const target = Number(tasks[task]?.target ?? 0);
    return {
        id: q.id,
        name: c.messages?.questName ?? c.messages?.gameTitle ?? "Quest",
        game: c.messages?.gameTitle ?? c.application?.name ?? "",
        publisher: c.messages?.gamePublisher ?? "",
        appId: c.application?.id ?? "",
        task, target, progress,
        configVersion: Number(c.configVersion ?? 2),
        rewards: (c.rewardsConfig?.rewards ?? []).map((r: any) => ({
            type: Number(r.type),
            name: r.messages?.name ?? r.messages?.nameWithArticle ?? "",
            orbs: Number(r.orbQuantity ?? 0),
            img: asset(q.id, r.asset)
        })),
        hero: asset(q.id, c.assets?.hero ?? c.assets?.questBarHero ?? c.assets?.gameTile),
        startsAt: time(c.startsAt),
        expiresAt: time(c.expiresAt),
        enrolledAt: time(us.enrolledAt),
        completedAt: time(us.completedAt),
        claimedAt: time(us.claimedAt),
        auto: TASK_ORDER.includes(task) && target > 0
    };
}

const questsStore = () => {
    try { return findStore("QuestsStore") as any; } catch { return null; }
};

function fromStore(): Quest[] {
    const m = questsStore()?.quests;
    const list: any[] = m instanceof Map ? [...m.values()] : m ? Object.values(m) : [];
    return list.map(normalize).filter(Boolean) as Quest[];
}

const known = new Map<string, Quest>();
let loaded = false, loadErr = "", orbs: number | null = null, lastLoad = 0;

// Obrazek Orbów, którego Discord używa na swojej stronie zadań (statyczny plik discord.com/assets/…).
// Szukamy go w tym, co Discord ma akurat na ekranie: <img> obok napisu „… Orbs”.
let orbIcon = "", orbLook = 0;
function findOrbIcon() {
    if (orbIcon || Date.now() - orbLook < 5000) return;
    orbLook = Date.now();
    try {
        for (const img of Array.from(document.querySelectorAll("img"))) {
            const src = img.currentSrc || img.src;
            if (!src || !/\/assets\/[^?#]+\.(svg|png|webp|gif)/i.test(src) || img.naturalWidth < 24) continue;
            const alt = `${img.alt ?? ""} ${img.getAttribute("aria-label") ?? ""}`;
            let hit = /\borbs?\b/i.test(alt);
            for (let n: HTMLElement | null = img.parentElement, i = 0; !hit && n && i < 4; n = n.parentElement, i++)
                hit = n.children.length <= 6 && /\d[\d\s.,]*\s*Orbs?\b/i.test(n.textContent ?? "") && (n.textContent ?? "").length < 220;
            if (hit) { orbIcon = new URL(src, location.href).href; return; }
        }
    } catch { }
}

async function load() {
    lastLoad = Date.now();
    for (const q of fromStore()) known.set(q.id, q);
    try {
        const res = await RestAPI.get({ url: "/quests/@me" });
        const list = res?.body?.quests ?? (Array.isArray(res?.body) ? res.body : null);
        if (list) {
            known.clear();
            for (const raw of list) {
                const q = normalize(raw);
                if (q) known.set(q.id, q);
            }
            loaded = true;
            loadErr = "";
        }
    } catch (e: any) {
        loadErr = e?.body?.message ?? e?.message ?? String(e);
        console.warn(TAG, "lista zadań:", e);
        if (known.size) loaded = true;
    }
    try {
        const res = await RestAPI.get({ url: "/users/@me/virtual-currency/balance" });
        if (typeof res?.body?.balance === "number") orbs = res.body.balance;
    } catch { /* saldo Orbów jest tylko dodatkiem */ }
}

// postęp gier liczy sam Discord – dociągamy go z jego pamięci
function mergeLive() {
    for (const q of fromStore()) {
        const old = known.get(q.id);
        if (!old || q.progress > old.progress || (q.completedAt && !old.completedAt) || (q.claimedAt && !old.claimedAt) || (q.enrolledAt && !old.enrolledAt))
            known.set(q.id, q);
    }
}

// ---------- automat ----------

interface Run { id: string; task: string; progress: number; progressAt: number; note: string; since: number; stop: () => void; }
const runs = new Map<string, Run>();
const queue: string[] = [];
let autoAllowed = false, maxParallel = 1; // 0 = bez limitu
const events: any[] = [];
const say = (k: string, q: Quest, msg = "") => events.push({ k, id: q.id, name: q.name, msg });

const sleep = (ms: number) => new Promise(r => setTimeout(r, ms));
const isDesktop = () => typeof (window as any).DiscordNative !== "undefined";
const errText = (e: any) => e?.body?.message ?? e?.message ?? String(e);

async function enroll(q: Quest): Promise<Quest> {
    if (q.enrolledAt) return q;
    const res = await RestAPI.post({ url: `/quests/${q.id}/enroll`, body: { location: 11, is_targeted: false, metadata_raw: null } });
    const us = camel(res?.body ?? {});
    const out = { ...q, enrolledAt: time(us.enrolledAt) || Date.now() };
    known.set(q.id, out);
    return out;
}

function setProgress(run: Run, id: string, p: number) {
    if (p !== run.progress) run.progressAt = Date.now();
    run.progress = p;
    const q = known.get(id);
    if (q && p > q.progress) known.set(id, { ...q, progress: p });
}

// film: postęp kawałkami, nie szybciej niż od zapisu do zadania minęło czasu (+10 s)
async function runVideo(q: Quest, run: Run, stopped: () => boolean) {
    const speed = 7, ahead = 10;
    let done = q.progress, completed = false;
    while (!stopped() && done < q.target) {
        const allowed = Math.floor((Date.now() - q.enrolledAt) / 1000) + ahead;
        const next = done + speed;
        if (allowed - done >= speed) {
            const res = await RestAPI.post({ url: `/quests/${q.id}/video-progress`, body: { timestamp: Math.min(q.target, next + Math.random()) } });
            completed = res?.body?.completed_at != null;
            done = Math.min(q.target, next);
            setProgress(run, q.id, done);
        }
        await sleep(1000);
    }
    if (!stopped() && !completed) await RestAPI.post({ url: `/quests/${q.id}/video-progress`, body: { timestamp: q.target } });
}

// Udawane gry: kilka naraz. Discord widzi je wszystkie jako włączone i sam wysyła heartbeaty dla każdej.
const fakes = new Map<string, any>();
let realGames: any = null, realForPid: any = null;

function setFake(questId: string, fake: any | null) {
    const games = findStore("RunningGameStore") as any;
    const old = fakes.get(questId);
    if (fake) fakes.set(questId, fake); else fakes.delete(questId);
    if (fakes.size && !realGames) {
        realGames = games.getRunningGames;
        realForPid = games.getGameForPID;
        games.getRunningGames = () => [...fakes.values()];
        games.getGameForPID = (p: number) => [...fakes.values()].find(f => f.pid === p);
    }
    const all = [...fakes.values()];
    FluxDispatcher.dispatch({ type: "RUNNING_GAMES_CHANGE", removed: old && !fake ? [old] : [], added: fake ? [fake] : [], games: all } as any);
    if (!fakes.size && realGames) {
        games.getRunningGames = realGames;
        games.getGameForPID = realForPid;
        realGames = realForPid = null;
        FluxDispatcher.dispatch({ type: "RUNNING_GAMES_CHANGE", removed: [], added: [], games: games.getRunningGames() } as any);
    }
}

async function runDesktop(q: Quest, run: Run, stopped: () => boolean, stream: boolean) {
    if (!isDesktop()) throw new Error("tylko w aplikacji Discord na komputerze");
    const streaming = findStore("ApplicationStreamingStore") as any;
    const res = await RestAPI.get({ url: `/applications/public?application_ids=${q.appId}` });
    const app = res?.body?.[0];
    if (!app) throw new Error("Discord nie zna tej gry");
    const exe = (app.executables?.find((x: any) => x.os === "win32")?.name ?? `${app.name}.exe`).replace(">", "");
    const name = String(app.name).replace(/[\\/:*?"<>|]/g, "");
    const pid = Math.floor(Math.random() * 30000) + 1000;
    const fake = {
        cmdLine: `C:\\Program Files\\${name}\\${exe}`, exeName: exe, exePath: `c:/program files/${name.toLowerCase()}/${exe}`,
        hidden: false, isLauncher: false, id: q.appId, name: app.name, pid, pidPath: [pid], processName: app.name, start: Date.now()
    };
    const realStream = streaming?.getStreamerActiveStreamMetadata;
    if (stream) {
        run.note = "run.streamHint";
        streaming.getStreamerActiveStreamMetadata = () => ({ id: q.appId, pid, sourceName: null });
    } else {
        setFake(q.id, fake);
    }
    let done = false;
    const onBeat = (data: any) => {
        const qid = data?.questId ?? data?.userStatus?.questId ?? data?.user_status?.quest_id;
        if (qid && qid !== q.id) return;
        const us = camel(data?.userStatus ?? data?.user_status ?? {});
        const p = q.configVersion === 1 ? Number(us.streamProgressSeconds ?? 0) : Number(us.progress?.[q.task]?.value ?? 0);
        if (p) setProgress(run, q.id, p);
        if (p >= q.target || us.completedAt) done = true;
    };
    FluxDispatcher.subscribe("QUESTS_SEND_HEARTBEAT_SUCCESS", onBeat);
    try {
        while (!stopped() && !done) {
            await sleep(1000);
            const live = known.get(q.id); // Discord może zaktualizować postęp też we własnej pamięci
            if (live && live.progress > run.progress) setProgress(run, q.id, live.progress);
            if (live?.completedAt || run.progress >= q.target) done = true;
        }
    } finally {
        FluxDispatcher.unsubscribe("QUESTS_SEND_HEARTBEAT_SUCCESS", onBeat);
        if (stream) streaming.getStreamerActiveStreamMetadata = realStream;
        else setFake(q.id, null);
    }
}

// aktywność (gra w Discordzie): heartbeat co 20 s
async function runActivity(q: Quest, run: Run, stopped: () => boolean) {
    const dm = ChannelStore.getSortedPrivateChannels()[0]?.id;
    const vc = (Object.values(GuildChannelStore.getAllGuilds()) as any[]).find(g => g?.VOCAL?.length)?.VOCAL[0]?.channel?.id;
    const channelId = dm ?? vc;
    if (!channelId) throw new Error("potrzebny jest jakikolwiek kanał (DM albo głosowy)");
    const key = `call:${channelId}:1`;
    while (!stopped()) {
        const res = await RestAPI.post({ url: `/quests/${q.id}/heartbeat`, body: { stream_key: key, terminal: false } });
        const p = Number(res?.body?.progress?.PLAY_ACTIVITY?.value ?? 0);
        setProgress(run, q.id, p);
        if (p >= q.target) {
            await RestAPI.post({ url: `/quests/${q.id}/heartbeat`, body: { stream_key: key, terminal: true } });
            return;
        }
        for (let i = 0; i < 20 && !stopped(); i++) await sleep(1000);
    }
}

async function runOne(q0: Quest) {
    let stop = false;
    const run: Run = { id: q0.id, task: q0.task, progress: q0.progress, progressAt: Date.now(), note: "", since: Date.now(), stop: () => { stop = true; } };
    runs.set(q0.id, run);
    say("start", q0);
    try {
        const q = await enroll(q0);
        if (!q0.enrolledAt) await sleep(3000);
        if (VIDEO.includes(q.task)) await runVideo(q, run, () => stop);
        else if (q.task === "PLAY_ON_DESKTOP") await runDesktop(q, run, () => stop, false);
        else if (q.task === "STREAM_ON_DESKTOP") await runDesktop(q, run, () => stop, true);
        else if (q.task === "PLAY_ACTIVITY") await runActivity(q, run, () => stop);
        else throw new Error("tego rodzaju zadania plugin nie umie wykonać");
        if (stop) {
            say("stop", q);
        } else {
            known.set(q.id, { ...(known.get(q.id) ?? q), progress: q.target, completedAt: Date.now() });
            say("done", q);
            setTimeout(load, 5000); // potwierdź na liście Discorda
        }
    } catch (e: any) {
        console.error(TAG, q0.name, e);
        say(stop ? "stop" : "fail", q0, errText(e));
    } finally {
        runs.delete(q0.id);
        pump();
    }
}

// uruchamia z kolejki tyle zadań naraz, ile pozwala ustawienie (stream zawsze jeden)
function pump() {
    if (!autoAllowed) return;
    for (let i = 0; i < queue.length;) {
        if (maxParallel > 0 && runs.size >= maxParallel) return;
        const q = known.get(queue[i]);
        if (!q || q.completedAt || (q.expiresAt && q.expiresAt <= Date.now()) || runs.has(q.id)) {
            queue.splice(i, 1);
            continue;
        }
        if (q.task === "STREAM_ON_DESKTOP" && [...runs.values()].some(r => r.task === "STREAM_ON_DESKTOP")) {
            i++;
            continue;
        }
        queue.splice(i, 1);
        runOne(q);
    }
}

function stopAll() {
    queue.length = 0;
    for (const r of runs.values()) r.stop();
}

// ---------- połączenie z Orbkiem ----------

async function doCmd(c: any) {
    const q = c.id ? known.get(c.id) : undefined;
    switch (c.op) {
        case "refresh":
            await load();
            break;
        case "enroll":
            if (!q) break;
            try { await enroll(q); say("enrolled", q); } catch (e) { say("enrollFail", q, errText(e)); }
            break;
        case "run":
            if (!autoAllowed) break;
            for (const id of c.ids ?? []) if (!queue.includes(id) && !runs.has(id)) queue.push(id);
            pump();
            break;
        case "stop": {
            const i = queue.indexOf(c.id);
            if (i >= 0) queue.splice(i, 1);
            else runs.get(c.id)?.stop();
            break;
        }
        case "stopAll":
            stopAll();
            break;
        case "claim":
            try { NavigationRouter.transitionTo("/quest-home"); } catch (e) { console.error(TAG, e); }
            Native.focusDiscord();
            break;
    }
}

let timer: ReturnType<typeof setTimeout> | undefined;
let alive = false, lastOk = 0;

async function tick() {
    if (!alive) return;
    try {
        if (!lastLoad || Date.now() - lastLoad > 10 * 60_000) await load();
        mergeLive();
        findOrbIcon();
        const me = UserStore.getCurrentUser();
        const body = JSON.stringify({
            v: 1,
            plugin: PLUGIN_VERSION,
            user: me ? { id: me.id, name: (me as any).globalName ?? me.username, avatar: me.getAvatarURL?.(undefined, 64) ?? "" } : null,
            quests: [...known.values()],
            loaded, err: loadErr, orbs,
            runs: [...runs.values()].map(r => ({ id: r.id, progress: r.progress, progressAt: r.progressAt, note: r.note, since: r.since })),
            queue: [...queue],
            events: events.splice(0),
            desktop: isDesktop(),
            orbIcon
        });
        const raw = await Native.sync(body);
        if (raw) {
            lastOk = Date.now();
            const res = JSON.parse(raw);
            autoAllowed = !!res.auto;
            maxParallel = Math.max(0, Number(res.parallel ?? 1));
            if (!autoAllowed && (runs.size || queue.length)) stopAll();
            for (const c of res.cmds ?? []) await doCmd(c);
        } else if (Date.now() - lastOk > 30_000 && (runs.size || queue.length)) {
            stopAll(); // Orbek zniknął – nic nie robimy na ślepo
        }
    } catch (e) {
        console.warn(TAG, e);
    }
    pump();
    if (alive) timer = setTimeout(tick, Date.now() - lastOk > 15_000 ? 8000 : runs.size ? 1000 : 2000);
}

export default definePlugin({
    name: "Orbek",
    description: "Łączy Discorda z programem Orbek: zadania (Questy) i Orby w osobnym oknie (by Sówka)",
    authors: [{ name: "Sówka", id: 0n }],
    dependencies: ["VencordToolbox"],

    toolboxActions: {
        "Otwórz Orbka": () => { Native.openOrbek(); }
    },

    start() {
        alive = true;
        timer = setTimeout(tick, 3000);
    },

    stop() {
        alive = false;
        if (timer) clearTimeout(timer);
        stopAll();
    }
});
