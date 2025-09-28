SERVICE_NAME := comments
ABS_PATH_TO_SERVICE_MAKE :=

ifeq (,$(findstring /${SERVICE_NAME},${CURDIR}))
 ABS_PATH_TO_SERVICE_MAKE := ${CURDIR}/make
else
 ABS_PATH_TO_SERVICE_MAKE := $(subst /${SERVICE_NAME},/make,${CURDIR})
endif

include ${ABS_PATH_TO_SERVICE_MAKE}/service_make.mk

BINDIR=${CURDIR}/bin
MINIMOCK_TAG=v3.4.5
LOCAL_BIN:=$(CURDIR)/bin
SHELL := /bin/bash

MIGRATIONS_FOLDER := ./migration

install-goose:
	$(info Installing goose binary into [$(LOCAL_BIN)]...)
	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

.PHONY: install-sqlc
install-sqlc:
	$(info Installing sqlc binary into [$(LOCAL_BIN)]...)
	GOBIN=$(LOCAL_BIN) go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0

.PHONY: generate-sqlc
generate-sqlc:
	$(LOCAL_BIN)/sqlc generate

run-migrations: ##Накатывает миграции
	echo "run migrations"
	goose -dir ./loms/migration/ postgres "$(DB_DSN)" up

.PHONY: .bin-deps-proto
.bin-deps-proto:
	$(info Installing binary dependencies...)

	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1 && \
    GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0 && \
	GOBIN=$(LOCAL_BIN) go install github.com/envoyproxy/protoc-gen-validate@v1.0.4 && \
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.19.1 && \
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.19.1 && \
	GOBIN=$(LOCAL_BIN) go install github.com/go-swagger/go-swagger/cmd/swagger@v0.30.5
	GOBIN=$(LOCAL_BIN) go install github.com/rakyll/statik@v0.1.7

.vendor-proto: .vendor-rm \
  vendor-proto/google/protobuf \
  vendor-proto/validate \
  vendor-proto/google/api \
  vendor-proto/protoc-gen-openapiv2/options

PROTO_DIRS := \
  ./api/loms/v1

.PHONY: .protoc-generate
.protoc-generate: .bin-deps-proto generate-swagger
	@dirs=($(PROTO_DIRS)); \
	for i in "$${!dirs[@]}"; do \
		proto_dir=$${dirs[$$i]}; \
		out_dir=internal/pb; \
		echo "Generating from $$proto_dir into $$out_dir"; \
		mkdir -p $$out_dir $$out_dir/openapiv2; \
		protoc \
			-I ./api \
			-I ./vendor-proto \
			--plugin=protoc-gen-go=$(LOCAL_BIN)/protoc-gen-go \
			--go_out=$$out_dir \
			--go_opt=paths=source_relative \
			--plugin=protoc-gen-go-grpc=$(LOCAL_BIN)/protoc-gen-go-grpc \
			--go-grpc_out=$$out_dir \
			--go-grpc_opt=paths=source_relative \
			--plugin=protoc-gen-validate=$(LOCAL_BIN)/protoc-gen-validate \
			--validate_out="lang=go,paths=source_relative:$$out_dir" \
			--plugin=protoc-gen-grpc-gateway=$(LOCAL_BIN)/protoc-gen-grpc-gateway \
			--grpc-gateway_out=$$out_dir \
			--grpc-gateway_opt=logtostderr=true \
			--grpc-gateway_opt=paths=source_relative \
			--grpc-gateway_opt=generate_unbound_methods=true \
			--plugin=protoc-gen-openapiv2=$(LOCAL_BIN)/protoc-gen-openapiv2 \
			--openapiv2_out=$$out_dir/openapiv2 \
			--openapiv2_opt=logtostderr=true \
			$$proto_dir/*.proto || exit 1; \
	done

.PHONY: .serve-swagger
.serve-swagger: .protoc-generate
	bin/swagger serve internal/pb/openapiv2/loms/v1/api.swagger.json

generate-swagger:
	mkdir -p internal/pb/openapiv2/loms/v1/
	$(LOCAL_BIN)/statik -src=internal/pb/openapiv2/loms/v1/ -include='*.css,*.html,*.js,*.json,*.png'

.PHONY: generate
generate: ## Генерирует pb файлы
	@make .vendor-proto
	@make .protoc-generate
	go mod tidy

bindir:
	mkdir -p ${BINDIR}

build: bindir
	echo "build comments"


run-migrations:
	echo "run migrations"
