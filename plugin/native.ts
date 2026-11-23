/*
 * Orbek by Sówka – część natywna (proces główny Electrona).
 * Tylko stąd da się zajrzeć na dysk i rozmawiać z programem Orbek na tym komputerze (127.0.0.1).
 */

import { spawn } from "child_process";
import { BrowserWindow, IpcMainInvokeEvent } from "electron";
import { existsSync } from "fs";
import { readFile } from "fs/promises";
import { homedir } from "os";
import { join } from "path";

const dir = () => join(process.env.APPDATA ?? join(homedir(), "AppData", "Roaming"), "Orbek");

async function readJson(name: string): Promise<any> {
    try {
        return JSON.parse(await readFile(join(dir(), name), "utf8"));
    } catch {
        return null;
    }
}

/** Wysyła stan zadań do Orbka i zwraca jego odpowiedź (co kliknięto); null = Orbek nie działa. */
export async function sync(_: IpcMainInvokeEvent, body: string): Promise<string | null> {
    const b = await readJson("bridge.json");
    if (!b?.port || !b?.key) return null;
    try {
        const res = await fetch(`http://127.0.0.1:${b.port}/bridge/sync`, {
            method: "POST",
            headers: { "Content-Type": "application/json", "X-Orbek-Key": b.key },
            body,
            signal: AbortSignal.timeout(4000)
        });
        return res.ok ? await res.text() : null;
    } catch {
        return null;
    }
}

/** Uruchamia Orbka (zapamiętane miejsce z where.json). */
export async function openOrbek(_: IpcMainInvokeEvent): Promise<boolean> {
    const w = await readJson("where.json");
    if (!w?.exe || !existsSync(w.exe)) return false;
    spawn(w.exe, [], { detached: true, stdio: "ignore" }).unref();
    return true;
}

/** Pokazuje okno Discorda (np. żeby odebrać nagrodę). */
export function focusDiscord(_: IpcMainInvokeEvent): void {
    const win = BrowserWindow.getAllWindows().find(w => !w.isDestroyed());
    if (!win) return;
    if (win.isMinimized()) win.restore();
    win.show();
    win.focus();
}
