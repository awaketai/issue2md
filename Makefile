APP_NAME    := issue2md
BIN_DIR     := bin
IMAGE_TAG   := $(APP_NAME):latest
GO_FLAGS    := -trimpath -ldflags='-s -w'
WEB_ENTRY   := cmd/issue2mdweb/main.go

.PHONY: build test lint docker-build clean

build: $(BIN_DIR)/issue2md
	@if [ -f $(WEB_ENTRY) ]; then \
		go build $(GO_FLAGS) -o $(BIN_DIR)/issue2mdweb ./cmd/issue2mdweb; \
		echo "Built: $(BIN_DIR)/issue2mdweb"; \
	fi

$(BIN_DIR)/issue2md:
	go build $(GO_FLAGS) -o $(BIN_DIR)/issue2md ./cmd/issue2md
	@echo "Built: $(BIN_DIR)/issue2md"

test:
	go test ./...

lint:
	golangci-lint run ./...

docker-build:
	docker build -t $(IMAGE_TAG) .

clean:
	rm -rf $(BIN_DIR)/
