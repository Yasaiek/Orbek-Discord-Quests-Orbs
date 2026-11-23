<p align="center"><img src="../../docs/img/icon.png" width="104" alt=""></p>
<h1 align="center">Orbek – Discord Quests &amp; Orbs</h1>
<p align="center">Orbek muestra todas las misiones de Discord – por Orbs, cosméticos de Discord y recompensas en juegos – ordenadas como quieras, con el progreso en directo. Instala todo lo que necesita en Discord con un clic.</p>
<p align="center"><a href="../../README.md" title="English"><img src="../../docs/flags/en.svg" width="24" alt="English"></a> <a href="../../docs/readme/README.pl.md" title="Polski"><img src="../../docs/flags/pl.svg" width="24" alt="Polski"></a> <a href="../../docs/readme/README.de.md" title="Deutsch"><img src="../../docs/flags/de.svg" width="24" alt="Deutsch"></a> <a href="../../docs/readme/README.fr.md" title="Français"><img src="../../docs/flags/fr.svg" width="24" alt="Français"></a> <a href="../../docs/readme/README.es.md" title="Español"><img src="../../docs/flags/es.svg" width="24" alt="Español"></a> <a href="../../docs/readme/README.it.md" title="Italiano"><img src="../../docs/flags/it.svg" width="24" alt="Italiano"></a> <a href="../../docs/readme/README.pt.md" title="Português (BR)"><img src="../../docs/flags/pt.svg" width="24" alt="Português (BR)"></a> <a href="../../docs/readme/README.ru.md" title="Русский"><img src="../../docs/flags/ru.svg" width="24" alt="Русский"></a> <a href="../../docs/readme/README.uk.md" title="Українська"><img src="../../docs/flags/uk.svg" width="24" alt="Українська"></a> <a href="../../docs/readme/README.tr.md" title="Türkçe"><img src="../../docs/flags/tr.svg" width="24" alt="Türkçe"></a> <a href="../../docs/readme/README.ja.md" title="日本語"><img src="../../docs/flags/ja.svg" width="24" alt="日本語"></a> <a href="../../docs/readme/README.zh.md" title="简体中文"><img src="../../docs/flags/zh.svg" width="24" alt="简体中文"></a></p>
<p align="center">
  <a href="https://github.com/Yasaiek/Orbek-Discord-Quests-Orbs/releases/latest/download/Orbek.exe"><b>⬇ Descargar para Windows</b></a> &nbsp;·&nbsp;
  <a href="https://yasaiek.github.io/Orbek-Discord-Quests-Orbs/"><b>🌐 Web</b></a> &nbsp;·&nbsp;
  <a href="../../CHANGELOG.md">Cambios</a> &nbsp;·&nbsp;
  <a href="https://github.com/Yasaiek/Orbek-Discord-Quests-Orbs/issues/new">Informar de un error</a>
</p>
<p align="center"><img src="../../docs/img/panel.png" alt="Orbek" width="880"></p>

## Qué hace

- **Todas las misiones en un sitio** – Juegos, vídeos y actividades en pestañas, con el tiempo restante y anillos de progreso en directo.
- **Recompensas de un vistazo** – Orbs, decoraciones de avatar, efectos de perfil, objetos de juego – ve qué da una misión antes de empezarla.
- **Orden y filtros** – Terminan antes, más Orbs, más cortas o más nuevas; solo cosméticos de Discord; vista mini y ventana pequeña.
- **Instalación con un clic** – Orbek descarga Vencord, añade su plugin, lo compila y lo conecta a Discord. Nada que instalar a mano.
- **Automatización opcional** – Orbek puede hacer las misiones solo, varias a la vez. Desactivada por defecto – Discord lo prohíbe, así que bajo tu responsabilidad.
- **Se actualiza solo** – Las nuevas versiones vienen de los releases de GitHub, comprobadas con SHA-256. 12 idiomas.

## Tutorial

1. **Descarga y abre Orbek** – Descarga Orbek.exe de la página del proyecto, ponlo en una carpeta propia y ábrelo. Su icono aparece junto al reloj.
2. **Pulsa «Instalar todo»** – En la pestaña Instalación, Orbek descarga Node.js y Vencord, pone el plugin, lo compila y lo conecta a Discord. Tarda unos minutos.
3. **Discord se reinicia solo** – Tras la instalación, Discord se abre de nuevo con el plugin Orbek. «Discord conectado» se pone en verde abajo en la ventana.
4. **Explora las misiones** – Pestañas Juegos, Vídeos y Por Orbs, ordenación y filtro de cosméticos. Cada misión muestra recompensas, progreso y tiempo restante.
5. **Acepta y reclama** – «Aceptar misión» la toma. Cuando esté completada, pulsa «Reclamar en Discord» – reclamas la recompensa en la página de misiones (hay un captcha).
6. **Automatización – bajo tu responsabilidad** – En Ajustes puedes dejar que Orbek haga las misiones solo, varias a la vez. Discord lo prohíbe y puede quitar las recompensas o banear la cuenta.
7. **Vista mini** – Cambia la vista a «Mini» o abre «Ventana pequeña» para ver solo las recompensas y los anillos de progreso.
8. **Actualizaciones** – Orbek se actualiza solo desde GitHub. Cuando salga un plugin más nuevo verás la barra «Actualizar plugin».

## Preguntas

**¿Está permitido?**  
Ver y aceptar misiones es justo lo que hace Discord. Hacerlas automáticamente rompe sus normas – puede quitar las recompensas o banear la cuenta.

**¿Ve Orbek mi contraseña o mi token de Discord?**  
No. Todo pasa por un plugin dentro de tu propia app de Discord. Orbek solo recibe la lista de misiones; tu token nunca sale de Discord.

**¿Qué es Vencord?**  
Un mod popular para la app de Discord de escritorio. El plugin de Orbek funciona dentro; la instalación lo pone por ti.

**Windows avisa de un editor desconocido.**  
El archivo no está firmado. El código es público y cada release lo compila GitHub Actions.

## Compilar desde el código

Go 1.24 o más nuevo:

```
go build -trimpath -ldflags "-H=windowsgui -s -w" -o Orbek.exe .
```

## Aviso

Hacer misiones de Discord automáticamente va contra sus normas. Orbek no está afiliado a Discord. Lo usas bajo tu responsabilidad.
