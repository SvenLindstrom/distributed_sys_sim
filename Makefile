WORKERS ?= 3
FOLLOWERS ?= 2

start:
	sudo docker compose up --scale worker=${WORKERS} --scale follower=${FOLLOWERS}

stop:
	sudo docker compose down

restart: stop start

build:
	sudo docker compose build

reset-logs:
	rm -rf logs/*
