Данный программный код реализует сервис единой точки доступа (gateway) к информационной системе, реализованной в микросервисной архитектуре. Сервис принимает все http - запросы к системе, проверяет права доступа для каждого запроса и проксирует запрос на тот или иной компонент информационной системы.

Сервис gateway выполняет следующие задачи:

* авторизация запросов к системе
* проксирование запроса на соответствующий компонент системы в зависимости от реазультатов авторизации
* защита от bruteforce атак на систему с помощью ограничения количества запросов с одного ip-адреса
* приём от внутренних компонентов системы безопасности сообщений о попытках подбора пароля и блокирование ip-адресов, с которых осуществляются такие попытки
* отправка администратору системы сообщения email c предупреждением об инцидентах, связанных с безопасностью

Сервис gateway выполняет запросы к сервису централизованного логирования (https://github.com/NemanCat/logging) и сервису отправки email (https://github.com/NemanCat/mailer).

Документация по сервису gateway:  https://github.com/NemanCat/gateway/wiki

