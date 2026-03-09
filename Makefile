WORKERS ?= 5

start:
	sudo docker compose up --scale worker=${WORKERS}

stop:
	sudo docker compose down

restart: stop start

build:
	sudo docker compose build