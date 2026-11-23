# Changelog

## 0.3.9
- Claim test tries up to three rewards (phone-video quests last) and both places Discord claims from; one quest Discord refuses no longer stops the others – it waits 6 hours and goes to the journal.

## 0.3.8
- Claiming goes through Discord's own claim function (the one behind its "Claim reward" button), which adds the sealed data Discord wants for each quest. If Discord asks for a captcha, it shows it itself. Needs Discord's quests page opened once.

## 0.3.7
- Anonymous statistics for the website (on by default, Settings → "Send anonymous statistics"): every 15 minutes a random installation number, the version and the average CPU and RAM use. The site shows how many copies run now and their average use.

## 0.3.6
- The Orbs icon is animated; Settings → "Animated Orbs icon" switches it to a still picture.

## 0.3.5
- Discord's Orbs picture everywhere in the app: counters, rewards, tabs, labels.
- Claiming learns from Discord: claim one reward by hand in Discord and the plugin copies exactly how Discord asks for it. When the default way is refused, the panel says so. Details go to data/claim-debug.json.
- The plugin remembers which Orbs tile Discord shows for which amount.

## 0.3.4
- CPU and RAM have their own tiles: now, lowest and highest since start, with a range bar.
- The claim test no longer hangs: with an old plugin it says to update the plugin, and without an answer it stops after 45 s.

## 0.3.3
- After an update the open panel reloads itself, so the new version shows right away.
- The tray menu is dark (with the colourful icons), also when Windows uses the light theme.

## 0.3.2
- Coloured buttons no longer show a thin line at their left edge.

## 0.3.1
- Discord activity is on: Orbek's own Discord application is built in (Settings → Discord activity).

## 0.3.0
- Quests page: counters for your Orbs, rewards waiting to be claimed and rewards in quests not done yet (Orbs plus cosmetics, in-game items, codes), and PC usage now, lowest and highest.
- Settings redone: tick-box chips for rewards, tasks and games ("All games" or only the ticked ones), "how many at once" as buttons.
- Claiming (experimental) has its own card: when switched on it is tested first; if Discord wants a captcha or refuses, it switches itself off and says why. "Test now" button.
- How often to look for new quests: minutes, hours or days.
- Discord activity (Rich Presence): shown only while Orbek does a quest – quest, progress, Orbs, a "Get Orbek" button; each part can be switched off. Turns on once Orbek's Discord application is set.
- The plugin looks harder for Discord's own Orbs picture and notes what it saw (data/orb-debug.json).
- Colourful tray menu icons; nicer buttons; guide steps with icon tiles (also on the website).

## 0.2.3
- Games in the automat's filter show the game's tile; reward kinds have coloured icons (Orbs with Discord's own picture).

## 0.2.2
- Settings → Automat: choose which rewards (Orbs, Discord cosmetics, in-game items, codes, other), which tasks (videos, playing, activities, streaming) and which games (all, only chosen, all but chosen – from the games Orbek has seen) the automat takes.
- Experimental: claim finished quests by themselves, with Discord's own request. If Discord wants a captcha, it shows it in Discord or Orbek opens the quests page.
- Right-click menu in the tray: state, PC usage, open, small window, stop everything, updates, website, folder, quit – with icons.
- PC usage (CPU and RAM of Orbek with its windows) on the quests page and in the tray.
- Discord's own Orbs pictures: the colourful reward tile, or Discord's Orbs sign until the tile is seen.
- Progress on one line: time, bar and the ring centred together.
- Language picker with flags.

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
