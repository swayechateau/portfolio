# Description: Makefile for building and running the application
# Author: Swaye Chateau
# Last Modified: 2024-07-22
# Usage: make <target>
# Options:
# 	build: Builds the Go application
# 	build-linux: Builds the Go application for Linux
# 	css-build: Builds CSS files
# 	css-watch: Watches CSS files
# 	docker-build: Builds Docker image
# 	docker-run: Runs Docker container
# 	docker-stop: Stops Docker container
# 	run: Runs the Go application
# 	dev: Runs Go application in development mode
# 	prod: Runs Go application in production mode
# 	stop: Stops Docker container

# Directory for built binaries and CSS files
BUILD_DIR=build
PROJECT_NAME=portfolio

# Builds the Go application
build:
	@echo "Building Go application"
	GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) go build -o $(BUILD_DIR)/server main.go
	@echo "Build complete"

# Builds the Go application for Linux
build-linux:
	@echo "Building Go application for Linux"
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/server main.go
	@echo "Build complete"

# Builds CSS files
css-build:
	@echo "Building CSS"
	npm run build:css
	@echo "CSS build complete"

# Watches CSS files
css-watch:
	@echo "Watching CSS"
	npm run watch:css
	@echo "CSS watch started"

# Builds Docker image
docker-build:
	@echo "Building Docker image"
	docker build -t $(PROJECT_NAME) .
	@echo "Docker image built"

# Runs Docker container
docker-run: docker-build
	@echo "Running Docker container"
	docker run -p 5050:5050 $(PROJECT_NAME)
	@echo "Docker container running"

# Stops Docker container
docker-stop:
	@echo "Stopping Docker container"
	docker stop $(shell docker ps -a -q)
	@echo "Docker container stopped"

# Runs the Go application
run:
	@echo "Running Go application"
	go run main.go
	@echo "Go application running"

# Runs Go application in development mode
dev: css-watch
	@echo "Running Docker container in development mode"
	docker compose -f docker-compose.dev.yml up --build
	@echo "Docker container running"

# Runs Go application in production mode
prod:
	@echo "Running Docker container in production mode"
	docker compose up -d --build
	@echo "Docker container running"

# Stops Docker container
stop:
	@echo "Stopping Docker container"
	docker compose down
	@echo "Docker container stopped"
