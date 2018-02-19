Что это?
========

Этот репозиторий служит для разработки и распространения драйвера для интеграции
устройств серии [систем безопасности "Астра"](http://www.teko.biz/) и
[контроллеров автоматизации WirenBoard](http://contactless.ru/).

Проверена работоспособность драйвера совместно со следующими устройствами:
1. [Wiren Board 5 rev 5.8](http://contactless.ru/wiki/index.php/Wiren_Board_5:_%D0%90%D0%BF%D0%BF%D0%B0%D1%80%D0%B0%D1%82%D0%BD%D1%8B%D0%B5_%D1%80%D0%B5%D0%B2%D0%B8%D0%B7%D0%B8%D0%B8)
2. [Wiren Board 3.5](http://contactless.ru/wiki/index.php/Wiren_Board_Smart_Home_3.5)
3. [Астра РИ-М РР](http://www.teko.biz/catalog/823/7006/) с прошивкой `RRs-rim-av3_0`
4. [Различные извещатели](/sensors.md) 

Преверена работа с единственным Астра РИ-М на канале. Теоретически драйвер
должен работать с любым количеством РИ-М на канале (в пределах 250 штук, в связи
со спецификой протокола), а так же с Астра-Z и любыми извещателями от Теко,
совместимыми с Астра РИ-М и Астра-Z.

FYI: Рекомендую покупать Астра-Z, а не Астра РИ-М, из-за того, что у последнего
есть небольшая путаница в радиорежимах (например если вы захотите использовать
датчик протечки и датчик температуры — ничего не выйдет, т.к. первому нужен
радиорежим 1, а второму — 2), в то же время для Астра-Z нет (и не будет) датчика
изменения положения (впрочем и сам этот датчик для РИ-М снят с производства).

Перед началом работы
====================

Устройства Астра-РИ-М выходят с завода в режиме "автономный", он не
предполагает работу в качестве ведомого устройства на линии RS-485, в связи с
этим необходимо переключить РИ-М в режим "системный". Делается это заменой
прошивки, для чего потребуется загрузить и установить
[ПО ПКМ Астра Pro](http://www.teko.biz/support/programms/pc/) (во время
установки убедитесь, что среди модулей ПКМ активирована галочка рядом с модулем
смены ПО), и далее следуйте документации из [Инструкции пользователя](http://www.teko.biz/upload/rukovod/RR-RI-M_%D0%98%D0%BD%D1%81%D1%82%D1%80%D1%83%D0%BA%D1%86%D0%B8%D1%8F%20%D0%BF%D0%BE%D0%BB%D1%8C%D0%B7%D0%BE%D0%B2%D0%B0%D1%82%D0%B5%D0%BB%D1%8F.pdf),
раздел "Смена ПО на РР для работы в режиме «Системный»".

**NB** После обновления прошивки не забудьте обновить прошивку радиомодуля.
И уже после прошивки радиомодуля необходимо сделать очистку памяти, согласно документации
(очистка памяти требует сноровки и ловкости рук — может понадобиться несколько попыток,
прежде чем у вас получится это сделать).

Использование
=============

## Установка
Для удобства установки и обновления, с помощью сервиса [packagecloud.io](https://packagecloud.io) был создан [репозиторий](https://packagecloud.io/wb-mqtt-astra/main/install),
после настройки которого вы сможете установить пакет как обычно, с помощью команды `apt-get install wb-mqtt-astra`.

Так же можно установить всё вручную, из текущего репозитория: найдите последнюю
версию драйвера в разделе [Releases](https://github.com/andrey-yantsen/wb-mqtt-astra/releases/latest)
и загрузите её на WirenBoard командой (не забудьте заменить ссылку на корректную)
```
wget https://github.com/andrey-yantsen/wb-mqtt-astra/releases/download/v0.1/wb-mqtt-astra_0.1_armel.deb
```

Затем установите (имя файла определяется в результате выполнения предыдущей команды):
```
dpkg -i wb-mqtt-astra_0.1_armel.deb
```

## Настройка
Отредактируйте файл `/etc/default/wb-mqtt-astra` в соответствии с вашей
установкой: для корректной работы драйвера достаточно указать опции `serial` (по
умолчанию `/dev/ttyAPP4`) и `address` (при наличии нескольких устройств Астра
РИ-М на одной шине нужно указывать адреса в таков формате: `-address 1 -address 
2 -address 3`). Если вы ещё не регистрировали РИ-М (т.е. не изменяли стандартный
адрес), то значение адреса можно выбрать любое, в диапазоне от 1 до 250.
Чтобы работать с РИ-М, который подключен к порту `/dev/ttyAPP3` и имеет (или будет
иметь) адрес 5 — необходимо установить переменную
`ASTRA_OPTIONS='-serial /dev/ttyAPP3 -address 5'`.
После задания настроек запустите фоновый процесс: `invoke-rc.d wb-mqtt-astra start`.

## Привязка Астра РИ-М и извещателей
Откройте раздел `Devices` в вэб-интерфейсе WirenBoard и найдите блок с
заголовком *astra_1* (где `1` — адрес устройства, указанный при настройке).
Переключите триггер `register` в положение `on`, чтобы зарегистрировать РИ-М.
При успешной регистрации триггер `delete` перейдёт в `off`, а `register` станет
чекбоксом с установленной галочкой.

По умолчанию РИ-М будет работать в частоте "1" (соответстует извещателям с
пометкой *лит. 1*), чтобы изменить канал — необходимо переключить контрол
`l2_channel` в желаемое положение. При изменении частоты все
зарегистрированные извещатели будут забыты.

После регистрации РИ-М можно привязывать извещатели, для этого в том же
интерфейсе нужно активировать триггер `register_l2`, дождаться, пока на РИ-М
замигает светодиод *радиосеть* и установить аккумулятор в регистрируемый
извещатель. После регистрации извещатель появится в списке устройств, а триггер
`register_l2` переключится в состояние `off`.

При регистрации нового извещателя в MQTT публикутеся минимальное количество
контроллов (в интерфейсе WB будет видно только `delete`), для
получения полного списка необходимо заставить извещатель сработать — намочить
датчик протечки, передвинуть датчик изменения положения и т.д.

Видео-инструкция доступна по ссылке [http://take.ms/xpOEm](http://take.ms/xpOEm).

Часть контролов отображаются в 2 экземплярах — "подтверждённая" и "не
подтверджённая" тревога, например `Channel1` и `Channel1_confirmed` для датчика
протечки. "Подтверждённая" тревога меняет своё состояние через несколько секунд
после изменения "не подтверждённой". К сожалению мне не удалось добиться
стабильной отправки с извещателей событий об изменении состояния подтверждённой
тревоги, в связи с этим я рекомендую её игнорирвать.

## MQTT
Обработка и заполнение MQTT происходит в соответствии с [соглашением WirenBoard](https://github.com/contactless/homeui/blob/master/conventions.md).
Список контроллов и описание выполняемых ими действий доступны в файле [controls.md](controls.md).

## Удаление зарегистрированого РИ-М / Z
Чтобы удалить устройство с адресом 1 запустите следующую команду в консоли:
```
wb-mqtt-astra -address 1 -serial /dev/ttyAPP4 -delete-device
```

После удаления РИ-М / Z перестают отвечать по зарегистирированному ранее адресу
и отвечают только на широковещательные запросы (поиск устройств, регистрация).
После удаления необходимо очистить MQTT, удалив как информацию об удалённом
РИ-М, так и данные о всех связанных извещателях (не забудьте подставить
корректный адрес РИ-М и номер извещателя, и повторите вторую команду необходимое
количество раз, в зависимости от количества зарегистрированных извещателей):

```
mqtt-delete-retained '/devices/astra_1/#'
mqtt-delete-retained '/devices/astra_1_sensor_1/#'
```

## Troubleshooting (оформление ошибок)
При возникновении нештатных ситуаций (вылетает демон, не работают какие-то
функции, и т.п.) необходимо правильно составить баг-репорт, в противном случае
помочь вам в решении данной проблемы у меня не получится.

1. В первую очередь обновите демон до последней версии и убедитесь, что проблема
   всё ещё не исправлена
2. Проверьте корректность подключения Астра РИ-М / Z и WirenBoard
3. Выполните очистку памяти РИ-М / Z согласно документации (только при проблемах
   связи между Астра и WB)
4. Если после выполнения первых 3 пунктов проблема не устранилась, то остановите
   демоны `wb-mqtt-serial` и `wb-mqtt-astra`:
```bash
invoke-rc.d wb-mqtt-serial stop
invoke-rc.d wb-mqtt-astra stop
```
5. Запустите демон wb-mqtt-astra в отладочном режиме, с записью лога в файл: 
   `. /etc/default/wb-mqtt-astra && wb-mqtt-astra -debug ${ASTRA_OPTIONS} >astra-debug.log 2>&1`.
   После того, как в интерфейсе проделаны все действия для повторения проблемы —
   завершите процесс нажатием Ctrl+C. Просмотрите получившийся лог-файл
   `astra-debug.log` и убедитесь, что проблема вызвана не ошибкой
   конфигурирования
6. Опубликуйте полученный лог-файл в виде [gist](https://gist.github.com)
7. Создайте [issue](https://github.com/andrey-yantsen/wb-mqtt-astra/issues) с
   детальным описанием проблемы (как минимум — полностью опишите все шаги,
   проделанные для воспроизведения проблемы, так же опишите, что должно было
   получиться в результате выполнения этих действий, и что на самом деле
   получилось) и добавьте ссылку на gist из предыдущего пункта.
