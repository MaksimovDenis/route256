SERVICE_NAME := cart
PORT := 8080
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

bindir:
	echo "build cart"
	mkdir -p ${BINDIR}

.PHONY: .bin-deps-proto
.bin-deps-proto:
	$(info Installing binary dependencies...)

	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1 && \
    GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

.vendor-proto: .vendor-rm \
  vendor-proto/google/protobuf \
  vendor-proto/validate \
  vendor-proto/google/api \
  vendor-proto/protoc-gen-openapiv2/options

PROTO_DIRS := \
  ../loms/api/loms/v1

PHONY: .protoc-generate
.protoc-generate: .bin-deps-proto
	@dirs=($(PROTO_DIRS)); \
	for i in "$${!dirs[@]}"; do \
		proto_dir=$${dirs[$$i]}; \
		out_dir=internal/pb; \
		echo "Generating from $$proto_dir into $$out_dir"; \
		mkdir -p $$out_dir; \
		protoc \
			-I ../loms/api \
			-I ./vendor-proto \
			--plugin=protoc-gen-go=$(LOCAL_BIN)/protoc-gen-go \
			--go_out=$$out_dir \
			--go_opt=paths=source_relative \
			--plugin=protoc-gen-go-grpc=$(LOCAL_BIN)/protoc-gen-go-grpc \
			--go-grpc_out=$$out_dir \
			--go-grpc_opt=paths=source_relative \
			$$proto_dir/*.proto || exit 1; \
	done
	
build: bindir
	go build -o ${BINDIR}/cart ./cmd/cart

.PHONY: generate
generate: # Генерирует pb файлы
	@make .vendor-proto
	@make .protoc-generate
	go mod tidy

run: build
	CONFIG_FILE=./configs/values_local.yaml ${BINDIR}/cart

pprof-heap:
	go tool pprof -http=":8000" "http://localhost:$(PORT)/debug/pprof/heap?seconds=5"

pprof-cpu:
	go tool pprof -http=":8001" "http://localhost:$(PORT)/debug/pprof/profile?seconds=5"

pprof:  ## Запускает heap и CPU профилирование параллельно
	@echo "Запуск heap и CPU профилирования..."
	$(MAKE) pprof-heap &
	$(MAKE) pprof-cpu &
	wait