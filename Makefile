include make/lint.mk
include make/build.mk

lint: cart-lint loms-lint notifier-lint comments-lint

build: cart-build loms-build notifier-build comments-build

run-cart:
	cd cart && make run .

run-loms:
	cd loms && make run .

up:
	echo "starting docker build"
	docker-compose up --build

down:
	docker-compose down

run-all: build up

pprof-cart:
	cd cart && make pprof

pprof-loms:
	cd loms && make pprof