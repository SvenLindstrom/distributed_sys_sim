WORKERS ?= 3


start:
	sudo docker compose up --scale worker=${WORKERS}

stop:
	sudo docker compose down

restart: stop start

build:
	sudo docker compose build

reset-logs:
	rm -rf logs/*
