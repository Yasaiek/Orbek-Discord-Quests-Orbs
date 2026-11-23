<p align="center"><img src="../../docs/img/icon.png" width="104" alt=""></p>
<h1 align="center">Orbek – Discord Quests &amp; Orbs</h1>
<p align="center">O Orbek mostra todas as missões do Discord – por Orbs, cosméticos do Discord e recompensas em jogos – ordenadas do seu jeito, com o progresso ao vivo. Ele instala tudo o que precisa no Discord com um clique.</p>
<p align="center"><a href="../../README.md" title="English"><img src="../../docs/flags/en.svg" width="24" alt="English"></a> <a href="../../docs/readme/README.pl.md" title="Polski"><img src="../../docs/flags/pl.svg" width="24" alt="Polski"></a> <a href="../../docs/readme/README.de.md" title="Deutsch"><img src="../../docs/flags/de.svg" width="24" alt="Deutsch"></a> <a href="../../docs/readme/README.fr.md" title="Français"><img src="../../docs/flags/fr.svg" width="24" alt="Français"></a> <a href="../../docs/readme/README.es.md" title="Español"><img src="../../docs/flags/es.svg" width="24" alt="Español"></a> <a href="../../docs/readme/README.it.md" title="Italiano"><img src="../../docs/flags/it.svg" width="24" alt="Italiano"></a> <a href="../../docs/readme/README.pt.md" title="Português (BR)"><img src="../../docs/flags/pt.svg" width="24" alt="Português (BR)"></a> <a href="../../docs/readme/README.ru.md" title="Русский"><img src="../../docs/flags/ru.svg" width="24" alt="Русский"></a> <a href="../../docs/readme/README.uk.md" title="Українська"><img src="../../docs/flags/uk.svg" width="24" alt="Українська"></a> <a href="../../docs/readme/README.tr.md" title="Türkçe"><img src="../../docs/flags/tr.svg" width="24" alt="Türkçe"></a> <a href="../../docs/readme/README.ja.md" title="日本語"><img src="../../docs/flags/ja.svg" width="24" alt="日本語"></a> <a href="../../docs/readme/README.zh.md" title="简体中文"><img src="../../docs/flags/zh.svg" width="24" alt="简体中文"></a></p>
<p align="center">
  <a href="https://github.com/Yasaiek/Orbek-Discord-Quests-Orbs/releases/latest/download/Orbek.exe"><b>⬇ Baixar para Windows</b></a> &nbsp;·&nbsp;
  <a href="https://yasaiek.github.io/Orbek-Discord-Quests-Orbs/"><b>🌐 Site</b></a> &nbsp;·&nbsp;
  <a href="../../CHANGELOG.md">Mudanças</a> &nbsp;·&nbsp;
  <a href="https://github.com/Yasaiek/Orbek-Discord-Quests-Orbs/issues/new">Relatar um bug</a>
</p>
<p align="center"><img src="../../docs/img/panel.png" alt="Orbek" width="880"></p>

## O que faz

- **Todas as missões em um lugar** – Jogos, vídeos e atividades em abas, com o tempo restante e anéis de progresso ao vivo.
- **Recompensas num relance** – Orbs, decorações de avatar, efeitos de perfil, itens de jogo – veja o que uma missão dá antes de começar.
- **Ordenação e filtros** – Acabam antes, mais Orbs, mais curtas ou mais novas; só cosméticos do Discord; visão mini e janela pequena.
- **Instalação com um clique** – O Orbek baixa o Vencord, adiciona o plugin, compila e liga ao Discord. Nada para instalar à mão.
- **Automação opcional** – O Orbek pode fazer as missões sozinho, várias ao mesmo tempo. Desligada por padrão – o Discord proíbe, então por sua conta e risco.
- **Atualiza sozinho** – Novas versões vêm dos releases do GitHub, verificadas por SHA-256. 12 idiomas.

## Tutorial

1. **Baixe e abra o Orbek** – Baixe o Orbek.exe na página do projeto, coloque numa pasta própria e abra. O ícone aparece perto do relógio.
2. **Clique em “Instalar tudo”** – Na aba Instalação, o Orbek baixa o Node.js e o Vencord, coloca o plugin, compila e liga ao Discord. Leva alguns minutos.
3. **O Discord reinicia sozinho** – Depois da instalação, o Discord abre de novo com o plugin Orbek. “Discord conectado” fica verde embaixo na janela.
4. **Veja as missões** – Abas Jogos, Vídeos e Por Orbs, ordenação e filtro de cosméticos. Cada missão mostra recompensas, progresso e tempo restante.
5. **Aceite e resgate** – “Aceitar missão” pega a missão. Quando estiver concluída, clique em “Resgatar no Discord” – você resgata a recompensa na página de missões (tem captcha).
6. **Automação – por sua conta e risco** – Nas Configurações você pode deixar o Orbek fazer as missões sozinho, várias ao mesmo tempo. O Discord proíbe isso e pode tirar as recompensas ou banir a conta.
7. **Visão mini** – Mude a visão para “Mini” ou abra “Janela pequena” para ver só as recompensas e os anéis de progresso.
8. **Atualizações** – O Orbek se atualiza sozinho pelo GitHub. Quando sair um plugin mais novo, aparece a barra “Atualizar plugin”.

## Perguntas

**Isso é permitido?**  
Ver e aceitar missões é exatamente o que o próprio Discord faz. Fazê-las automaticamente quebra as regras – ele pode tirar as recompensas ou banir a conta.

**O Orbek vê minha senha ou token do Discord?**  
Não. Tudo passa por um plugin dentro do seu próprio app do Discord. O Orbek só recebe a lista de missões; seu token nunca sai do Discord.

**O que é o Vencord?**  
Um mod popular para o app do Discord no computador. O plugin do Orbek roda nele; a instalação coloca tudo para você.

**O Windows avisa sobre um editor desconhecido.**  
O arquivo não é assinado. O código é público e cada release é compilado dele pelo GitHub Actions.

## Compilar do código

Go 1.24 ou mais novo:

```
go build -trimpath -ldflags "-H=windowsgui -s -w" -o Orbek.exe .
```

## Aviso

Fazer missões do Discord automaticamente vai contra as regras do Discord. O Orbek não tem ligação com o Discord. Use por sua conta e risco.
