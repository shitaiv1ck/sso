# SSO-сервис

gRPC сервис аутентификации и авторизации

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
│  │  ├─ logger
│  │  │  ├─ config.go
│  │  │  └─ logger.go
│  │  ├─ repository
│  │  │  └─ postgres
│  │  │     ├─ config.go
│  │  │     └─ postgres.go
│  │  └─ transport
│  │     └─ grpc
│  │        └─ server
│  │           ├─ config.go
│  │           └─ server.go
│  └─ features
└─ migrations
   ├─ 000001_init.down.sql
   └─ 000001_init.up.sql

```