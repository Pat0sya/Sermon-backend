# SerMon Backend

## Описание

Серверная часть системы мониторинга серверов и управления инцидентами.

Реализует:

* авторизацию пользователей (JWT)
* управление серверами
* приём метрик от агентов
* генерацию инцидентов
* работу с комментариями
* управление пользователями

---

## Требования

* Go 1.21+
* PostgreSQL 13+
* Открытый порт 8080

---

## Настройка базы данных

### 1. Установка PostgreSQL

Установите PostgreSQL на отдельный сервер или локально.

---

### 2. Создание базы данных

Подключитесь к PostgreSQL и выполните:

```sql
CREATE DATABASE sermon;

CREATE USER sermon_user WITH PASSWORD 'strong_password';

GRANT ALL PRIVILEGES ON DATABASE sermon TO sermon_user;
```

Затем:

```sql
\c sermon

GRANT ALL ON SCHEMA public TO sermon_user;
```

---

### 3. Настройка доступа

В файле `postgresql.conf`:

```conf
listen_addresses = '*'
port = 5432
```

В файле `pg_hba.conf`:

```conf
host    sermon    sermon_user    192.168.1.0/24    md5
```

Перезапустите PostgreSQL.

---

## Конфигурация backend

Создайте файл `.env` рядом с бинарником:

```env
APP_PORT=8080
APP_ENV=prod

DB_HOST=192.168.1.80
DB_PORT=5432
DB_USER=sermon_user
DB_PASSWORD=strong_password
DB_NAME=sermon
DB_SSLMODE=disable

JWT_SECRET=super_secret_key
```

---

## Сборка

```bash
go build -o sermon-backend ./cmd/api
```

---

## Запуск

Linux:

```bash
./sermon-backend
```

Windows:

```powershell
.\sermon-backend.exe
```

---

## Проверка

```powershell
Invoke-WebRequest http://localhost:8080/api/v1/auth/login -Method POST -ContentType "application/json" -Body '{"username":"admin","password":"admin"}'
```

---

## API

Основные маршруты:

* POST /api/v1/auth/login
* GET /api/v1/servers
* POST /api/v1/servers
* GET /api/v1/incidents
* PATCH /api/v1/incidents/{id}/status
* POST /api/v1/agent/metrics

---

## Примечания

* Backend должен быть доступен по сети (не только localhost)
* Порт 8080 должен быть открыт в firewall
* В production рекомендуется использовать HTTPS через reverse proxy
