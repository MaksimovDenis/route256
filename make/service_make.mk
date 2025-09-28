test: ## Запуск юнит-тестов с гонками, 100 раз
	go test -race -count 100 ./...

.PHONY: help
help: ## Показывает список всех команд
	@echo "Доступные команды:"
	@grep -hE '^[a-zA-Z0-9_.-]+:.*##' $(MAKEFILE_LIST) | column -s: -t

generate-test-cover: ## Генерация покрытия тестами с отчётом в HTML
	go test -coverpkg=./internal/... -coverprofile=cover.out.tmp ./...
	grep -vE 'mock' cover.out.tmp > cover.out
	go tool cover -html=cover.out -o coverage.html && open coverage.html
	@echo "Total coverage:"
	@go tool cover -func=cover.out | grep total:


.PHONY: generate-mocks
generate-mocks: ## Генерация моков с помощью minimock
	$(info Generating mocks...)
	PATH="$(LOCAL_BIN):$$PATH" go generate -x -run=minimock ./...

.PHONY: .bin-deps
.bin-deps: ## Установка бинарных зависимостей
	$(info Installing custom binary dependencies...)
	tmp=$$(mktemp -d) && cd $$tmp && pwd && go mod init temp && \
		GOBIN=$(LOCAL_BIN) go install github.com/gojuno/minimock/v3/cmd/minimock@$(MINIMOCK_TAG) && \
	rm -rf $$tmp 
	
lint: ## Запуск линтера
	golangci-lint run -c ../.golangci.yaml 

prepush: lint test ## Линт + тесты + go mod tidy
	go mod tidy

.PHONY: .vendor-rm ## Удаление вендор файлов
.vendor-rm:
	rm -rf ./vendor-proto

vendor-proto/google/protobuf: ## Устанавливаем proto описания google/protobuf
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
		https://github.com/protocolbuffers/protobuf vendor-proto/protobuf &&\
	cd vendor-proto/protobuf &&\
	git sparse-checkout set --no-cone src/google/protobuf &&\
	git checkout
	mkdir -p vendor-proto/google
	mv vendor-proto/protobuf/src/google/protobuf vendor-proto/google
	rm -rf vendor-proto/protobuf

vendor-proto/validate: ## Устанавливаем proto описания validate
	git clone -b main --single-branch --depth=2 --filter=tree:0 \
		https://github.com/bufbuild/protoc-gen-validate vendor-proto/tmp && \
		cd vendor-proto/tmp && \
		git sparse-checkout set --no-cone validate &&\
		git checkout
		mkdir -p vendor-proto/validate
		mv vendor-proto/tmp/validate vendor-proto/
		rm -rf vendor-proto/tmp

vendor-proto/google/api: ## Устанавливаем proto описания google/googleapis
	git clone -b master --single-branch -n --depth=1 --filter=tree:0 \
 		https://github.com/googleapis/googleapis vendor-proto/googleapis && \
 	cd vendor-proto/googleapis && \
	git sparse-checkout set --no-cone google/api && \
	git checkout
	mkdir -p  vendor-proto/google
	mv vendor-proto/googleapis/google/api vendor-proto/google
	rm -rf vendor-proto/googleapis

vendor-proto/protoc-gen-openapiv2/options: ## Устанавливаем proto описания protoc-gen-openapiv2/options
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
 		https://github.com/grpc-ecosystem/grpc-gateway vendor-proto/grpc-ecosystem && \
 	cd vendor-proto/grpc-ecosystem && \
	git sparse-checkout set --no-cone protoc-gen-openapiv2/options && \
	git checkout
	mkdir -p vendor-proto/protoc-gen-openapiv2
	mv vendor-proto/grpc-ecosystem/protoc-gen-openapiv2/options vendor-proto/protoc-gen-openapiv2
	rm -rf vendor-proto/grpc-ecosystem

db-migration-status:  ## Статус миграций
	$(LOCAL_BIN)/goose -dir ${MIGRATION_DIR} postgres ${PG_DSN} status -v

db-create-migration: ## Создание миграции
	$(LOCAL_BIN)/goose -dir $(MIGRATIONS_FOLDER) create $(n) sql

db-migrate: ## Установка миграций
	$(LOCAL_BIN)/goose -dir $(MIGRATIONS_FOLDER) postgres "$(LOCAL_DB_DSN)" up

db-migrate-down: ## Откат миграции
	$(LOCAL_BIN)/goose -dir $(MIGRATIONS_FOLDER) postgres "$(LOCAL_DB_DSN)" down