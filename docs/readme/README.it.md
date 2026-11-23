<p align="center"><img src="../../docs/img/icon.png" width="104" alt=""></p>
<h1 align="center">Orbek – Discord Quests &amp; Orbs</h1>
<p align="center">Orbek mostra tutte le missioni di Discord – per Orb, cosmetici di Discord e premi nei giochi – ordinate come vuoi, con l'avanzamento in diretta. Installa tutto ciò che gli serve in Discord con un clic.</p>
<p align="center"><a href="../../README.md" title="English"><img src="../../docs/flags/en.svg" width="24" alt="English"></a> <a href="../../docs/readme/README.pl.md" title="Polski"><img src="../../docs/flags/pl.svg" width="24" alt="Polski"></a> <a href="../../docs/readme/README.de.md" title="Deutsch"><img src="../../docs/flags/de.svg" width="24" alt="Deutsch"></a> <a href="../../docs/readme/README.fr.md" title="Français"><img src="../../docs/flags/fr.svg" width="24" alt="Français"></a> <a href="../../docs/readme/README.es.md" title="Español"><img src="../../docs/flags/es.svg" width="24" alt="Español"></a> <a href="../../docs/readme/README.it.md" title="Italiano"><img src="../../docs/flags/it.svg" width="24" alt="Italiano"></a> <a href="../../docs/readme/README.pt.md" title="Português (BR)"><img src="../../docs/flags/pt.svg" width="24" alt="Português (BR)"></a> <a href="../../docs/readme/README.ru.md" title="Русский"><img src="../../docs/flags/ru.svg" width="24" alt="Русский"></a> <a href="../../docs/readme/README.uk.md" title="Українська"><img src="../../docs/flags/uk.svg" width="24" alt="Українська"></a> <a href="../../docs/readme/README.tr.md" title="Türkçe"><img src="../../docs/flags/tr.svg" width="24" alt="Türkçe"></a> <a href="../../docs/readme/README.ja.md" title="日本語"><img src="../../docs/flags/ja.svg" width="24" alt="日本語"></a> <a href="../../docs/readme/README.zh.md" title="简体中文"><img src="../../docs/flags/zh.svg" width="24" alt="简体中文"></a></p>
<p align="center">
  <a href="https://github.com/Yasaiek/Orbek-Discord-Quests-Orbs/releases/latest/download/Orbek.exe"><b>⬇ Scarica per Windows</b></a> &nbsp;·&nbsp;
  <a href="https://yasaiek.github.io/Orbek-Discord-Quests-Orbs/"><b>🌐 Sito</b></a> &nbsp;·&nbsp;
  <a href="../../CHANGELOG.md">Novità</a> &nbsp;·&nbsp;
  <a href="https://github.com/Yasaiek/Orbek-Discord-Quests-Orbs/issues/new">Segnala un bug</a>
</p>
<p align="center"><img src="../../docs/img/panel.png" alt="Orbek" width="880"></p>

## Cosa fa

- **Tutte le missioni in un posto** – Giochi, video e attività in schede, con il tempo rimasto e anelli di avanzamento in diretta.
- **Premi a colpo d'occhio** – Orb, decorazioni avatar, effetti profilo, oggetti di gioco – vedi cosa dà una missione prima di iniziarla.
- **Ordinamento e filtri** – Scadono prima, più Orb, più brevi o più nuove; solo cosmetici di Discord; vista mini e finestra piccola.
- **Installazione con un clic** – Orbek scarica Vencord, aggiunge il suo plugin, lo compila e lo collega a Discord. Niente da installare a mano.
- **Automazione a scelta** – Orbek può fare le missioni da solo, più insieme. Disattivata di default – Discord lo vieta, quindi a tuo rischio.
- **Si aggiorna da solo** – Le nuove versioni arrivano dalle release di GitHub, verificate con SHA-256. 12 lingue.

## Guida

1. **Scarica e avvia Orbek** – Scarica Orbek.exe dalla pagina del progetto, mettilo in una cartella a parte e avvialo. L'icona compare vicino all'orologio.
2. **Clicca «Installa tutto»** – Nella scheda Installazione Orbek scarica Node.js e Vencord, inserisce il plugin, lo compila e lo collega a Discord. Ci vogliono pochi minuti.
3. **Discord si riavvia da solo** – Dopo l'installazione Discord si riavvia con il plugin Orbek. In basso nella finestra «Discord collegato» diventa verde.
4. **Sfoglia le missioni** – Schede Giochi, Video e Per Orb, ordinamento e filtro dei cosmetici. Ogni missione mostra premi, avanzamento e tempo rimasto.
5. **Accetta e riscatta** – «Accetta missione» la prende. Quando è completata, clicca «Riscatta in Discord» – riscatti il premio nella pagina delle missioni (c'è un captcha).
6. **Automazione – a tuo rischio** – Nelle Impostazioni puoi lasciare che Orbek faccia le missioni da solo, più insieme. Discord lo vieta e può togliere i premi o bannare l'account.
7. **Vista mini** – Passa alla vista «Mini» o apri «Finestra piccola» per vedere solo i premi e gli anelli di avanzamento.
8. **Aggiornamenti** – Orbek si aggiorna da solo da GitHub. Quando esce un plugin più nuovo, compare la barra «Aggiorna plugin».

## Domande

**È permesso?**  
Sfogliare e accettare missioni è esattamente ciò che fa Discord. Farle automaticamente viola le sue regole – può togliere i premi o bannare l'account.

**Orbek vede la mia password o il mio token di Discord?**  
No. Tutto passa da un plugin nella tua app Discord. Orbek riceve solo la lista delle missioni; il token non esce mai da Discord.

**Cos'è Vencord?**  
Una mod popolare per l'app Discord su computer. Il plugin di Orbek gira lì dentro; l'installazione la mette per te.

**Windows avvisa di un editore sconosciuto.**  
Il file non è firmato. Il codice è pubblico e ogni release viene compilata da GitHub Actions.

## Compilare dai sorgenti

Go 1.24 o più recente:

```
go build -trimpath -ldflags "-H=windowsgui -s -w" -o Orbek.exe .
```

## Avvertenza

Fare le missioni di Discord automaticamente viola le regole di Discord. Orbek non è affiliato a Discord. Lo usi a tuo rischio.
