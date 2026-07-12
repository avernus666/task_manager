# Task Manager API

RESTful API для управления проектами, задачами и комментариями с JWT-аутентификацией и ролями (пользователь / администратор).

## Возможности

- Регистрация и вход (JWT)
- Управление проектами (CRUD)
- Управление задачами (CRUD) с фильтрацией по проекту, статусу, приоритету, исполнителю
- Комментарии к задачам
- Роли: пользователь (видит только свои проекты), администратор (видит всё)
- Swagger-документация (`/docs`)

## Стек

- Python 3.10+
- FastAPI
- SQLAlchemy
- SQLite / PostgreSQL
- JWT (python-jose)
- bcrypt (passlib)

## Запуск

1. Клонируйте репозиторий
2. Установите зависимости: `pip install -r requirements.txt`
3. Создайте `.env` (см. пример)
4. Запустите: `uvicorn app.main:app --reload`
5. Откройте Swagger: `http://localhost:8000/docs`

## Примеры запросов

### Регистрация
`POST /auth/register`
```json
{
  "username": "john",
  "email": "john@example.com",
  "password": "secret"
}Логин
POST /auth/login (form-data, поля username и password)

Создание проекта
POST /projects (требуется токен)

json
{
  "name": "Мой проект",
  "description": "Описание"
}
Создание задачи
POST /tasks (требуется токен)

json
{
  "title": "Сделать отчёт",
  "project_id": 1,
  "assignee_id": 2
}
Фильтрация задач
GET /tasks?project_id=1&status=new&priority=high

Контакты
Telegram: @Yosu354
Email: seijuro6665@gmail.com

## ▶️ Запуск

```bash
cd task_manager
pip install -r requirements.txt
uvicorn app.main:app --reload