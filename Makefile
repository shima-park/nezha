APP_NAME=nezha

.PHONY: build
build:
	go build -o $(APP_NAME) -trimpath cmd/$(APP_NAME)/main.go

.PHONY: run
run:
	./$(APP_NAME) server run --http :8080 --meta ./meta
