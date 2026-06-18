# Первый деплой на пустой сервер

Инструкция рассчитана на чистый Ubuntu 22.04/24.04 сервер с root-доступом.

## 1. Подготовить DNS

Создайте DNS-запись для домена:

```text
A     game.example.com     <SERVER_IPV4>
AAAA  game.example.com     <SERVER_IPV6, если есть>
```

Дождитесь, чтобы домен начал резолвиться на сервер:

```bash
dig +short game.example.com
```

## 2. Подключиться к серверу

```bash
ssh root@<SERVER_IP>
```

## 3. Установить Docker и открыть порты

Склонируйте репозиторий или временно скопируйте только `scripts/bootstrap-ubuntu.sh`, затем выполните:

```bash
sudo bash scripts/bootstrap-ubuntu.sh
```

Скрипт установит Docker Engine, Docker Compose plugin, Git, включит Docker и откроет порты `22`, `80`, `443` через `ufw`.

## 4. Склонировать приложение

```bash
mkdir -p /opt
cd /opt
git clone https://github.com/mirmikov/Ggame.git
cd Ggame
git switch main
```

Для приватного репозитория используйте SSH-ключ деплой-пользователя или GitHub token.

## 5. Запустить первый деплой

```bash
bash scripts/deploy.sh game.example.com
```

Скрипт создаст `.env`, подтянет актуальный `main`, соберёт Docker image и поднимет compose stack.

## 6. Проверить

На сервере:

```bash
docker compose ps
curl -fsS http://127.0.0.1/api/health
```

Снаружи:

```bash
curl -I https://game.example.com
```

Если DNS уже указывает на сервер и порты `80/443` открыты, Caddy автоматически выпустит HTTPS-сертификат. Приложение будет доступно по `https://game.example.com` без порта.

## Повторный деплой

```bash
cd /opt/Ggame
bash scripts/deploy.sh
```

## Логи

```bash
docker compose logs -f caddy
docker compose logs -f prometheus-battle
```
