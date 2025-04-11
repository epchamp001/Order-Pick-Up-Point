SWAGGER_FILE = ./task/swagger.json
SERVER_OUTPUT = generated_server
MODELS_OUTPUT = generated_models

# Генерация серверного кода
generate-server:
	openapi-generator generate -i $(SWAGGER_FILE) -g go-server -o $(SERVER_OUTPUT)

# Генерация клиентского кода
generate-client:
	openapi-generator generate -i $(SWAGGER_FILE) -g go -o $(CLIENT_OUTPUT)

# Генерация models
generate-models:
	openapi-generator generate -i $(SWAGGER_FILE) -g go -o $(MODELS_OUTPUT) --global-property models

# Очистка сгенерированных директорий
clean:
	rm -rf $(SERVER_OUTPUT) $(CLIENT_OUTPUT) $(MODELS_OUTPUT)

proto-pvz:
	protoc \
      -I api/proto \
      -I api/googleapis \
      -I $(shell go list -m -f '{{.Dir}}' github.com/grpc-ecosystem/grpc-gateway/v2) \
      -I /opt/homebrew/include \
      --go_out=. \
      --go-grpc_out=. \
      --grpc-gateway_out=. \
      --openapiv2_out=./docs\
      api/proto/pvz.proto

http-docs:
	swag init -g cmd/pick-up-point/main.go

# Миграции
MIGRATIONS_DIR := ./migrations
DB_DSN := postgres://champ001:123champ123@localhost:5432/pick-up-point?sslmode=disable

.PHONY: up down status create migrate

up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" up

# Откатить последнюю миграцию
down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" down

status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" status

# Создать новый файл миграции
create:
	goose -dir $(MIGRATIONS_DIR) create $(NAME) sql

# Откатить все миграции
reset:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" reset


k6-run:
	K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9000/api/v1/write \
    k6 run \
      --out experimental-prometheus-rw \
      scripts/k6/load_tests.js



