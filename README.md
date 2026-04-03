# SerMon Backend

Серверная часть системы мониторинга серверов и инцидентов.

## Требования

- Go 1.20+
- PostgreSQL 13+

## Установка

Склонировать репозиторий:

```bash
git clone https://github.com/Pat0sya/Sermon-backend
cd sermon-backend
```
Собрать приложение:
```bash
go build -o sermon-backend ./cmd/api
```
# Настройка базы данных
## 1. Создание базы данных

Подключитесь к PostgreSQL и выполните:
```sql
CREATE DATABASE sermon;

CREATE USER sermon_user WITH PASSWORD 'strong_password';

GRANT ALL PRIVILEGES ON DATABASE sermon TO sermon_user;
```
Далее:
```sql
\c sermon

GRANT ALL ON SCHEMA public TO sermon_user;
```
## 2. Настройка доступа

В postgresql.conf:
```sql
listen_addresses = '*'
port = 5432
```
В pg_hba.conf:
```sql
host    sermon    sermon_user    192.168.1.0/24    md5
```
Перезапустить PostgreSQL.
## Конфигурация

Создать файл .env:
```json
APP_PORT=8080
APP_ENV=prod

DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=sermon_user
DB_PASSWORD=strong_password
DB_NAME=sermon
DB_SSLMODE=disable

JWT_SECRET=your_secret_key
```
## Запуск
```bash
./sermon-backend
```
