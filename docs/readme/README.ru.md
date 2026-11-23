<p align="center"><img src="../../docs/img/icon.png" width="104" alt=""></p>
<h1 align="center">Orbek – Discord Quests &amp; Orbs</h1>
<p align="center">Orbek показывает все задания Discord – за Orbs, косметику Discord и награды в играх – отсортированные как вам удобно, с прогрессом в реальном времени. Всё нужное в Discord он ставит одним нажатием.</p>
<p align="center"><a href="../../README.md" title="English"><img src="../../docs/flags/en.svg" width="24" alt="English"></a> <a href="../../docs/readme/README.pl.md" title="Polski"><img src="../../docs/flags/pl.svg" width="24" alt="Polski"></a> <a href="../../docs/readme/README.de.md" title="Deutsch"><img src="../../docs/flags/de.svg" width="24" alt="Deutsch"></a> <a href="../../docs/readme/README.fr.md" title="Français"><img src="../../docs/flags/fr.svg" width="24" alt="Français"></a> <a href="../../docs/readme/README.es.md" title="Español"><img src="../../docs/flags/es.svg" width="24" alt="Español"></a> <a href="../../docs/readme/README.it.md" title="Italiano"><img src="../../docs/flags/it.svg" width="24" alt="Italiano"></a> <a href="../../docs/readme/README.pt.md" title="Português (BR)"><img src="../../docs/flags/pt.svg" width="24" alt="Português (BR)"></a> <a href="../../docs/readme/README.ru.md" title="Русский"><img src="../../docs/flags/ru.svg" width="24" alt="Русский"></a> <a href="../../docs/readme/README.uk.md" title="Українська"><img src="../../docs/flags/uk.svg" width="24" alt="Українська"></a> <a href="../../docs/readme/README.tr.md" title="Türkçe"><img src="../../docs/flags/tr.svg" width="24" alt="Türkçe"></a> <a href="../../docs/readme/README.ja.md" title="日本語"><img src="../../docs/flags/ja.svg" width="24" alt="日本語"></a> <a href="../../docs/readme/README.zh.md" title="简体中文"><img src="../../docs/flags/zh.svg" width="24" alt="简体中文"></a></p>
<p align="center">
  <a href="https://github.com/Yasaiek/Orbek-Discord-Quests-Orbs/releases/latest/download/Orbek.exe"><b>⬇ Скачать для Windows</b></a> &nbsp;·&nbsp;
  <a href="https://yasaiek.github.io/Orbek-Discord-Quests-Orbs/"><b>🌐 Сайт</b></a> &nbsp;·&nbsp;
  <a href="../../CHANGELOG.md">Изменения</a> &nbsp;·&nbsp;
  <a href="https://github.com/Yasaiek/Orbek-Discord-Quests-Orbs/issues/new">Сообщить об ошибке</a>
</p>
<p align="center"><img src="../../docs/img/panel.png" alt="Orbek" width="880"></p>

## Что он умеет

- **Все задания в одном месте** – Игры, видео и активности по вкладкам, с оставшимся временем и кольцами прогресса в реальном времени.
- **Награды сразу видны** – Orbs, украшения аватара, эффекты профиля, предметы в играх – видно, что даёт задание, ещё до начала.
- **Сортировка и фильтры** – Скоро закончатся, больше всего Orbs, самые короткие или новые; только косметика Discord; мини-вид и маленькое окно.
- **Установка в один клик** – Orbek скачивает Vencord, добавляет свой плагин, собирает его и подключает к Discord. Ничего не ставить вручную.
- **Автоматика по желанию** – Orbek может выполнять задания сам, по несколько сразу. По умолчанию выключено – Discord это запрещает, так что на ваш риск.
- **Обновляется сам** – Новые версии берутся из релизов GitHub и проверяются по SHA-256. 12 языков.

## Руководство

1. **Скачайте и запустите Orbek** – Скачайте Orbek.exe со страницы проекта, положите в отдельную папку и запустите. Значок появится возле часов.
2. **Нажмите «Установить всё»** – На вкладке Установка Orbek скачает Node.js и Vencord, добавит плагин, соберёт его и подключит к Discord. Это займёт несколько минут.
3. **Discord перезапустится сам** – После установки Discord запустится снова с плагином Orbek. Внизу окна «Discord подключён» станет зелёным.
4. **Смотрите задания** – Вкладки Игры, Видео и За Orbs, сортировка и фильтр косметики. У каждого задания видны награды, прогресс и оставшееся время.
5. **Принимайте и получайте** – «Принять задание» берёт его. Когда оно выполнено, нажмите «Получить в Discord» – награду получаете на странице заданий (там капча).
6. **Автоматика – на ваш риск** – В Настройках можно разрешить Orbek выполнять задания самостоятельно, по несколько сразу. Discord это запрещает и может забрать награды или заблокировать аккаунт.
7. **Мини-вид** – Переключите вид на «Мини» или откройте «Маленькое окно», чтобы видеть только награды и кольца прогресса.
8. **Обновления** – Orbek обновляется сам с GitHub. Когда выйдет новый плагин, появится полоса «Обновить плагин».

## Вопросы

**Это разрешено?**  
Просматривать и принимать задания – ровно то, что делает сам Discord. Выполнять их автоматически – нарушение правил: Discord может забрать награды или заблокировать аккаунт.

**Видит ли Orbek мой пароль или токен Discord?**  
Нет. Всё идёт через плагин в вашем собственном Discord. Orbek получает от него только список заданий; токен не покидает Discord.

**Что такое Vencord?**  
Популярный мод для приложения Discord на компьютере. В нём работает плагин Orbek; установка ставит его за вас.

**Windows предупреждает о неизвестном издателе.**  
Файл не подписан. Код открыт, и каждый релиз собирается из него в GitHub Actions.

## Сборка из исходников

Go 1.24 или новее:

```
go build -trimpath -ldflags "-H=windowsgui -s -w" -o Orbek.exe .
```

## Отказ от ответственности

Автоматическое выполнение заданий Discord нарушает его правила. Orbek не связан с Discord. Вы используете его на свой риск.
