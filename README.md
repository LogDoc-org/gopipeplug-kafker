Version: 0.0.1

Плагин сценария слушателя, реализующий отправку события в топик Кафки на каждое срабатывание слушателя.

Реализуем 2 метода из вышеуказанного интерфейса:

1. configure - метод конфигурации плагина, вызывается при старте управляющего модуля, если в его конфиге обнаружен плагин. Установка различных параметров, в нашем случае - чтение параметров Kafka из конфигурации управляющего модуля
2. fire - метод, который срабатывает при возникновении событий, на которые настроен слушатель

Пример селектора для того, чтобы слушатель сработал и отправил событие в Кафка:

{

&nbsp;&nbsp;&nbsp;&nbsp;"live": false,

&nbsp;&nbsp;&nbsp;&nbsp;"watcher": true,

&nbsp;&nbsp;&nbsp;&nbsp;"join": "AND",

&nbsp;&nbsp;&nbsp;&nbsp;"type": "Query",

&nbsp;&nbsp;&nbsp;&nbsp;"eventCounter": 1,

&nbsp;&nbsp;&nbsp;&nbsp;"repeatable": true,

&nbsp;&nbsp;&nbsp;&nbsp;"repeatLimit": 1,

&nbsp;&nbsp;&nbsp;&nbsp;"childs": [

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"sample": "DEBUG",

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"coOp": "equal",

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"type": "LevelMatch"

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;}

&nbsp;&nbsp;&nbsp;&nbsp;],

&nbsp;&nbsp;&nbsp;&nbsp;"watchersPlugins": [

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"id": 1,

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"params": []

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;},

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"id": 3,

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"params": []

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;}

&nbsp;&nbsp;&nbsp;&nbsp;]

}



Разберем каждый блок отдельно:

Срабатываем, при получении лога с типом DEBUG

"childs": [

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"sample": "DEBUG",

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"coOp": "equal",

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"type": "LevelMatch"

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;}

&nbsp;&nbsp;&nbsp;&nbsp;]



Счетчики срабатываний:

eventCounter - это порог, когда срабатывает плагин

т.е. сколько "подошедших" событий должно случиться, чтобы в плагине запустился метод watcherActuated



Например: ты выставляешь условия селектора, типа "источник должен быть 'com.sand.job' и уровнеь лога должен быть = 'DEBUG'"

каждое событие, подходящее под эти условия, будет инкрементить счётчик на единичку



Предположим, при настройке плагина ты выставил eventCounter = 5

это означает, что 4 таких события пролетят просто так, а на 5-е - твой плагин выполнит метод  watcherActuated и отработает



Затем, если у тебя стоит repeatable = true, то всё вышеописанное будет считаться "циклом"

после срабатывания плагина счётчик событий сбросится на 0, счётчик "циклов" увеличится на единичку и погнали далее по кругу

Когда счётчик циклов дойдёт до значения repeatLimit — плагин закончит работу и будет выгружен из наблюдения



Если нужно срабатывание на каждое событие без ограничений, то оба счётчика надо в 0 выставить

тогда будет срабатывать просто на каждое подходящее под условия



"watchersPlugins": [

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"id": 1,

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"params": []

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;},

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;{

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"id": 3,

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;"params": []

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;}

&nbsp;&nbsp;&nbsp;&nbsp;]

1 - Это стандартный плагин entrypipes.WsNotifier, перенаправляет логи в WebSocket, прописан в конфигурации Управляющего модуля

3 - Это наш плагин Kafker, добавился в конец списка при сканировании конфигурации

Пример kafker.conf конфигурации для подключения нового плагина:

```
include "application.conf"

logdoc {
  plugins {
    pipes {

      enabled += "ru.gang.logdoc.plugins.Kafker"

      ru.gang.logdoc.plugins.Kafker {
        title = "Отправка метрик в Kafka"
        description = "Уведомляет создателя слушателя о каждом срабатывании слушателя, отправляя событие в Kafka"

        kafka {
          topic = "kafker"

          properties {
            bootstrap.servers = "localhost:9092" # Сервера Кафка
          }
        }

        params = [
        ]
      }

    }
  }
}
```