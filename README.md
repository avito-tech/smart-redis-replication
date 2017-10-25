# Smart Redis Replication

[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg)](LICENSE.md)

Это библиотека для подключения к redis-серверу в качестве slave и разбора всех данных репликации.

В компании Avito использовалась для синронизации новой версии сервиса, с помощью неё было перелито порядка N ключей из старого кластера в новый сервис.

    Направление данных:

    users requests -> old service -> redis cluster

    redis cluster -> smart-redis-replication -> service -> redis cluster

    В процессе переноса были:

    1. модифицированы ключи

    2. отсеяны более не требующиеся данные

    3. заново сгенерированы дополнительные данные (на уровне нового сервиса)

Схема переноса данных не потребовала остановки обслуживания клиентов и позволила поддерживать два сервиса в синхронном состоянии достаточное количество времени, что бы провести тесты и подготовиться к переключению пользовательских запросов в новый сервис.

Библиотека имеет встроенный Backlog, благодаря которому можно обновлять сервис будучи в синхронном состоянии без потери синхронизации.
Данные накапливаются в backlog и отправляются в сервис как только тот будет перезапущен, а благодаря rolling update в kubernetes это происходит и вовсе незаметно для пользователей.

## Поддерживаются форматы:

    rdb - описание формата https://rdb.fnordig.de/file_format.html

    resp - описание формата https://redis.io/topics/protocol

## Поддерживаются типы ключей:

    Sorted Set

    Integer Set

    Set

    Map

    List

    String

## Поддерживаются типы данных:

    0 = String Encoding

    1 = List Encoding

    2 = Set Encoding

    3 = Sorted Set Encoding

    4 = Hash Encoding

    9 = Zipmap Encoding

    10 = Ziplist Encoding

    11 = Intset Encoding

    12 = Sorted Set in Ziplist Encoding

    13 = Hashmap in Ziplist Encoding (Introduced in RDB version 4)

    14 = List in Quicklist encoding (Introduced in RDB version 7)

## Installation

    $ go get github.com/avito-tech/smart-redis-replication

## Examples

    $ ls ./example

## Author

[Oleg Shevelev][mantyr]

[mantyr]: https://github.com/mantyr

