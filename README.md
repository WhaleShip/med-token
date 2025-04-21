# Refresh Токены

[![linters](https://github.com/WhaleShip/med-token/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/WhaleShip/med-token/actions/workflows/golangci-lint.yml)
[![tests](https://github.com/WhaleShip/med-token/actions/workflows/unit-tests.yaml/badge.svg)](https://github.com/WhaleShip/med-token/actions/workflows/unit-tests.yaml)

## Запуск

### 1. Создать .env

переименовать [examplse.env](example.env) в .env


### 2. Запустить через докер
```sh
docker compose up
```


приложение будет доступно на http://localhost:8080
GET /token?user_id=<UUID> -> создать пару токенов на uuid

POST /refresh (тело {"refresh_token":"refresh token"} -> обновить токены

### контакты
[![Telegram Icon](https://raw.githubusercontent.com/CLorant/readme-social-icons/main/large/light/telegram.svg)](https://t.me/PanHater)
[![medium-light-discord](https://raw.githubusercontent.com/CLorant/readme-social-icons/main/large/light/discord.svg)](https://discord.com/users/1249015796852719617)
