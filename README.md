# Semantic Notes Bot

Структура приведена к `PROJECT3.md`:

```text
.
├── bot/
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
├── ai/
│   ├── server.py
│   ├── requirements.txt
│   └── Dockerfile
├── .env
├── .env.example
└── docker-compose.yml
```

## Что внутри

- `bot/` - Telegram бот на Go
- `ai/` - embedding service на Python (`sentence-transformers`, `all-MiniLM-L6-v2`, FastAPI)
- Заметки хранятся локально в `notes.json` (путь через `NOTES_FILE`)
- Семантический поиск выполняется через `AI_URL`
- Системные команды: `/info`, `/status`, `/reboot`

## Кнопочный сценарий

- `Добавить заметку` -> отправляешь текст следующим сообщением
- `Поиск по смыслу` -> отправляешь запрос следующим сообщением
- `Показать заметки` -> список последних заметок
- `Удалить заметку` -> бот показывает последние заметки, отправляешь `ID` для удаления
- `/info` -> внешний IP, локальный IP, SSID
- `/status` -> uptime, CPU, RAM
- `/reboot` -> перезагрузка Raspberry Pi

## Локальный запуск

1. AI сервис:

```bash
cd ai
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
uvicorn server:app --host 0.0.0.0 --port 8000
```

2. Бот (в другом терминале):

```bash
cp .env.example .env
# заполни BOT_TOKEN и ALLOWED_USER_ID, для локального запуска AI_URL=http://localhost:8000
# при необходимости задай REBOOT_COMMAND (по умолчанию sudo reboot)
cd bot
go mod tidy
go run .
```

## Docker Compose

```bash
cp .env.example .env
docker compose up -d --build
docker compose logs -f
```

По умолчанию в docker `AI_URL=http://ai:8000`, а заметки лежат в volume `notes_data`.
Порт AI сервиса наружу не публикуется: доступ к нему есть только у `bot` через внутреннюю docker-сеть.
Для `/status` контейнер бота читает host `/proc` и `/sys` через mount `/proc:/host_proc:ro` и `/sys:/host_sys:ro` (`HOST_PROC=/host_proc`, `HOST_SYS=/host_sys`), поэтому показывает статус всей машины и температуру CPU.
