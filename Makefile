# Builds go application
build-linux:
	@echo "Building go application"
	GOOS=linux GOARCH=amd64 go build -o build/portfolio main.go
	@echo "Build complete"

docker-build: build-linux
	@echo "Building docker image"
	docker build -t portfolio .
	@echo "Docker image built"