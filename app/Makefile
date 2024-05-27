init:
	go mod init github.com/SakyaSumedh/irsa
install:
	go mod download
clean:
	go mod tidy
build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GOFLAGS=-trimpath go build -o ./demo-irsa
run:
	go run main.go
docker-build:
	docker build -t <docker-hub-ac>/demo-irsa .
docker-push:
	docker push <docker-hub-ac>/demo-irsa:latest
docker-run:
	docker run --rm -d -p 8080:8080 <docker-hub-ac>/demo-irsa:latest