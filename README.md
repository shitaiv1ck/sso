# SSO-сервис

gRPC-сервис для аутентификации и авторизации пользователей с использованием JWT-токенов.

## Функционал

- Регистрация новых пользователей
- Аутентификация пользователей с выдачей JWT-токенов
- Авторизация запросов на основе JWT-токенов

## Требования

- Go 1.25+
- Docker & Docker Compose
- Make

## Быстрый старт

### 1. Клонирование репозитория

```bash
git clone https://github.com/shitaiv1ck/sso.git
cd sso
```

### 2. Настройка окружения

Создайте файл `.env` в корневой директории проекта на основе `.env.example`:

```env
# PostgreSQL
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=sso
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_TIMEOUT=15s

# gRPC сервер
GRPC_ADDR=:50051
GRPC_TIMEOUT=30s

# Логирование
LOG_LEVEL=debug
```

### 3. Запуск базы данных и миграций

```bash
# Запуск PostgreSQL в Docker
make env-up

# Применение миграций
make migrate-up
```

### 4. Настройка приложений

Добавьте информацию о приложении в таблицу `sso.apps`:

**Вариант 1: Ручное добавление**
```sql
INSERT INTO apps (name) VALUES ('your_app_name');
```

**Вариант 2: Через миграцию**
```bash
make migrate-create seq=value
# Затем добавьте SQL зарос в созданный файл миграции и выполните make migrate-up
```

### 5. Настройка JWT

Добавьте в `.env` конфигурацию JWT для вашего приложения:

```env
JWT_YOUR_APP_NAME_KEY=your-secret-key-here
JWT_TTL=12h
```

> **Важно:** `YOUR_APP_NAME` замените на название приложения в верхнем регистре, которое вы добавили в БД.

### 6. Запуск сервиса

```bash
make sso-run
```

Сервис будет доступен на `localhost:50051` (порт настраивается через `GRPC_ADDR`).

## Взаимодействие с сервисом

Для работы с сервисом используйте клиент, сгенерированный из `sso.proto`:

- **Исходный контракт:** [protos репозиторий](https://github.com/shitaiv1ck/protos)
- **Локальный файл:** `sso.proto` (в корне проекта)

**Способы тестирования:**
1. Реализуйте gRPC-клиента на вашем языке программирования
2. Используйте Postman с импортированным `sso.proto` файлом

## Контракты

Сервис использует [protos](https://github.com/shitaiv1ck/protos) как источник контрактов.

**Обновление контрактов:**
```bash
go get -u github.com/shitaiv1ck/protos@latest
```

### Доступные RPC-методы

| Метод | Описание | Запрос | Ответ |
|-------|----------|--------|-------|
| `Register` | Регистрация нового пользователя | `email`, `password` | `user_id` |
| `Login` | Аутентификация пользователя | `email`, `password`, `app_id` | `token` (JWT) |

## Структура проекта

```
sso/
├── cmd/
│   └── sso/              # Точка входа
│       └── main.go
├── internal/
│   ├── core/             # Общие компоненты
│   │   ├── domain/       # Доменные модели (User, App)
│   │   ├── errors/       # Обработка ошибок
│   │   ├── logger/       # Настройка логирования
│   │   ├── repository/   # Работа с БД (PostgreSQL)
│   │   ├── transport/    # gRPC сервер и статусы
│   │   └── validation/   # Валидация данных
│   └── features/
│       └── auth/         # Feature: Аутентификация
│           ├── repository/   # Репозиторий для Auth
│           ├── service/      # Бизнес-логика
│           └── transport/    # gRPC обработчики
├── migrations/           # SQL миграции
│   ├── 000001_init.down.sql
│   └── 000001_init.up.sql
├── docker-compose.yaml   # Docker Compose для всего сервиса
├── Makefile             # Команды для сборки и запуска
├── go.mod
└── go.sum
```

## Схема базы данных

### Таблица `users`

Хранит информацию о пользователях системы.

| Поле | Тип | Ограничения | Описание |
|------|-----|-------------|----------|
| `id` | INT | `GENERATED ALWAYS AS IDENTITY`, `PRIMARY KEY` | Уникальный идентификатор пользователя |
| `email` | VARCHAR(255) | `NOT NULL`, `UNIQUE` | Электронная почта (используется для входа) |
| `pass_hash` | VARCHAR(255) | `NOT NULL` | Хеш пароля (алгоритм bcrypt) |

**Ограничения:**
- `CHECK (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')` — валидация формата email

**Индексы:**
- `idx_users_email` — для быстрого поиска по email

### Таблица `apps`

Хранит информацию о приложениях, которые могут использовать сервис.

| Поле | Тип | Ограничения | Описание |
|------|-----|-------------|----------|
| `id` | INT | `GENERATED ALWAYS AS IDENTITY`, `PRIMARY KEY` | Уникальный идентификатор приложения |
| `name` | VARCHAR(255) | `NOT NULL`, `UNIQUE` | Название приложения |

**Ограничения:**
- `CHECK (char_length(name) BETWEEN 2 AND 255)` — название от 2 до 255 символов

**Индексы:**
- `idx_apps_name` — для быстрого поиска по названию

---

## Команды Makefile

| Команда | Описание |
|---------|----------|
| `make env-up` | Запуск PostgreSQL в Docker |
| `make env-down` | Остановка PostgreSQL |
| `make migrate-up` | Применение миграций |
| `make migrate-down` | Откат миграций |
| `make migrate-create` | Создание новой миграции (требует `seq=value`) |
| `make sso-run` | Запуск gRPC-сервера |

---
