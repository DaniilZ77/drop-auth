# Drop Authorization Service [![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)


Микросервис для аутентификации и авторизации пользователей.

## Основные возможности

- **Ролевая модель**:
  - Пользователи (`user`)
  - Администраторы (`minor/major`)
- Фильтрация пользователей по параметрам
- Админ-токены с расширенными правами
- Авторизация через Telegram + JWT
- CRUD операции с пользователями


## Стек

- Go
- PostgreSQL (pgx + sqlc)
- Redis (для хранения refresh-токенов)
- Docker
- gRPC + gRPC Gateway


## Запуск сервиса

1. Клонируйте репозиторий:
   ```bash
   $ git clone https://github.com/MAXXXIMUS-tropical-milkshake/drop-auth.git

   $ cd drop-auth
   ```

2. Запустите сервис:
   ```bash
   $ docker compose up
   ```

## API

| Метод | Эндпоинт                      | Требуемая роль | Описание                  |
|-------|-------------------------------|----------------|---------------------------|
| GET   | `/health`         | `-`        | Проверка доступности сервиса  |
| POST   | `/v1/admin/init`    | `-`        | Создание `major` админа (запрос только с `localhost`)           |
| POST   | `/v1/admin`    | `major admin`   | Добавление `minor` админа по `username` (нужен `jwt` токен)  |
| DELETE | `/v1/admin`    | `major admin`        | Удаление `minor` админа по `username` (нужен `jwt` токен)      |
| POST| `/v1/auth/login`    | `-`   | Создание пользователя и выдача токенов (нужен `telegram mini apps` токен авторизации)     |
| POST| `/v1/auth/token/refresh`    | `-`   | Ротация `jwt` токенов (нужен `refresh token`)     |
| PATCH| `/v1/user`    | `-`   | Обновление пользователя (нужен `access token`, где хранится `id` пользователя)     |
| GET| `/v1/users`    | `-`   | Множественная фильтрация пользователей по различным параметрам     |

## База данных

Схема базы данных находится на следующем ресурсе:

- https://dbdiagram.io/d/drop-auth-67b3a7fa263d6cf9a07aeb87