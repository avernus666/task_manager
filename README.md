<div align="center">

# ✦ TASK MANAGER REWORK

### Космическое пространство для проектов, задач и командной работы

**Go · PostgreSQL · React · TypeScript · Neon · Render · Vercel**

![Go](https://img.shields.io/badge/Go-1.23%2B-7c6cff?style=for-the-badge&logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-TypeScript-9d7bff?style=for-the-badge&logo=react&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Neon-6f55d9?style=for-the-badge&logo=postgresql&logoColor=white)
![Security](https://img.shields.io/badge/Security-Hardened-5bdbb0?style=for-the-badge&logo=shielddotio&logoColor=white)

Тёмная космическая эстетика, фиолетовое свечение и единое рабочее пространство для твоих проектов.

</div>

---

## ✨ Возможности

| Раздел | Возможности |
|---|---|
| **Аккаунт** | Регистрация, вход, JWT-аутентификация, профиль |
| **Проекты** | Создание, просмотр и управление собственными проектами |
| **Задачи** | Приоритеты, статусы, дедлайны, исполнители и фильтрация |
| **Комментарии** | Обсуждение задач в контексте проекта |
| **Права доступа** | Пользовательские роли и серверная проверка владения ресурсами |
| **Уведомления** | Настраиваемый email, тестовое письмо, браузерные уведомления |
| **Интерфейс** | Космическая фиолетовая тема, адаптация под узкие экраны |
| **Установка** | PWA-манифест и Service Worker для поддерживаемых браузеров |
| **Инфраструктура** | Docker, PostgreSQL, миграции, CI, health endpoints |

## 🌌 Дизайн

Интерфейс построен вокруг глубокого ночного фона, градиентов фиолетовой туманности, полупрозрачных панелей, выразительной типографики и акцентных состояний. На телефоне основные панели и настройки адаптируются под узкий экран.

## 🧱 Архитектура

```text
task_manager-rework/
├── cmd/api/                 # Точка входа Go API
├── internal/
│   ├── config/              # Конфигурация окружения
│   ├── httpapi/             # HTTP API, авторизация и права доступа
│   └── notifications/       # Отправка email через SMTP
├── db/migrations/           # Версионируемые SQL-миграции
├── frontend/                # React + TypeScript + Vite
│   ├── public/              # PWA manifest, service worker, icon
│   └── src/                 # UI и клиент API
├── .github/workflows/       # CI
├── Dockerfile               # Контейнер Go API
├── compose.yaml             # Локальная среда
├── render.yaml              # Blueprint для Render
├── vercel.json              # Настройки сборки Vercel
└── DEPLOYMENT.md            # Production deployment checklist
```

### Технологии

- **Go** — HTTP API на Gin.
- **PostgreSQL** — постоянное хранение данных; для production рекомендуется Neon.
- **pgx** — драйвер и пул PostgreSQL-соединений.
- **JWT + bcrypt** — аутентификация и хеширование паролей.
- **React + TypeScript + Vite** — клиентское приложение.
- **Lucide** — иконки.
- **Docker Compose** — локальная разработка.
- **Render** — хостинг API.
- **Vercel** — хостинг frontend.
- **Neon** — управляемый PostgreSQL.

---

## 🚀 Локальный запуск

### Требования

- Docker и Docker Compose.
- Git — если клонируешь репозиторий.
- Для запуска без контейнеров: Go 1.23+ и Node.js LTS.

### Запуск одной командой

1. Скопируй `.env.example` в `.env`.
2. Замени `JWT_SECRET` на случайную строку длиной не менее 32 символов.
3. При необходимости измени пароль PostgreSQL и SMTP-параметры.
4. Запусти из корня проекта:

```bash
docker compose up --build
```

После успешного старта:

| Сервис | Адрес |
|---|---|
| Frontend | `http://localhost:5173` |
| API | `http://localhost:8080` |
| Liveness check | `http://localhost:8080/health/live` |
| Readiness check | `http://localhost:8080/health/ready` |

Остановить контейнеры:

```bash
docker compose down
```

Удалить локальную базу данных вместе с volume:

```bash
docker compose down -v
```

> Последняя команда удаляет локальные данные PostgreSQL без возможности восстановить их из этого volume. Не используй её, если хочешь сохранить данные.

---

## 🔔 Настройка уведомлений

Открой приложение, войди в аккаунт и перейди в настройки уведомлений.

### Email

Можно задать адрес, отличный от адреса регистрации, включить/отключить email и выбрать категории событий. Поддерживаются письма о создании задач, изменении статуса задачи и новых комментариях в предусмотренных сценариях.

Добавь SMTP-переменные в окружение **API-сервера**:

```env
SMTP_HOST=
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=
```

Значения зависят от провайдера почты. Убедись, что SMTP-доступ разрешён для выбранного почтового аккаунта. После настройки отправь тестовое письмо в настройках приложения.

**SMTP-пароль должен оставаться только в серверных секретах.** Не добавляй его в `frontend`, Git, скриншоты или публичные issue.

### Уведомления устройства

Браузерные уведомления запрашивают разрешение на конкретном устройстве и в конкретном браузере. Чтобы получать уведомления на телефоне и ПК, открой приложение на каждом устройстве и отдельно разреши уведомления.

Email отправляется сервером и не зависит от того, открыт ли сайт, при условии что API работает, SMTP настроен и событие подпадает под настройки пользователя.

### Что важно знать о push

PWA-манифест и Service Worker подготовлены. Однако **полноценные удалённые push-уведомления при закрытом приложении ещё не включены**: для них нужно добавить Web Push/VAPID, регистрацию и хранение push-подписок, отправку push-событий и обработку отписок. Не путай разрешение браузера с гарантированной доставкой push в фоне.

---

## ☁️ Деплой: Neon + Render + Vercel

Рекомендуемая схема: **Neon → PostgreSQL**, **Render → Go API**, **Vercel → React frontend**. Пошаговая инструкция находится в [`DEPLOYMENT.md`](DEPLOYMENT.md).

### 1. Neon — база данных

Создай проект PostgreSQL в Neon и скопируй connection string. Для production используй строку с TLS и, где подходит, pooled endpoint. Секретную строку базы не помещай во frontend и не публикуй в репозитории.

### 2. Render — API

Импортируй GitHub-репозиторий в Render. Blueprint описан в `render.yaml`. В переменных окружения сервиса укажи:

- `DATABASE_URL` — connection string Neon.
- `JWT_SECRET` — сильный случайный секрет.
- `CORS_ALLOWED_ORIGINS` — точный origin frontend на Vercel.
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM` — если нужны email-уведомления.

API-образ содержит SQL-миграции и migration runner: при старте он применяет ожидающие миграции к базе из `DATABASE_URL`, затем запускает сервер. Для нестандартных сценариев можно задать `AUTO_MIGRATE=false` и применять миграции отдельно.

### 3. Vercel — frontend

Импортируй тот же GitHub-репозиторий в Vercel и выставь **Root Directory** в `frontend`. Docker Compose использует production-сборку на Nginx; Vercel собирает те же статические файлы. Добавь переменную:

```env
VITE_API_URL=https://YOUR-RENDER-API-URL
```

После первого деплоя укажи фактический домен Vercel в `CORS_ALLOWED_ORIGINS` на Render и запусти повторный deploy API.

### 4. Проверка

- `https://YOUR-RENDER-API/health/live`
- `https://YOUR-RENDER-API/health/ready`
- Регистрация тестового пользователя.
- Создание проекта и задачи.
- Изменение статуса задачи.
- Тестовое письмо после настройки SMTP.

> Облачные аккаунты, репозиторий, секреты и домены должны быть доступны и настроены. Нельзя безопасно подставить чужие или выдуманные credentials. Смотри также [`DEPLOYMENT.md`](DEPLOYMENT.md).

---

## 🛡️ Безопасность

В проекте предусмотрены:

- Пароли хешируются через bcrypt; пароли не возвращаются API.
- JWT с ограниченным временем жизни и проверкой алгоритма подписи.
- Проверка авторизации и прав на стороне сервера.
- Параметризованные SQL-запросы.
- Проверка входных данных на уровне HTTP-обработчиков.
- Ограничение размера HTTP-запроса.
- Базовый rate limiting по IP на экземпляр API.
- Заголовки защиты браузера.
- CORS с явным списком разрешённых origin.
- Секреты и SMTP-учётные данные задаются переменными окружения.
- Контейнер API запускается не от root.
- Health endpoints для мониторинга.
- Git ignore для `.env`, сборок и локальных артефактов.

### Границы защиты

Это не является независимым аудитом безопасности и не гарантирует абсолютную защищённость. In-memory rate limiter работает только в рамках одного экземпляра и не заменяет распределённый лимитер. Перед публичным production-запуском нужно выполнить актуальное сканирование зависимостей, интеграционные тесты, проверку восстановления резервной копии, аудит разрешений и проверку логов на утечки. Используй HTTPS и закрывай доступ к секретам на уровне Render/Neon/Vercel.

---

## 🧪 Проверки

Backend:

```bash
gofmt -w ./cmd ./internal
go test ./...
go vet ./...
```

Frontend:

```bash
cd frontend
npm install
npm run build
```

Docker:

```bash
docker compose config
docker compose up --build
```

Сборка зависит от доступности реестров зависимостей. После изменения кода запускай тесты и сборки в CI до выкладки.

---

## 📚 Документация

- [Инструкция production-деплоя](DEPLOYMENT.md)
- [Пример переменных окружения](.env.example)
- [Миграции PostgreSQL](db/migrations/)
- [CI workflow](.github/workflows/ci.yml)

---

<div align="center">

**TASK MANAGER REWORK**  
*Make your workflow orbit.*

</div>
