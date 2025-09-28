SERVICE_NAME := notifier
ABS_PATH_TO_SERVICE_MAKE :=

ifeq (,$(findstring /${SERVICE_NAME},${CURDIR}))
 ABS_PATH_TO_SERVICE_MAKE := ${CURDIR}/make
else
 ABS_PATH_TO_SERVICE_MAKE := $(subst /${SERVICE_NAME},/make,${CURDIR})
endif

include ${ABS_PATH_TO_SERVICE_MAKE}/service_make.mk

BINDIR=${CURDIR}/bin
PACKAGE=route256/notifier

bindir:
	echo "build cart"
	mkdir -p ${BINDIR}

build: bindir
	go build -o ${BINDIR}/notifier ./cmd/notifier

run: build
	CONFIG_FILE=./configs/values_local.yaml ${BINDIR}/notifier