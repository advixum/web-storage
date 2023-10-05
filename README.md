# Web Storage
Демонстрационная версия доступна по адресу: https://advixum.freemyip.com:1111/

Данное приложение предоставляет функционал дистанционного хранения файлов с управлением через веб-интерфейс. Пользователи могут регистрировать аккаунты, входить в систему, загружать новые, скачивать существующие файлы, переименовывать и удалять их. Приложение реализует рендеринг страниц на стороне клиента и взаимодействие с сервером через API. При успешной аутентификации пользователю выдается JWT токен. Операции на главной странице доступны только при его наличии, в противном случае клиент перенаправляет на страницу входа. Пароли пользователей в базе данных хешированы адаптивным алгоритмом bcrypt. Реализована логика асинхронной обработки загружаемых файлов с ограничением одновременно выполняемых операций.

Используемые технологии: Gin (API), GORM, PostgreSQL, JWT Auth, Logrus, Testify, HTML, CSS, React, Axios
