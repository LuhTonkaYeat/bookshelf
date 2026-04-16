# BookShelf

Учебный микросервис: HTTP → gRPC → PostgreSQL.

## Запуск

```bash
docker compose up
```

## API

```bash
# Сохранить
curl -X POST http://localhost:8080/save -H "Content-Type: application/json" -d '{"quote": "...", "author": "Tolkien"}'

# Все цитаты
curl http://localhost:8080/quotes

# Удалить
curl -X DELETE http://localhost:8080/quotes/1
```