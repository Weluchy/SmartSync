# SmartSync — Инструкция по бесплатному деплою (бюджет 0$)

## Архитектура деплоя

| Компонент | Сервис | Цена | Регистрация |
|---|---|---|---|
| **Фронтенд** (React) | **Cloudflare Pages** | Бесплатно (безлимит) | Беларусь: да |
| **PostgreSQL** | **Supabase Free** | 500 МБ БД | Беларусь: да |
| **MongoDB** | **MongoDB Atlas Free (M0)** | 512 МБ БД | Беларусь: да |
| **Redis** | **Upstash Free** | 10 МБ | Беларусь: да |
| **Go-микросервисы** | **AWS EC2 t2.micro** | **Free Tier (12 мес)** | Беларусь: да |
| **NATS** | Там же, на EC2 | — | — |
| **Prometheus + Grafana** | Там же, на EC2 | — | — |
| **CDN + Tunnel** | **Cloudflare** | Бесплатно | Беларусь: да |

**Почему AWS EC2 (не Fly.io, не Render)**: твой проект содержит NATS + Prometheus + Grafana + 6 Go-микросервисов. Только полноценный VPS с Docker Compose может всё это поднять. AWS t2.micro (1 GB RAM) даётся бесплатно на 12 месяцев и доступен в Беларуси.

---

## 1. AWS EC2 — бесплатный сервер

### 1.1 Регистрация
1. Перейди на [aws.amazon.com/free](https://aws.amazon.com/free/)
2. Нажми **"Create a Free Account"**
3. Заполни Email, пароль, имя аккаунта
4. Введи данные — **кредитная карта** потребуется (холд ~1$ на сутки, потом возвращается)
5. Подтверди номер телефона
6. Выбери **Basic support (Free)**

> Для Беларуси: AWS доступен и работает отлично.

### 1.2 Создание EC2 инстанса (t2.micro)
1. Войди в [AWS Console](https://console.aws.amazon.com/)
2. В поиске введи **EC2** → перейди
3. Нажми **"Launch instance"**
4. **Name:** `smartsync-vm`
5. **Application and OS Images:** выбери **Ubuntu Server 24.04 LTS** (Free Tier eligible)
6. **Architecture:** 64-bit (x86)
7. **Instance type:** t2.micro (1 vCPU, 1 GiB RAM) — помечено "Free Tier eligible"
8. **Key pair (login):**
   - Нажми **"Create new key pair"**
   - Название: `smartsync-key`
   - Type: RSA
   - Format: .ppk для PuTTY **ИЛИ** .pem для OpenSSH (Ubuntu/macOS)
   - **Скачай файл!** Без него не зайдёшь.
9. **Network settings:**
   - Нажми **"Edit"**
   - ✅ Allow SSH traffic from → **Anywhere** (0.0.0.0/0)
   - ✅ Allow HTTP traffic from the internet
   - ✅ Allow HTTPS traffic from the internet
10. **Storage:** 30 GiB gp2 (входит в Free Tier)
11. Нажми **"Launch instance"**

### 1.3 Настройка Security Group (открыть порты)
1. В EC2 → слева **Security Groups**
2. Найди группу для твоего инстанса (обычно `sg-... / launch-wizard-1`)
3. Правый клик → **Edit inbound rules**
4. Добавь правила:

| Type | Protocol | Port Range | Source | Description |
|---|---|---|---|---|
| SSH | TCP | 22 | 0.0.0.0/0 | SSH |
| HTTP | TCP | 80 | 0.0.0.0/0 | HTTP |
| HTTPS | TCP | 443 | 0.0.0.0/0 | HTTPS |
| Custom TCP | TCP | 8000 | 0.0.0.0/0 | Gateway API |
| Custom TCP | TCP | 3000 | 0.0.0.0/0 | Grafana (опционально) |
| Custom TCP | TCP | 9090 | 0.0.0.0/0 | Prometheus (опционально) |
| Custom TCP | TCP | 4222 | 0.0.0.0/0 | NATS |

### 1.4 Подключение к серверу

**На Windows (PuTTY):**
1. Скачай [PuTTY](https://www.putty.org/)
2. Скачай [PuTTYgen](https://www.puttygen.com/)
3. Открой PuTTYgen → Load → выбери скачанный .ppk файл
4. Введи имя пользователя: `ubuntu`
5. Запомни публичный IP инстанса (EC2 → Instances → Public IPv4 DNS)
6. Открой PuTTY:
   - Host Name: `ubuntu@<PUBLIC_IP>`
   - Connection → SSH → Auth → Private key: выбери .ppk
   - Open

**На macOS/Linux:**
```bash
chmod 400 ~/Downloads/smartsync-key.pem
ssh -i ~/Downloads/smartsync-key.pem ubuntu@<PUBLIC_IP>
```

---

## 2. Установка Docker на сервер

```bash
# Обновить пакеты
sudo apt-get update && sudo apt-get upgrade -y

# Установить Docker (официальный скрипт)
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Добавить пользователя в группу docker
sudo usermod -aG docker $USER

# Установить Docker Compose
sudo apt-get install -y docker-compose-plugin

# Выйти и зайти снова (или выполнить newgrp docker)
exit
# Зайди снова
ssh -i ~/Downloads/smartsync-key.pem ubuntu@<PUBLIC_IP>
```

---

## 3. Регистрация managed БД

### 3.1 PostgreSQL → Supabase Free

1. [supabase.com](https://supabase.com) → Sign up (можно через GitHub)
2. **New project**
   - Name: `smartsync`
   - Database password: придумать
   - Region: **West EU (Ireland)** — ближайший свободный
3. Жди ~1-2 минуты
4. В меню слева → **Project Settings → Database**
5. Скопируй **Connection string**
   - Она выглядит так: `postgresql://postgres:your_password@db.xxxxx.supabase.co:5432/postgres`
   - Допиши в конце `?sslmode=require`

### 3.2 MongoDB → MongoDB Atlas Free

1. [mongodb.com/atlas](https://www.mongodb.com/atlas) → Try Free
2. Создай **M0 Free Cluster**
   - Provider: AWS
   - Region: **Frankfurt (eu-central-1)**
3. В Security → Network Access → **Add IP Address** → `0.0.0.0/0` (Allow All)
4. Database Access → **Add New Database User**
   - Username + Password (запомни)
5. Cluster → **Connect → Drivers**
   - Скопируй URI: `mongodb+srv://<user>:<password>@cluster0.xxxxx.mongodb.net/`

### 3.3 Redis → Upstash Free

1. [upstash.com](https://upstash.com) → Sign up (GitHub)
2. **Create Database**
   - Name: `smartsync-redis`
   - Region: **eu-west-1** (Ireland)
3. Free tier: 10 MB — хватит
4. В **REST API** → скопируй **UPSTASH_REDIS_REST_URL**
   - Удали `https://` и порт — останется `xxxxx.upstash.io:6379`

---

## 4. Деплой проекта на сервер

### 4.1 Клонирование репозитория

```bash
cd ~
git clone https://github.com/Weluchy/SmartSync.git
cd SmartSync
```

### 4.2 Создание .env.cloud со своими данными

```bash
cp .env.cloud .env.cloud.prod
nano .env.cloud.prod
```

Заполни:

```ini
JWT_SECRET=<придумай_длинную_строку_32_символа>

# PostgreSQL (Supabase)
DATABASE_URL=postgresql://postgres:password@db.xxxxx.supabase.co:5432/postgres?sslmode=require

# MongoDB (Atlas)
MONGO_URI=mongodb+srv://user:password@cluster0.xxxxx.mongodb.net/

# Redis (Upstash)
REDIS_ADDR=xxxxx.upstash.io:6379

# Grafana
GF_SECURITY_ADMIN_PASSWORD=<придумай_пароль>

# Архитектура
TARGETARCH=amd64
```

Затем:

```bash
mv .env.cloud.prod .env.cloud
```

### 4.3 Подготовка БД (PostgreSQL)

Нужно инициализировать таблицы. Есть два способа:

**Способ A: psql (через Docker)**
```bash
# Запустить psql
docker run --rm -it postgres:15-alpine psql "$DATABASE_URL"
# Выполнить содержимое db/init.sql
```

**Способ B: Supabase SQL Editor**
1. В консоли Supabase → **SQL Editor**
2. Открой `db/init.sql` локально и скопируй весь SQL
3. Вставь в SQL Editor → Run

### 4.4 Запуск

```bash
# Сделать скрипт исполняемым
chmod +x start-cloud.sh

# Запустить
./start-cloud.sh
```

Скрипт сам определит архитектуру (amd64), установит Docker если нужно, соберёт образы и запустит все контейнеры.

**Проверка:**
```bash
docker compose -f docker-compose.cloud.yml ps
```

Ты должен увидеть 9 контейнеров (nats, prometheus, grafana, gateway, auth-service, task-service, priority-service, deadline-service, audit-service).

---

## 5. Фронтенд — Cloudflare Pages

### 5.1 Регистрация Cloudflare

1. [dash.cloudflare.com/sign-up](https://dash.cloudflare.com/sign-up)
2. Добавь свой домен (если есть)
3. Если домена нет — пропусти этот шаг, фронтенд будет доступен по URL `xxx.pages.dev`

### 5.2 Деплой на Cloudflare Pages

**Через Git (рекомендую):**
1. В Cloudflare Dashboard → **Workers & Pages → Create → Pages → Connect to Git**
2. Выбери репозиторий SmartSync
3. Build settings:
   - Framework preset: **Vite**
   - Build command: `cd frontend && npm install && npm run build`
   - Build output directory: `frontend/dist`
4. Environment variables (если нужно):
   - `VITE_API_URL` = `https://ваш-сервер:8000` (или `https://api.ваш-домен.com`)
5. **Save and Deploy**

**Через Wrangler CLI:**
```bash
npm install -g wrangler
cd frontend
npm install
wrangler pages deploy dist --project-name smartsync
```

---

## 6. Домен и HTTPS (Cloudflare Tunnel)

### 6.1 Если есть свой домен

1. В Cloudflare → DNS → **Add record**:
   - Type: A
   - Name: `smart` (или `@` для корневого)
   - IPv4: `<PUBLIC_IP сервера>`
   - Proxy status: **Proxied** (оранжевое облако) — даёт HTTPS + защиту

2. В Cloudflare → **SSL/TLS → Overview**
   - Выбери **Full (strict)**

3. **(Рекомендую) Cloudflare Tunnel — безопаснее, без открытия портов:**
   ```bash
   # На сервере
   sudo apt-get install -y cloudflared
   
   # Авторизация (откроет браузер)
   cloudflared tunnel login
   
   # Создать туннель
   cloudflared tunnel create smartsync
   
   # Настроить маршруты
   cloudflared tunnel route dns smartsync smart.ваш-домен.com
   cloudflared tunnel route dns smartsync api.ваш-домен.com
   ```

   Файл `~/.cloudflared/config.yml`:
   ```yaml
   tunnel: <TUNNEL-ID>
   credentials-file: /home/ubuntu/.cloudflared/<TUNNEL-ID>.json
   ingress:
     - hostname: api.ваш-домен.com
       service: http://localhost:8000
     - hostname: smart.ваш-домен.com
       service: http://localhost:8000
     - service: http_status:404
   ```

   ```bash
   # Установить как сервис
   sudo cloudflared service install
   ```

### 6.2 Если домена нет (бесплатный Cloudflare Pages URL)

Фронтенд будет доступен по адресу `https://smartsync-xxxxx.pages.dev` (автоматически сгенерируется)

Для API нужно настроить в коде фронтенда переменную `VITE_API_URL` на `http://<PUBLIC_IP>:8000`

---

## 7. Проверка работоспособности

### Что должно работать:

| URL | Описание |
|---|---|
| `http://<PUBLIC_IP>:80` (или домен) | Фронтенд SmartSync |
| `http://<PUBLIC_IP>:8000` | Gateway API |
| `http://<PUBLIC_IP>:3000` | Grafana (admin / ваш пароль) |
| `http://<PUBLIC_IP>:9090` | Prometheus |

### Проверка API:

```bash
# Регистрация
curl -X POST http://<PUBLIC_IP>:8000/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@test.com","password":"test123"}'

# Логин
curl -X POST http://<PUBLIC_IP>:8000/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"test123"}'

# Получить токен и использовать:
TOKEN="ваш_токен"

# Создать проект
curl -X POST http://<PUBLIC_IP>:8000/projects \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Дипломный проект"}'
```

---

## 8. Команды для управления

```bash
# Посмотреть логи
docker compose -f docker-compose.cloud.yml logs -f
docker compose -f docker-compose.cloud.yml logs -f gateway

# Перезапустить конкретный сервис
docker compose -f docker-compose.cloud.yml restart task-service

# Остановить всё
docker compose -f docker-compose.cloud.yml down

# Запустить снова (не пересобирая)
docker compose -f docker-compose.cloud.yml up -d

# Пересобрать один сервис
docker compose -f docker-compose.cloud.yml build task-service
docker compose -f docker-compose.cloud.yml up -d task-service

# Зайти в контейнер
docker exec -it smartsync_gateway sh
```

---

## 9. Экономия RAM на t2.micro (1 GB)

Если будет не хватать памяти:

1. **Prometheus:** закомментируй в docker-compose.cloud.yml (освободит ~50 МБ)
2. **Grafana:** используй **Grafana Cloud Free** (10k метрик) вместо своего (освободит ~70 МБ)
3. **Deadline-service:** отключи на время демонстрации (освободит ~20 МБ)
4. **Приоритет:** чем меньше трафика, тем меньше памяти едят Go-сервисы

Но обычно на t2.micro всё влезает: 6 Go-сервисов (~150 МБ) + NATS (~30 МБ) + Prometheus (~50 МБ) + Grafana (~70 МБ) + система (~300 МБ) = ~600 МБ. Остаётся 400 МБ запаса.

---

## 10. Комиссия на дипломе

Чтобы комиссия могла открыть сайт:

1. Если есть домен → раздаёшь URL: `https://smart.ваш-домен.com`
2. Если есть Cloudflare Pages URL → раздаёшь `https://smartsync-xxxxx.pages.dev`
3. Если только IP сервера → раздаёшь `http://<PUBLIC_IP>:8000`

**Важно:** Cloudflare работает в РФ и РБ без ограничений. AWS EC2 тоже доступен. AWS Free Tier не блокирует трафик из этих стран.

### Альтернатива AWS — если не хочешь вводить карту

Если AWS не подходит из-за кредитки:

| Провайдер | Бесплатный сервер | Доступность в РБ |
|---|---|---|
| **Hetzner** | Нет бесплатного, но от 3.99€/мес | ✅ |
| **Fly.io** | 3 VM по 256 МБ (768 МБ всего) | ✅ — регистрация без карты |
| **Koyeb** | 1 VM (1 vCPU, 1 GB RAM) | ✅ |

Но Fly.io не поддерживает docker-compose — придётся настраивать каждый сервис отдельно через `fly.toml`.

---

## На этом всё! Удачи на защите диплома 🎉