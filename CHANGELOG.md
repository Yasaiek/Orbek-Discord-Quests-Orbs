# Changelog

## 0.4.21
- Plugin 0.4.5: "play a game" quests no longer stall until you open Discord's Quests page. When Orbek accepted a quest, Discord's own quest list didn't know it yet, and Discord only counts game time for quests it knows as accepted; now Orbek tells it right away and reloads that list (what opening the Quests page does).
- If Discord still sends no game-time heartbeat for 75 seconds, Orbek reloads its quest list and shows it the game again.

## 0.4.20
- The Discord status test button is gone again (the status works); the status line still says what went wrong, if anything.

## 0.4.19
- Settings → Activity on Discord: a "Test" button shows a test status in Discord for 30 seconds, whatever Orbek is doing, and says what came out (works / Discord not running / Discord refused it, with Discord's reason).
- The status line also tells what went wrong last time.

## 0.4.18
- Hand-only quests: no "Enroll" button any more (you accept and play them in Discord anyway) – "Show in Discord" is the button now.
- The activity's name is shown readably: the build tag of its developer app is dropped ("[Prod] Full-moon" → "Full Moon (Werwulf)").
- The "Do the goal in …" line and the repeated "Needed" count are gone from those cards; the note explains it all.

## 0.4.17
- When everything is installed and Discord runs the newest plugin, the Installation page leaves the menu, and if opened it just says all is set (no big install button, no steps or log).
- It comes back by itself when the plugin is out of date or Discord hasn't connected for a minute (e.g. a Discord update removed Vencord).

## 0.4.16
- Smoother CPU and RAM charts: values are averaged and softened, drawn as a rounded curve, and the curve glides to each new shape instead of jumping.
- "Check now" for updates shows that it is checking and what came out ("✓ Up to date"), and the update line says when it last checked.

## 0.4.15
- Hand-only quests ("Achievement in an activity") now say exactly why: the achievement has to be earned in the activity (a mini-game run inside Discord), the activity itself records the progress on its servers, and Discord accepts it only from there – not from Orbek or the plugin.
- The card shows the task's goal in Discord's own words (when Discord gives it) and has a "Show in Discord" button that opens Discord's quest page on that quest.
- Plugin 0.4.4: reports what each task asks and can open a quest in Discord.

## 0.4.14
- "What's new" shows file names cleanly (no code marks or doubled backslashes).

## 0.4.13
- The plugin install button is off (shows "Up to date") while Discord already runs the newest plugin; Orbek also refuses a reinstall then.
- Quests Orbek can't do by itself (e.g. "Achievement in an activity", where you have to earn the achievement yourself) get a clear note on the card, a proper name instead of the raw type, and a plain count (0 / 1) instead of minutes. The automat never picks them.

## 0.4.12
- After a plugin update Orbek no longer offers the same update again while Discord restarts: it shows that it waits for Discord to come back with the new plugin, and the install buttons stay off meanwhile.
- If Discord still runs the old plugin a few minutes after the update, Orbek says so (restart Discord or install again).

## 0.4.11
- Data now lives in `%LOCALAPPDATA%\Yasaiek\Orbek` (shared Yasaiek folder, each program in its own subfolder). Files from the old places are moved over on the first start.
- Settings → Data folder: open it, or pick another place; the data is moved and Orbek restarts.
- Quitting from the tray (or the panel) now closes Orbek's windows too.
- Orbek only ever looks at its own processes (checked by program name) and tags only its own windows: no other program, game included, is touched.
- Plugin 0.4.3: while a quest pretends to play a game, the games you really play stay visible to Discord (its overlay no longer loses them), and the pretend game gets an id no real program can have.

## 0.4.10
- Orbek keeps its files in its own Orbek-data folder. With Loorek in the same folder both programs shared `data\settings.json` and overwrote each other's settings; Orbek's files are moved over on the first start.

## 0.4.9
- Smoother CPU and RAM charts: averaged values drawn as a soft curve with a round marker, never cut off at the edges.
- In a narrower window the three Orb tiles stay in one row and CPU and RAM sit side by side below them.
- Dialog buttons no longer show a white frame when the window opens (Enter still confirms).

## 0.4.8
- Five counter tiles fit in one row again; the mini window has no claim bar.
- Guide: the claim step describes claiming one by one.

## 0.4.7
- CPU and RAM tiles show a small chart of the last minutes, with min and max chips under it.
- The counter tiles wrap onto the next row in a narrow window instead of being cut off.

## 0.4.6
- When a newer plugin comes with Orbek, the panel asks once whether to update it now (it says Discord will restart). Yes shows the update in a window: the steps and the live log, like the setup; then "Done".
- "Update plugin" buttons use the same window.

## 0.4.5
- Update screen: while Orbek downloads a new version the panel shows a progress window (downloading with MB, then installing and restarting).
- After the update a "What's new" window lists the changes since your previous version.

## 0.4.4
- An empty list says what's going on: "All done!" when every quest is finished, "No quests right now" when Discord has none (both say when Orbek checks again), and only when the tab, search or a filter hides the rest – that, with a "Show all" button.

## 0.4.3
- Claiming one by one: the plugin presses "Claim reward" itself (Settings → Claiming rewards); a captcha, when Discord shows one, is yours to solve. If Discord doesn't give the reward, the bar says so and waits for your click.

## 0.4.2
- Reminders: when a quest is done but its reward isn't claimed and the quest ends within 24 h (or 6, 12, 48 h – Settings → Reminders), Orbek shows a Windows notification with a sound, plays a sound in the open panel and offers to claim one by one. Once per quest.

## 0.4.1
- Claiming one by one: the "Congratulations" window closes by itself and the next reward's claim button is selected – press Enter or click to claim it.

## 0.4.0
- Automatic claiming removed: Discord refuses claims that a person didn't click.
- Instead: "Claim one by one in Discord". Orbek opens Discord's quests page, scrolls to the first reward waiting, highlights it and shows a bar "Claim … 1/5" (Skip, End). You click "Claim reward" in Discord; once Discord gives it, the next one comes up.

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
