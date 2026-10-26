# Task Manager Rework

Переработанная версия Task Manager: основной backend на Go, PostgreSQL и React + TypeScript frontend. Rust оставлен как возможный дополнительный инструмент для отдельных высоконагруженных компонентов — в текущем проекте он намеренно не используется без необходимости.

## Стек

- **Backend:** Go 1.23+, Gin, pgx, JWT, bcrypt, zap.
- **Database:** PostgreSQL 16 + SQL migrations.
- **Frontend:** React, TypeScript, Vite, Lucide icons.
- **Dev/Deploy:** Docker Compose, GitHub Actions.

## Функциональность

- Регистрация и вход, JWT bearer-токены.
- Просмотр профиля текущего пользователя.
- Роли `user` и `admin`; администратор может просматривать список пользователей и все проекты.
- CRUD проектов.
- CRUD задач, фильтры по проекту, статусу, приоритету и исполнителю.
- Комментарии к задачам.
- Проверка доступа на сервере: пользователь работает только со своими проектами.
- Health endpoints: `/health/live` и `/health/ready`.
- Адаптивный интерфейс: обзор, проекты, задачи, статусы, комментарии.
- Автоматические базовые unit-тесты и CI workflow.

## Запуск через Docker Compose

1. Скопируй `.env.example` в `.env`.
2. Измени `JWT_SECRET` на случайный секрет длиной минимум 32 символа.
3. Запусти:

   ```bash
   docker compose up --build
   ```

Compose поднимет PostgreSQL, применит миграции, запустит API и frontend.

- Frontend: http://localhost:5173
- API: http://localhost:8080
- Live check: http://localhost:8080/health/live
- Readiness: http://localhost:8080/health/ready

Для остановки: `docker compose down`. Чтобы также удалить данные БД: `docker compose down -v`.

## Локальный запуск без Docker

Нужны Go 1.23+, PostgreSQL и `golang-migrate`.

```bash
cp .env.example .env
# Укажи DATABASE_URL и JWT_SECRET в .env
docker compose up -d postgres
```

Применить миграции:

```bash
export DATABASE_URL='postgres://task_manager:change_this_password@localhost:5432/task_manager?sslmode=disable'
migrate -path db/migrations -database "$DATABASE_URL" up
```

Запустить backend:

```bash
go mod tidy
go run ./cmd/api
```

Запустить frontend в другом терминале:

```bash
cd frontend
npm install
npm run dev
```

## API endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/auth/register` | Регистрация: `username`, `email`, `password` |
| `POST` | `/auth/login` | Вход: `username`, `password`; возвращает JWT |
| `GET` | `/users/me` | Профиль текущего пользователя |
| `GET` | `/users` | Список пользователей для администратора |
| `GET/POST` | `/projects/` | Список / создание проектов |
| `GET/PUT/DELETE` | `/projects/:id` | Получение / изменение / удаление проекта |
| `GET/POST` | `/tasks/` | Список / создание задач |
| `GET/PUT/DELETE` | `/tasks/:id` | Получение / изменение / удаление задачи |
| `POST` | `/comments/` | Комментарий: `task_id`, `text` |
| `GET` | `/comments/task/:task_id` | Комментарии задачи |

Фильтры задач: `project_id`, `status` (`new`, `in_progress`, `completed`), `priority` (`low`, `medium`, `high`), `assignee_id`.

## Создание администратора

Регистрация создаёт обычного пользователя. После регистрации назначь роль вручную в PostgreSQL:

```sql
UPDATE users SET role = 'admin' WHERE username = 'your_username';
```

Затем войди заново, чтобы получить JWT с обновлённой ролью.

## Проверки

```bash
gofmt -w ./cmd ./internal
go test ./...
go vet ./...
```

Для frontend:

```bash
cd frontend
npm install
npm run build
```

## Примечания

- Не коммить `.env`, реальные пароли или токены.
- В `.env.example` находятся только демонстрационные значения.
- Перед публичным production-развёртыванием добавь rate limiting, refresh/revoke tokens, полноценные интеграционные тесты, строгую настройку CORS и централизованную обработку ошибок.
- Этот проект сохраняет основную функциональность исходного FastAPI-приложения, но использует новую схему PostgreSQL. Автоматической миграции старых данных из SQLite не предусмотрено.
