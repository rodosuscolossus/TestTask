# Имя Docker-образа для PostgreSQL
POSTGRES_IMAGE = postgres:latest
# Имя Docker-контейнера для PostgreSQL
POSTGRES_CONTAINER = postgres-db-task

# Имя Docker-образа для вашего приложения
APP_IMAGE = task-service-app
# Имя Docker-контейнера для вашего приложения
APP_CONTAINER = task-service-app-container

# Имя сети Docker
DOCKER_NETWORK = my-docker-network

# Параметры для подключения к PostgreSQL
POSTGRES_USER = user
POSTGRES_PASSWORD = password
POSTGRES_DB = taskdb
POSTGRES_HOST = postgres-db-task

# Цель по умолчанию
all: run

# Создание Docker-сети
create-network:
	docker network create $(DOCKER_NETWORK)

# Запуск PostgreSQL в Docker
run-postgres: create-network
	docker run --name $(POSTGRES_CONTAINER) -e POSTGRES_USER=$(POSTGRES_USER) -e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) -e POSTGRES_DB=$(POSTGRES_DB) -p 5432:5432 -d --network $(DOCKER_NETWORK) $(POSTGRES_IMAGE)

# Сборка Docker-образа
build:
	docker build -t $(APP_IMAGE) .

# Запуск Docker-контейнера
run: run-postgres build
	docker run -p 8080:8080 --name $(APP_CONTAINER) -e POSTGRES_USER=$(POSTGRES_USER) -e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) -e POSTGRES_DB=$(POSTGRES_DB) -d --network $(DOCKER_NETWORK) $(APP_IMAGE)

# Запуск тестов внутри Docker-контейнера
test:
	docker exec -it $(APP_CONTAINER) go test ./...

# Остановка и удаление Docker-контейнеров
stop: stop-app stop-postgres

# Остановка и удаление Docker-контейнера с приложением
stop-app:
	docker stop $(APP_CONTAINER)
	docker rm $(APP_CONTAINER)

# Остановка и удаление Docker-контейнера с PostgreSQL
stop-postgres:
	docker stop $(POSTGRES_CONTAINER)
	docker rm $(POSTGRES_CONTAINER)

stop-network:
	docker network rm $(DOCKER_NETWORK)

# Очистка (остановка, удаление, удаление образов)
clean: stop
	docker rmi $(APP_IMAGE)

.PHONY: all create-network run-postgres build run test stop stop-app stop-postgres clean