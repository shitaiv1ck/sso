
# SSO-сервис

gRPC-сервис для аутентификации и авторизации пользователей с использованием JWT-токенов.

## Функционал

- Регистрация новых пользователей c рассылкой события `user.created` через Kafka
- Аутентификация пользователей с выдачей JWT-токенов и refresh-токенов
- Обновление JWT-токенов через refresh-токены
- Логаут с добавлением JWT-токена в черный список в Redis
- Смена пароля и email

## Технологический стэк

- **Golang** — серверная часть
- **gRPC** — протокол взаимодействия между сервисами
- **PostgreSQL** — хранение данных о пользователях, сессиях и приложениях
- **Redis** — хранение черного списка недействительных JWT-токенов
- **Apache Kafka** — рассылка событий
- **Migrate** — управление миграциями базы данных
- **Make** + **Docker Compose** — сборка и запуск проекта

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

# Redis
REDIS_HOST=localhost
REDIS_PORT=6037
REDIS_PASSWORD=redis
REDIS_TIMEOUT=15s
REDIS_DB=0

# Kafka
KAFKA_HOST=localhost
KAFKA_PORT=9092
KAFKA_TIMEOUT=15s

# gRPC сервер
GRPC_ADDR=:50051
GRPC_TIMEOUT=30s

# Логирование
LOG_LEVEL=debug

# JWT и сессии
JWT_KEY=your-secret-key-here
JWT_TTL=15m
SESSION_TTL=720h
```

### 3. Запуск базы данных и миграций

```bash
# Запуск PostgreSQL и Redis в Docker
make env-up

# Применение миграций
make migrate-up
```

### 4. Настройка приложений

Добавьте информацию о приложении в таблицу `sso.apps`:

**Вариант 1: Ручное добавление**
```sql
INSERT INTO sso.apps (name) VALUES ('your_app_name');
```

**Вариант 2: Через миграцию**
```bash
make migrate-create seq=value
# Затем добавьте SQL запрос в созданный файл миграции и выполните make migrate-up
```

### 5. Запуск сервиса

```bash
make sso-run
```

Сервис будет доступен на `localhost:50051` (порт настраивается через `GRPC_ADDR`).

Графический интерфейс для управления Kafka будет доступен на `localhost:9092` (адрес настраивается через `KAFKA_HOST` и `KAFKA_PORT`).

## Взаимодействие с сервисом

Для работы с сервисом используйте клиент, сгенерированный из `sso.proto`:

- **Исходный контракт:** [protos](https://github.com/shitaiv1ck/protos)
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

### Доступные сервисы и RPC-методы

#### Сервис Auth

| Метод | Описание | Запрос | Ответ |
|-------|----------|--------|-------|
| `Register` | Регистрация нового пользователя | `email`, `password` | `user_id` |
| `Login` | Аутентификация пользователя | `email`, `password`, `app_id` | `access_token`, `refresh_token` |
| `Refresh` | Обновление access-токена | `refresh_token`, `app_id` | `access_token`, `refresh_token` |
| `Logout` | Выход из системы | `access_token`, `refresh_token` | `Empty` |

#### Сервис Account

| Метод | Описание | Запрос | Ответ |
|-------|----------|--------|-------|
| `ChangePassword` | Смена пароля пользователя | `user_id`, `old_password`, `new_password` | `Empty` |
| `ChangeEmail` | Смена email пользователя | `user_id`, `new_email`, `password` | `Empty` |

#### Описание методов

**Register** - регистрирует нового пользователя в системе. При успешной регистрации:
- Возвращает `user_id`
- Отправляет событие о регистрации в Kafka (топик `user.created`)

**Login** - аутентифицирует пользователя по email и паролю. При успешном входе:
- Создает сессию в БД с refresh-токеном
- Возвращает `access_token` и `refresh_token`
- `access_token` - JWT-токен для доступа к защищённым ресурсам (время жизни настраивается через `JWT_TTL`)
- `refresh_token` - токен для обновления access-токена (время жизни настраивается через `SESSION_TTL`)

**Refresh** - обновляет пару токенов:
- Принимает действующий `refresh_token` и `app_id`
- Проверяет валидность refresh-токена в БД
- Создает новую сессию и удаляет старую
- Возвращает новую пару `access_token` и `refresh_token`

**Logout** - выполняет выход пользователя из системы:
- Добавляет `access_token` в черный список в Redis (на время его TTL)
- Удаляет сессию с `refresh_token` из БД
- Оба токена становятся недействительными

**ChangePassword** - изменяет пароль пользователя:
- Требует подтверждения старого пароля
- После успешной смены все существующие сессии пользователя остаются активными

**ChangeEmail** - изменяет email пользователя:
- Требует подтверждения паролем
- Проверяет, что новый email не используется другим пользователем

## Структура проекта

```
sso/
├── Makefile                     # Команды для сборки, запуска и миграций
├── README.md                    # Документация проекта
├── docker-compose.yaml          # Docker Compose для всего сервиса
├── go.mod
├── go.sum
├── cmd/
│   └── sso/                     # Точка входа в приложение
├── internal/
│   ├── core/                    # Общие компоненты
│   │   ├── broker/              # Базовый клиент для работы с Kafka
│   │   ├── domain/              # Доменные модели (User, App, Session, Token)
│   │   ├── errors/              # Кастомные ошибки и их классификация
│   │   ├── logger/              # Настройка и обертка над логгером
│   │   ├── repository/          # PostgreSQL и Redis
│   │   ├── transport/           # gRPC-сервер и статусы
│   │   └── validation/          # Валидация входных данных
│   └── features/
│       └── auth/                # Feature: Аутентификация и управление аккаунтом
│           ├── broker/          # Публикация событий в Kafka
│           ├── repository/      # Работа с PostgreSQL и Redis (черный список)
│           ├── service/         # Бизнес-логика (Register, Login, Refresh, Logout)
│           └── transport/       # gRPC-обработчики
└── migrations/                  # SQL миграции для PostgreSQL
```
## Схема базы данных

### Схема `sso`

Все таблицы находятся в схеме `sso` для изоляции данных сервиса.

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

### Таблица `sessions`

Хранит активные сессии пользователей для управления refresh-токенами.

| Поле | Тип | Ограничения | Описание |
|------|-----|-------------|----------|
| `refresh_token` | VARCHAR(255) | `PRIMARY KEY` | Уникальный refresh-токен |
| `user_id` | INT | `NOT NULL`, `REFERENCES sso.users(id) ON DELETE CASCADE` | Идентификатор пользователя |
| `app_id` | INT | `NOT NULL`, `REFERENCES sso.apps(id) ON DELETE CASCADE` | Идентификатор приложения |
| `created_at` | TIMESTAMPTZ | `NOT NULL`, `DEFAULT NOW()` | Время создания сессии |
| `expires_at` | TIMESTAMPTZ | `NOT NULL` | Время истечения сессии |

**Ограничения:**
- `CHECK (expires_at > created_at)` — время истечения должно быть позже времени создания

**Особенности:**
- При удалении пользователя все его сессии автоматически удаляются (CASCADE)
- При удалении приложения все связанные сессии автоматически удаляются (CASCADE)

## Интеграция с Redis

Redis используется для хранения черного списка JWT-токенов:

- При вызове `Logout` access-токен добавляется в Redis с TTL, равным оставшейся продолжительности его жизни
- Автоматическая очистка истекших токенов происходит благодаря встроенному механизму TTL в Redis

## Интеграция с Kafka

При регистрации нового пользователя сервис отправляет событие в Kafka:

**Топик:** `user.created`

**Формат события:**
```json
{
  "user_id": 12345,
  "email": "user@example.com",
}
```

Это позволяет другим микросервисам реагировать на регистрацию новых пользователей (отправка приветственных писем, создание профиля и т.д.).

## Команды Makefile

| Команда | Описание |
|---------|----------|
| `make env-up` | Запуск PostgreSQL, Redis и Kafka в Docker |
| `make env-down` | Остановка всех контейнеров |
| `make migrate-up` | Применение миграций |
| `make migrate-down` | Откат миграций |
| `make migrate-create` | Создание новой миграции (требует `seq=value`) |
| `make sso-run` | Запуск gRPC-сервера |
