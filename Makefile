WORKERS ?= 10
FOLLOWERS ?= 2

start:
	sudo docker compose up --scale worker=${WORKERS} --scale follower=${FOLLOWERS}

stop:
	sudo docker compose down

restart: stop start
dev: build start

build:
	sudo docker compose build

reset-logs:
	rm -rf logs/*
