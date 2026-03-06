WORKERS ?= 5

start:
	docker compose up --scale worker=${WORKERS}

stop:
	docker compose down

restart:
	docker compose down
	docker compose up --scale worker=${WORKERS}