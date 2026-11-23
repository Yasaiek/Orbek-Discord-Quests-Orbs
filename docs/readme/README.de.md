<p align="center"><img src="../../docs/img/icon.png" width="104" alt=""></p>
<h1 align="center">Orbek – Discord Quests &amp; Orbs</h1>
<p align="center">Orbek zeigt alle Discord-Quests – für Orbs, Discord-Kosmetik und Spielbelohnungen – sortiert, wie du willst, mit Fortschritt live. Alles, was es in Discord braucht, installiert es mit einem Klick.</p>
<p align="center"><a href="../../README.md" title="English"><img src="../../docs/flags/en.svg" width="24" alt="English"></a> <a href="../../docs/readme/README.pl.md" title="Polski"><img src="../../docs/flags/pl.svg" width="24" alt="Polski"></a> <a href="../../docs/readme/README.de.md" title="Deutsch"><img src="../../docs/flags/de.svg" width="24" alt="Deutsch"></a> <a href="../../docs/readme/README.fr.md" title="Français"><img src="../../docs/flags/fr.svg" width="24" alt="Français"></a> <a href="../../docs/readme/README.es.md" title="Español"><img src="../../docs/flags/es.svg" width="24" alt="Español"></a> <a href="../../docs/readme/README.it.md" title="Italiano"><img src="../../docs/flags/it.svg" width="24" alt="Italiano"></a> <a href="../../docs/readme/README.pt.md" title="Português (BR)"><img src="../../docs/flags/pt.svg" width="24" alt="Português (BR)"></a> <a href="../../docs/readme/README.ru.md" title="Русский"><img src="../../docs/flags/ru.svg" width="24" alt="Русский"></a> <a href="../../docs/readme/README.uk.md" title="Українська"><img src="../../docs/flags/uk.svg" width="24" alt="Українська"></a> <a href="../../docs/readme/README.tr.md" title="Türkçe"><img src="../../docs/flags/tr.svg" width="24" alt="Türkçe"></a> <a href="../../docs/readme/README.ja.md" title="日本語"><img src="../../docs/flags/ja.svg" width="24" alt="日本語"></a> <a href="../../docs/readme/README.zh.md" title="简体中文"><img src="../../docs/flags/zh.svg" width="24" alt="简体中文"></a></p>
<p align="center">
  <a href="https://github.com/Yasaiek/Orbek-Discord-Quests-Orbs/releases/latest/download/Orbek.exe"><b>⬇ Für Windows laden</b></a> &nbsp;·&nbsp;
  <a href="https://yasaiek.github.io/Orbek-Discord-Quests-Orbs/"><b>🌐 Website</b></a> &nbsp;·&nbsp;
  <a href="../../CHANGELOG.md">Änderungen</a> &nbsp;·&nbsp;
  <a href="https://github.com/Yasaiek/Orbek-Discord-Quests-Orbs/issues/new">Fehler melden</a>
</p>
<p align="center"><img src="../../docs/img/panel.png" alt="Orbek" width="880"></p>
<table align="center"><tr>
<td align="center" valign="top"><img src="../../docs/img/settings.png" alt="" width="420"></td>
<td align="center" valign="top"><img src="../../docs/img/guide.png" alt="" width="420"></td>
<td align="center" valign="top"><img src="../../docs/img/mini.png" alt="" width="150"></td>
</tr></table>

## Was es kann

- **Alle Quests an einem Ort** – Spiele, Videos und Aktivitäten in Tabs, mit Restzeit und Live-Fortschrittsringen.
- **Belohnungen auf einen Blick** – Orbs, Avatar-Dekorationen, Profileffekte, Spielgegenstände – du siehst, was eine Quest bringt, bevor du anfängst.
- **Sortierung und Filter** – Enden am frühesten, meiste Orbs, kürzeste oder neueste; nur Discord-Kosmetik; Mini-Ansicht und kleines Fenster.
- **Einrichtung mit einem Klick** – Orbek lädt Vencord, fügt sein Plugin hinzu, baut es und hängt es in Discord ein. Nichts von Hand installieren.
- **Automatik nach Wahl** – Orbek kann Quests selbst erledigen, mehrere gleichzeitig. Standardmäßig aus – Discord verbietet es, also auf eigenes Risiko.
- **Aktualisiert sich selbst** – Neue Versionen kommen aus den GitHub-Releases, geprüft per SHA-256. 12 Sprachen.

## Anleitung

1. **Orbek laden und starten** – Lade Orbek.exe von der Projektseite, lege es in einen eigenen Ordner und starte es. Das Symbol erscheint neben der Uhr.
2. **Auf „Alles installieren“ klicken** – Im Tab Einrichtung lädt Orbek Node.js und Vencord, legt das Plugin hinein, baut es und hängt es in Discord ein. Das dauert ein paar Minuten.
3. **Discord startet von selbst neu** – Nach der Einrichtung startet Discord mit dem Orbek-Plugin neu. Unten im Fenster wird „Discord verbunden“ grün.
4. **Quests durchsehen** – Tabs Spiele, Videos und Für Orbs, Sortierung und ein Kosmetik-Filter. Jede Quest zeigt Belohnungen, Fortschritt und Restzeit.
5. **Annehmen und abholen** – „Quest annehmen“ nimmt die Quest an. Ist sie erledigt, klicke „Nacheinander in Discord abholen“ – Discord zeigt die Belohnungen nacheinander, du löst nur ein Captcha, wenn eins kommt.
6. **Automatik – auf eigenes Risiko** – In den Einstellungen kannst du Orbek Quests selbst erledigen lassen, mehrere gleichzeitig. Discord verbietet das und kann Belohnungen zurücknehmen oder das Konto sperren.
7. **Mini-Ansicht** – Schalte die Ansicht auf „Mini“ oder öffne „Kleines Fenster“, um nur Belohnungen und Fortschrittsringe zu sehen.
8. **Updates** – Orbek aktualisiert sich selbst von GitHub. Gibt es ein neueres Plugin, erscheint die Leiste „Plugin aktualisieren“.

## Fragen

**Ist das erlaubt?**  
Quests ansehen und annehmen ist genau das, was Discord selbst tut. Sie automatisch zu erledigen verstößt gegen Discords Regeln – Belohnungen können zurückgenommen oder das Konto gesperrt werden.

**Sieht Orbek mein Discord-Passwort oder Token?**  
Nein. Alles läuft über ein Plugin in deiner eigenen Discord-App. Orbek bekommt von ihm nur die Quest-Liste; dein Token verlässt Discord nie.

**Was ist Vencord?**  
Ein beliebter Mod für die Discord-Desktop-App. Orbeks Plugin läuft darin; die Einrichtung installiert ihn für dich.

**Windows warnt vor einem unbekannten Herausgeber.**  
Die Datei ist nicht signiert. Der Quellcode ist öffentlich und jedes Release wird von GitHub Actions daraus gebaut.

## Aus dem Quellcode bauen

Go 1.24 oder neuer:

```
go build -trimpath -ldflags "-H=windowsgui -s -w" -o Orbek.exe .
```

## Haftungsausschluss

Discord-Quests automatisch zu erledigen verstößt gegen Discords Regeln. Orbek ist nicht mit Discord verbunden. Nutzung auf eigenes Risiko.
