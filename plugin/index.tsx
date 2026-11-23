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
import { findStore, wreq } from "@webpack";
import { ChannelStore, FluxDispatcher, GuildChannelStore, NavigationRouter, RestAPI, UserStore } from "@webpack/common";

const Native = VencordNative.pluginHelpers.Orbek as PluginNative<typeof import("./native")>;

const PLUGIN_VERSION = "0.3.3";
const TAG = "[Orbek]";

// ---------- zadania ----------

const VIDEO = ["WATCH_VIDEO", "WATCH_VIDEO_ON_MOBILE"];
const TASK_ORDER = ["WATCH_VIDEO", "PLAY_ON_DESKTOP", "STREAM_ON_DESKTOP", "PLAY_ACTIVITY", "WATCH_VIDEO_ON_MOBILE"];

interface Reward { type: number; name: string; orbs: number; img: string; }
interface Quest {
    id: string; name: string; game: string; publisher: string; appId: string;
    task: string; target: number; progress: number; configVersion: number;
    rewards: Reward[]; hero: string; icon: string;
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
        icon: asset(q.id, c.assets?.gameTile ?? c.assets?.gameTileDark ?? c.assets?.logotype)
            || (c.application?.icon && c.application?.id ? `${CDN}app-icons/${c.application.id}/${c.application.icon}.png?size=64` : ""),
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

// Obrazki Orbów z samej strony zadań Discorda: duży kafelek nagrody (kolorowy) i mały znaczek (biały).
// Szukamy ich w tym, co Discord ma na ekranie, gdy otwarta jest strona zadań: od napisu z „Orbs” w górę
// drzewa, najbliższy obrazek (img, video albo tło). Kandydatów zapisujemy też do orbDebug – Orbek trzyma
// ich w data/orb-debug.json, żeby dało się sprawdzić, co Discord pokazuje.
let orbTile = "", orbGlyph = "", orbLook = 0, orbDebug: any[] = [], orbDebugSent = "";
const orbTiles: Record<string, string> = {}; // ilość Orbów → kafelek, jaki Discord przy niej pokazuje
function mediaOf(e: Element): string {
    if (e instanceof HTMLImageElement) return e.currentSrc || e.src;
    if (e instanceof HTMLVideoElement) return e.currentSrc || e.src || e.poster || (e.querySelector("source") as HTMLSourceElement | null)?.src || "";
    const m = /url\(["']?([^"')]+)/.exec(getComputedStyle(e).backgroundImage || "");
    return m ? m[1] : "";
}
function findOrbIcon() {
    if (Date.now() - orbLook < (orbTile ? 15000 : 4000)) return;
    orbLook = Date.now();
    if (!/quest/i.test(location.pathname)) return;
    try {
        const cands = new Map<string, { src: string; w: number; h: number; tag: string; lvl: number; txt: string; }>();
        const walk = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT);
        let n = 0;
        for (let t = walk.nextNode(); t && n < 400; t = walk.nextNode()) {
            if (!/\bOrbs?\b/i.test(t.nodeValue ?? "")) continue;
            n++;
            let el: Element | null = t.parentElement;
            for (let lvl = 0; el && lvl < 8; lvl++, el = el.parentElement) {
                const media = [el, ...Array.from(el.querySelectorAll("img,video,[style*='background'],[class*='image'],[class*='Image'],[class*='tile'],[class*='Tile'],[class*='icon'],[class*='Icon']"))];
                for (const m of media.slice(0, 60)) {
                    const src = mediaOf(m);
                    if (!src || src.startsWith("data:") || /\/quests\/\d+\/\d+\.(jpg|jpeg)/i.test(src)) continue;
                    const r = m.getBoundingClientRect();
                    if (r.width < 8) continue;
                    const abs = new URL(src, location.href).href;
                    const old = cands.get(abs);
                    if (!old || old.lvl > lvl) cands.set(abs, { src: abs, w: Math.round(r.width), h: Math.round(r.height), tag: m.tagName.toLowerCase(), lvl, txt: (t.nodeValue ?? "").trim().slice(0, 60) });
                }
            }
        }
        const list = [...cands.values()].sort((x, y) => x.lvl - y.lvl || y.w - x.w);
        orbDebug = list.slice(0, 25);
        const big = list.filter(c => c.w >= 36 && c.w <= 200 && Math.abs(c.w - c.h) <= 12 && !/\/avatars\/|\/icons\/|\/app-icons\//.test(c.src));
        const small = list.filter(c => c.w < 36 && /\/assets\//.test(c.src));
        if (big.length) orbTile = big[0].src;
        for (const c of big) {
            const m = /(\d[\d\s.,]*)\s*Orbs?/i.exec(c.txt);
            const n = m ? Number(m[1].replace(/[^\d]/g, "")) : 0;
            if (n > 0 && c.lvl <= 3 && /\/assets\//.test(c.src) && !orbTiles[n]) orbTiles[n] = c.src;
        }
        if (!orbGlyph && small.length) orbGlyph = small[0].src;
    } catch (e) { console.error(TAG, "orb icon", e); }
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
const say = (k: string, q?: Quest, msg = "") => events.push({ k, id: q?.id ?? "", name: q?.name ?? "", msg });

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

// ---------- odbieranie nagród (eksperymentalne) ----------
// To samo zapytanie, które wysyła przycisk „Odbierz nagrodę” w Discordzie. Jeśli Discord zażąda captchy,
// pokazuje ją sam (jak przy kliknięciu) albo zapytanie wraca z błędem – wtedy otwieramy stronę zadań, żeby
// odebrać ręcznie. Captchy nikt tu nie rozwiązuje za człowieka.
let autoClaim = false;
let refreshMs = 10 * 60_000; // co ile czytać listę zadań od nowa (Orbek to ustawia)
const claimTried = new Map<string, number>();
let claiming = false;

// Jak dokładnie Discord odbiera nagrodę, plugin uczy się z jednego odebrania ręcznie w Discordzie: zapytania
// Discorda o „/quests/…/claim…” są podglądane (adres i treść) i zapamiętywane jako wzór. Bez wzoru – próba
// z domyślną treścią. Wszystko (wzór, ostatni błąd) idzie do Orbka do data/claim-debug.json.
type ClaimTpl = { url: string; body: any; at: number; ok?: boolean; status?: number; err?: string; };
let testing = false, lastClaimEv: { k: string; q?: Quest; msg: string; } = { k: "claimFail", msg: "" };
let claimTpl: ClaimTpl | null = (() => { try { return JSON.parse(localStorage.getItem("orbek.claimTpl") ?? "null"); } catch { return null; } })();
let claimDebug: any = { tpl: claimTpl }, claimDebugSent = "";
let origPost: any = null, ownPost = false;
function watchClaims() {
    if (origPost) return;
    origPost = RestAPI.post;
    (RestAPI as any).post = function (opts: any, ...rest: any[]) {
        const p = origPost.call(this, opts, ...rest);
        const url = String(opts?.url ?? "");
        const m = /^(.*\/quests\/)(\d+)(\/[^?]*claim[^?]*)/i.exec(url);
        if (m && !ownPost) {
            const tpl: ClaimTpl = { url: m[1] + "{id}" + m[3], body: JSON.parse(JSON.stringify(opts?.body ?? null)), at: Date.now() };
            Promise.resolve(p).then(
                (r: any) => { tpl.ok = true; tpl.status = r?.status; claimTpl = tpl; try { localStorage.setItem("orbek.claimTpl", JSON.stringify(tpl)); } catch { } claimDebug = { ...claimDebug, tpl, seen: tpl }; },
                (e: any) => { tpl.ok = false; tpl.status = e?.status; tpl.err = errText(e); claimDebug = { ...claimDebug, seen: tpl }; }
            );
        }
        return p;
    };
}
// Funkcja Discorda, którą woła przycisk „Odbierz nagrodę” (sama dokłada zapieczętowane dane zadania i decyzji).
// Szukana w załadowanym kodzie Discorda po tym, że wysyła zapytanie odbioru zadania (nie „creatives”).
let claimFn: ((...a: any[]) => any) | null = null, claimFnLook = 0;
function discordClaimFn() {
    if (claimFn || Date.now() - claimFnLook < 20_000) return claimFn;
    claimFnLook = Date.now();
    try {
        const test = (f: any) => {
            if (typeof f !== "function") return false;
            const src = Function.prototype.toString.call(f);
            return src.length < 6000 && /QUESTS?_CLAIM_REWARD\b/.test(src) && !/CREATIVES/.test(src);
        };
        for (const id in wreq.c) {
            const ex = (wreq.c as any)[id]?.exports;
            if (!ex || typeof ex !== "object" && typeof ex !== "function") continue;
            for (const k of Object.keys(ex)) {
                let v: any;
                try { v = ex[k]; } catch { continue; }
                if (test(v)) { claimFn = v; break; }
                if (v && typeof v === "object") for (const k2 of Object.keys(v)) { let w: any; try { w = v[k2]; } catch { continue; } if (test(w)) { claimFn = w; break; } }
                if (claimFn) break;
            }
            if (claimFn) break;
        }
        if (claimFn) claimDebug = { ...claimDebug, fn: { args: claimFn.length, src: Function.prototype.toString.call(claimFn).slice(0, 2500) } };
    } catch (e) { console.error(TAG, "claim fn", e); }
    return claimFn;
}

function unwatchClaims() { if (origPost) (RestAPI as any).post = origPost; origPost = null; }

async function claim(q: Quest, byHand: boolean): Promise<boolean> {
    if (claiming) return false;
    claiming = true;
    claimTried.set(q.id, Date.now());
    try {
        const fn = discordClaimFn();
        if (fn) { // tak jak przycisk w Discordzie: platforma 0 (wszędzie), miejsce 10 (strona zadań)
            let thrown: any = null, seen: any = null, okSeen = false;
            for (const loc of [11, 10]) { // 11: pasek zadania, 10: strona zadań – Discord przyjmował oba
                const mark = Date.now();
                thrown = null;
                try { await fn(q.id, 0, loc); } catch (e) { thrown = e; }
                await new Promise(r => setTimeout(r, 600));
                seen = claimDebug.seen;
                okSeen = !!(seen && seen.at >= mark - 50 && seen.ok);
                if (okSeen || questsStore()?.getQuest?.(q.id)?.userStatus?.claimedAt) break;
            }
            const st = questsStore()?.getQuest?.(q.id)?.userStatus;
            if (okSeen || st?.claimedAt) {
                claimDebug = { ...claimDebug, lastTry: { via: "discord", at: Date.now(), ok: true } };
                known.set(q.id, { ...q, claimedAt: Date.now() });
                say("claimed", q);
                lastLoad = 0;
                return true;
            }
            claimDebug = { ...claimDebug, lastTry: { via: "discord", at: Date.now(), ok: false, err: thrown ? errText(thrown) : seen?.err, status: thrown?.status ?? seen?.status, raw: thrown?.body } };
            throw thrown ?? { body: { message: seen?.err ?? "Discord did not claim" } };
        }
        if (claimTpl?.ok) { // Discord już odbierał, ale jego funkcji nie ma w załadowanym kodzie – trzeba otworzyć stronę zadań
            say("claimNoFn", q);
            if (byHand) openQuests();
            return false;
        }
        let url = `/quests/${q.id}/claim-reward`;
        let body: any = { platform: 0, location: 11, is_targeted: false, metadata_raw: null, metadata_sealed: null, traffic_metadata_raw: null, traffic_metadata_sealed: null };
        if (claimTpl?.ok) { // to, co Discord sam wysłał – z zadaniem podmienionym; zapieczętowane dane były dla tamtego zadania
            url = claimTpl.url.replace("{id}", q.id);
            body = JSON.parse(JSON.stringify(claimTpl.body ?? {}));
            for (const k of Object.keys(body ?? {})) if (/sealed/i.test(k)) body[k] = null;
        }
        ownPost = true;
        try { await RestAPI.post({ url, body }); } catch (e: any) {
            claimDebug = { ...claimDebug, lastTry: { url, body, at: Date.now(), status: e?.status, code: e?.body?.code, err: errText(e), raw: e?.body } };
            throw e;
        } finally { ownPost = false; }
        claimDebug = { ...claimDebug, lastTry: { url, at: Date.now(), ok: true } };
        known.set(q.id, { ...q, claimedAt: Date.now() });
        say("claimed", q);
        lastLoad = 0; // przeczytaj listę (i Orby) od nowa
        return true;
    } catch (e: any) {
        const b = e?.body ?? {};
        claimRefused.set(q.id, Date.now());
        const k = b.captcha_key || b.captcha_sitekey || b.captcha_service ? "claimCaptcha" : claimTpl?.ok ? "claimFail" : "claimLearn";
        lastClaimEv = { k, q, msg: errText(e) };
        if (testing && k !== "claimCaptcha") events.push({ k: "claimSkip", id: q.id, name: q.name, msg: errText(e) }); // w teście: tylko wpis do dziennika
        else say(k, q, errText(e));
        if (byHand) openQuests();
        return false;
    } finally {
        claiming = false;
    }
}

function openQuests() {
    try { NavigationRouter.transitionTo("/quest-home"); } catch (e) { console.error(TAG, e); }
    Native.focusDiscord();
}

// Nagrody do odebrania: najpierw te z komputera, na końcu zadania „na telefonie” (Discord może chcieć ich odbioru
// w aplikacji mobilnej); te, których Discord niedawno nie przyjął, czekają 6 godzin.
const claimRefused = new Map<string, number>();
function claimable(): Quest[] {
    const now = Date.now();
    return [...known.values()]
        .filter(q => q.completedAt && !q.claimedAt && !(q.expiresAt && q.expiresAt < now) && now - (claimRefused.get(q.id) ?? 0) > 6 * 3600_000)
        .sort((a, b) => Number(a.task.includes("MOBILE")) - Number(b.task.includes("MOBILE")));
}

async function autoClaimTick() {
    if (!autoClaim || claiming) return;
    for (const q of claimable()) {
        if (Date.now() - (claimTried.get(q.id) ?? 0) < 30 * 60_000) continue;
        await claim(q, false);
        return; // jedna na raz, kolejna przy następnym obrocie
    }
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
        case "claimTest": { // spróbuj odebrać teraz – do trzech różnych nagród, aż któraś przejdzie
            const list = claimable().slice(0, 3);
            if (!list.length) { say("claimNone"); break; }
            testing = true;
            let ok = false;
            for (const c of list) if ((ok = await claim(c, false))) break;
            testing = false;
            if (!ok) say(lastClaimEv.k, lastClaimEv.q, lastClaimEv.msg);
            break;
        }
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
            if (q && autoClaim && q.completedAt && !q.claimedAt) await claim(q, true);
            else openQuests();
            break;
    }
}

let timer: ReturnType<typeof setTimeout> | undefined;
let alive = false, lastOk = 0;

async function tick() {
    if (!alive) return;
    try {
        if (!lastLoad || Date.now() - lastLoad > refreshMs) await load();
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
            orbTile, orbGlyph, orbTiles,
            claimDebug: (() => { const j = JSON.stringify(claimDebug); if (j === claimDebugSent) return undefined; claimDebugSent = j; return claimDebug; })(),
            orbDebug: (() => { const j = JSON.stringify(orbDebug); if (j === orbDebugSent) return undefined; orbDebugSent = j; return orbDebug; })()
        });
        const raw = await Native.sync(body);
        if (raw) {
            lastOk = Date.now();
            const res = JSON.parse(raw);
            autoAllowed = !!res.auto;
            autoClaim = !!res.claim;
            if (Number(res.refresh) >= 1) refreshMs = Number(res.refresh) * 60_000;
            maxParallel = Math.max(0, Number(res.parallel ?? 1));
            if (!autoAllowed && (runs.size || queue.length)) stopAll();
            for (const c of res.cmds ?? []) await doCmd(c);
            autoClaimTick();
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
        watchClaims();
        timer = setTimeout(tick, 3000);
    },

    stop() {
        alive = false;
        unwatchClaims();
        if (timer) clearTimeout(timer);
        stopAll();
    }
});
