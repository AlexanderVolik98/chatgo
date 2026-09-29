# ChatGo

Чат с комнатами, ролями и realtime-сообщениями через WebSocket.

> Проект в разработке, сейчас это MVP.

**Стек:** Go (net/http, PostgreSQL, WebSocket), Vue 3 (Vite, Pinia), Docker Compose.

## Возможности

- Регистрация и вход (JWT)
- Публичные и приватные комнаты
- Роли в комнате: `owner` → `admin` → `member`
- История сообщений с пагинацией и live-обновления по WebSocket

## Запуск

```bash
cp .env.example .env
docker compose up --build
```

Фронтенд: http://localhost:8081, бэкенд: http://localhost:8080.

Для локальной разработки: `make backend-dev` и `make frontend-dev`. Тестовые данные: `make seed`.

## Структура

```
backend/
  cmd/server           точка входа
  internal/httpapi     REST-хендлеры и роутинг
  internal/ws          WebSocket hub
  internal/repository  доступ к БД
  internal/db          подключение и миграции
  internal/auth        JWT, bcrypt
frontend/              Vue 3 SPA
```

## Разработка с Claude Code

Я активно использую [Claude Code](https://claude.com/claude-code) в разработке, но осознанно, а не как генератор кода «под ключ». Работаем в формате парного программирования с взаимным ревью: код, который пишет Claude, я читаю, проверяю и правлю, а свой код отдаю ему на ревью. В репозиторий попадает только то, что я понимаю и могу объяснить, и за каждую строчку отвечаю я, а не ИИ.

Правила работы с ассистентом описаны в [CLAUDE.md](CLAUDE.md).
