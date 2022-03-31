.PHONY: run stop

run:
	docker-compose up --build

stop:
	docker-compose down
