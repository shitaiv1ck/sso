# SSO-сервис

gRPC сервис аутентификации и авторизации

## Функционал

- Аутентификация и авторизация пользоваталей в системе при помощи JWT токенов

## Как запустить

1. Склонируйте репозиторий:

```bash
git clone https://github.com/shitaiv1ck/sso.git
```

2. Настройте переменные окружения в `.env` в корневой папке проекта по примеру `.env.example`:

```env
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=sso
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_TIMEOUT=15s

GRPC_ADDR=:50051
GRPC_TIMEOUT=30s

LOG_LEVEL=debug
```

3. Запустите:

```bash
make env-up && \
make migrate-up
```

4. Вручную добоавьте информацию о приложении(id, название) в БД в таблицу `sso.apps`, либо с помощью миграции, предварительно создав ее при помощи:

```bash
make migrate-create seq=value
```

5. В `.env` настройте JWT:

```env
JWT_EXAMPLE_KEY=your-secret-key-here
JWT_TTL=12h
```

Примечание: Вместо `EXAMPLE` впишите название приложения, которого вы сохранили в БД, в верхнем регистре

6. Запустите:

```bash
make sso-run
```

Для взамодействия с сервисом необходимо реализовать клиента из `sso.proto` файла из репозитория [**protos**](https://github.com/shitaiv1ck/protos)

Сервис будет доступен на localhost:50051 по умолчанию

P.S: в качестве клиента можно использовать Postman, скормив ему тот же `sso.proto` файл

## Контракты

Данный SSO-сервис использует репозиторий [**protos**](https://github.com/shitaiv1ck/protos) как источник контрактов

Для обновления контрактов используйте:

```bash
go get -u github.com/shitaiv1ck/protos@latest
```

## Структура проекта

```
sso
├─ Makefile
├─ README.md
├─ cmd
│  └─ sso
│     └─ main.go
├─ docker-compose.yaml
├─ go.mod
├─ go.sum
├─ internal
│  ├─ core
│  │  ├─ domain
│  │  │  ├─ app.go
│  │  │  └─ user.go
│  │  ├─ errors
│  │  │  └─ errors.go
│  │  ├─ logger
│  │  │  ├─ config.go
│  │  │  └─ logger.go
│  │  ├─ repository
│  │  │  └─ postgres
│  │  │     ├─ config.go
│  │  │     └─ postgres.go
│  │  ├─ transport
│  │  │  └─ grpc
│  │  │     ├─ server
│  │  │     │  ├─ config.go
│  │  │     │  └─ server.go
│  │  │     └─ status
│  │  │        └─ status.go
│  │  └─ validation
│  │     └─ validation.go
│  └─ features
│     └─ auth
│        ├─ repository
│        │  └─ repository.go
│        ├─ service
│        │  └─ service.go
│        └─ transport
│           └─ grpc
│              └─ transport.go
└─ migrations
   ├─ 000001_init.down.sql
   └─ 000001_init.up.sql

```

## Схема БД

### Таблица users

| Поле | Тип | Ограничения | Описание |
|------|-----|-------------|----------|
| id | INT | GENERATED ALWAYS AS IDENTITY, PRIMARY KEY | Уникальный идентификатор пользователя |
| email | VARCHAR(255) | NOT NULL, UNIQUE | Электронная почта пользователя (используется для входа) |
| pass_hash | VARCHAR(255) | NOT NULL | Хеш пароля пользователя |

**Ограничения:**
- CHECK(email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$') — проверка формата email

**Индексы:**
- idx_users_email — по полю email (для быстрого поиска)

### Таблица apps

| Поле | Тип | Ограничения | Описание |
|------|-----|-------------|----------|
| id | INT | GENERATED ALWAYS AS IDENTITY, PRIMARY KEY | Уникальный идентификатор приложения |
| name | VARCHAR(255) | NOT NULL, UNIQUE | Название приложения |

**Ограничения:**
- CHECK (char_length(name) BETWEEN 2 AND 255) — длина названия от 2 до 255 символов

**Индексы:**
- idx_apps_name — по полю name (для быстрого поиска)