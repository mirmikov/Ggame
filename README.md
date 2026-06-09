# Prometheus Battle

Браузерная multiplayer-игра для учеников 9–11 классов: две команды отвечают на вопросы по информатике, усиливают автоматических боевых юнитов и штурмуют башню соперника.

## Что работает

- создание комнаты с уникальным ID и подключение по нему;
- выбор команды NexGen или OmniSoft и запуск только host-игроком;
- тестовый BOT-7 для локального матча без второго браузера;
- REST API, realtime-обновления через WebSocket и in-memory состояние;
- 40 русскоязычных вопросов с объяснениями и Terminal с пятью задачами;
- очки, XP, уровни, усиление характеристик и блокировка после трёх ошибок;
- секундная симуляция авто-боя, атака юнитов и башен, респавн и результаты;
- синхронные снаряды, трассеры, взрывы и отображение нанесённого урона;
- top-down арена с тремя отдельными маршрутами для игроков;
- cyberpunk React-интерфейс для входа, лобби, арены и итоговой таблицы.

## Структура

```text
backend/
  cmd/server/main.go
  internal/{game,models,questions,rooms,ws}
frontend/
  src/{api,styles,types}
  src/App.tsx
ASSETS.md
TODO.md
```

## Запуск

Требования: Go 1.22+, Node.js 20+.

### Мгновенный запуск через Docker

```bash
docker compose up --build
```

Откройте `http://localhost:8080`. Остановить проект:

```bash
docker compose down
```

Другой внешний порт можно задать так: `PORT=3000 docker compose up --build`.

Dockerfile не требует Node-образа: он использует готовую production-сборку из `frontend/dist` и vendored Go-зависимости. После изменений frontend обновите её командой `cd frontend && npm run build`.

### Локальная разработка

Backend:

```bash
cd backend
go mod download
go run ./cmd/server
```

Frontend в другом терминале:

```bash
cd frontend
npm install
npm run dev
```

Откройте `http://localhost:5173`. API по умолчанию работает на `http://localhost:8080`. Другой адрес можно задать через `VITE_API_URL`.

Для проверки multiplayer откройте приложение в двух вкладках: создайте комнату в первой, войдите по ID во второй, выберите разные команды и запустите бой.

Для одиночной проверки выберите команду в лобби, нажмите `+ ДОБАВИТЬ TEST BOT`, затем запустите бой. BOT-7 автоматически отвечает на вопросы и усиливает свой юнит.

## API

- `POST /api/rooms`
- `GET /api/rooms/{id}`
- `POST /api/rooms/{id}/join`
- `POST /api/rooms/{id}/team`
- `POST /api/rooms/{id}/start`
- `POST /api/rooms/{id}/answer`
- `GET /api/questions?grade=9`
- `GET /api/tasks`
- `POST /api/rooms/{id}/task`
- `WS /ws/rooms/{id}`

## Ограничения MVP

Состояние пропадает при перезапуске backend. Нет аккаунтов, базы данных, полноценного запуска пользовательского кода и защиты от нечестного клиента. Вопросы выдаются случайно и находятся в Go seed-файле.
