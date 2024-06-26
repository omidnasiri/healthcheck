.PHONY: mock-alpha
mock-alpha:
	go run ./mocks/alpha/alpha.go

.PHONY: mock-beta
mock-beta:
	go run ./mocks/beta/beta.go

.PHONY: mock-webhook
mock-webhook:
	go run ./mocks/webhook/webhook.go

.PHONY: build
build:
	go build -o ./healthcheck cmd/main.go

.PHONY: run
run:build
	./healthcheck