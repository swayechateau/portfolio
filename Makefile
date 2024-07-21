# Builds go application
build-linux:
	@echo "Building go application"
	GOOS=linux GOARCH=amd64 go build -o build/server main.go
	@echo "Build complete"

docker-build: build-linux
	@echo "Building docker image"
	docker build -t portfolio .
	@echo "Docker image built"

dev:
	@echo "Running docker container"
	docker compose up -f docker-compose.dev.yml --build
	@echo "Docker container running"