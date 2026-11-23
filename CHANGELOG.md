# Changelog

## 0.2.1
- Fixed: the setup stopped at "pnpm install" (Cannot find module …pnpm.cjs). New npm versions block pnpm 12's install script; Orbek now runs pnpm from its own Node entry, which needs no script.
- Orbek has its own taskbar button instead of landing under another browser-app window (e.g. Loorek).
- Every quest says what to do and what you get, like Discord does ("Play X for 15 min to win: 700 Orbs").
- Orbs show Discord's own Orbs picture: the plugin picks it up from Discord's quests page once you have opened it.
- The live preview on the site has pictures on every quest.

## 0.2.0
- Progress is live: the plugin reports every second and the rings and bars move smoothly between reports.
- Settings: how many quests Orbek may do at once (1–50 or all).
- A progress ring next to the minutes and the bar on every quest.
- Mini view (just rewards and rings, side by side) and a small always-at-hand window.
- Sorting menu (ending soonest, most Orbs, shortest, newest) with icons; Discord cosmetics filter; hide claimed and ended.
- Setup tab: one click downloads Node.js and Vencord, adds the Orbek plugin, builds it, hooks it into Discord and turns it on. "Update plugin" when a newer plugin is out.
- Guide tab with the whole tutorial.
- Orbek updates itself from GitHub releases (checked with SHA-256).
- 12 languages.
- New look, Discord-like Orbs icon, bigger planet icon, own dialog windows instead of the browser's.

## 0.1.0
- First version: Discord quests in a separate window, with a Vencord plugin as the bridge.
