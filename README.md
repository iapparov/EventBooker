
# EventBooker

## Описание

EventBooker – веб-сервис для управления бронированиями на мероприятия (концерты, мастер-классы, курсы, лекции) с автоматическим контролем времени подтверждения брони.

Сервис решает проблему «мертвых броней»: если пользователь не подтвердил или не оплатил бронь в установленный срок, место автоматически освобождается для других участников.

### Основные возможности:
- Создание мероприятий: задаются название, дата, количество мест, время жизни брони.
- Бронирование мест: пользователи могут бронировать места на мероприятие.
- Подтверждение брони: бронь необходимо подтвердить в течение указанного времени.
- Автоматическая отмена: фоновый процесс удаляет неоплаченные/неподтвержденные брони.
- Просмотр мероприятий и статусов: отображение свободных мест и текущих бронирований.

### Дополнительные функции:
- Уведомления об отмене брони через Email или Telegram.
- Поддержка регистрации и аутентификации пользователей.
- Возможность настраивать индивидуальный срок жизни брони для разных мероприятий.

Цель сервиса — облегчить управление бронированиями, минимизировать «мертвые» места и упростить взаимодействие организаторов с участниками.


## Структура проекта

```
cmd/
  eventbooker/main.go            — точка входа, инициализация и запуск приложения

internal/
  app/app.go                     — сборка зависимостей, HTTP-сервер, graceful shutdown
  config/config.go               — загрузка конфигурации из YAML + env

  domain/                        — доменные модели
    booking/booking.go
    event/event.go
    user/user.go

  service/                       — бизнес-логика
    booking.go                   — создание и подтверждение бронирований
    event.go                     — создание и получение мероприятий
    user.go                      — регистрация, логин, валидация

  repository/postgres/           — слой хранения (PostgreSQL)
    postgres.go                  — подключение, пул соединений, query timeout
    event.go                     — CRUD для событий и бронирований
    user.go                      — CRUD для пользователей

  auth/jwt.go                    — генерация и валидация JWT (access + refresh)

  broker/rabbit/                 — интеграция с RabbitMQ
    rabbit.go                    — подключение, декларация exchange/queue
    producer.go                  — публикация сообщений в delay-очередь
    consumer.go                  — обработка истёкших бронирований

  notification/                  — отправка уведомлений
    email.go                     — SMTP
    telegram.go                  — Telegram Bot API

  transport/http/                — HTTP-слой
    router.go                    — маршрутизация
    dto/                         — request/response структуры
    handler/                     — обработчики запросов
    middleware/auth.go           — JWT-мидлварь

config/local.yaml                — конфигурация приложения
migrations/                      — SQL-миграции для PostgreSQL
docs/                            — Swagger-документация
web/index.html                   — простой веб-интерфейс
docker-compose.yml               — PostgreSQL + RabbitMQ
```

## Быстрый старт

### 1. Запуск инфраструктуры

```sh
docker-compose up -d
```

Поднимет контейнеры: PostgreSQL (порт 5433), RabbitMQ (порт 5672 / UI 15672).

### 2. Переменные окружения

Скопировать `.env.example` → `.env` и заполнить креды (Postgres, RabbitMQ, JWT-секреты, Telegram-токен, SMTP).

### 3. Миграции

```sh
migrate -path migrations -database "postgres://user:password@localhost:5433/dbname?sslmode=disable" up
```

### 4. Запуск

```sh
go run ./cmd/eventbooker/main.go
```

Сервис стартует на `localhost:8080`.

## API

| Метод | Путь | Описание | Авторизация |
|-------|------|----------|-------------|
| POST | `/api/auth/register` | Регистрация | — |
| POST | `/api/auth/login` | Логин, получение JWT | — |
| POST | `/api/auth/refresh` | Обновление токенов | — |
| POST | `/api/events` | Создание мероприятия | Bearer |
| GET | `/api/events/{id}` | Информация о мероприятии | Bearer |
| POST | `/api/events/{id}/book` | Бронирование места | Bearer |
| POST | `/api/events/{id}/confirm` | Подтверждение (оплата) брони | Bearer |

Swagger: [http://localhost:8080/api/swagger/](http://localhost:8080/api/swagger/)

Все защищённые эндпоинты требуют заголовок `Authorization: Bearer <token>`.

## Веб-интерфейс

Открыть `web/index.html` в браузере — страница для регистрации, создания мероприятий и бронирования мест.

## Тесты

```sh
go test ./internal/...
```

## Миграции

| Файл | Описание |
|------|----------|
| `000001_create_user_table.up.sql` | Таблица пользователей |
| `000002_create_event_table.up.sql` | Таблица мероприятий |
| `000003_create_booking_table.up.sql` | Таблица бронирований |

Для каждой миграции есть соответствующий `.down.sql`.

## Конфигурация

Основная конфигурация — `config/local.yaml`. Креды и секреты подтягиваются из переменных окружения (`.env`).

Ключевые параметры:
- `db_config.query_timeout` — таймаут на SQL-запросы (по умолчанию 5s).
- `retry_strategy` — количество попыток, задержка и коэффициент backoff для запросов к БД и RabbitMQ.
- `event_config.booking_ttl` — допустимые значения TTL брони (в минутах).
- `jwt.jwt_exp_access_token` / `jwt.jwt_exp_refresh_token` — время жизни access/refresh токенов.

## Зависимости

- Go 1.25+
- PostgreSQL 16+
- RabbitMQ 3.13+
- Docker (для инфраструктуры)

## Логирование

Используется `wbf/zlog` (zerolog). Уровень задаётся через `logger.level` в конфиге.

## Swagger

Документация генерируется через `swag` и доступна по `/api/swagger/`.
