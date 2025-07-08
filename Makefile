APP_NAME = azdevops

VERSION ?= dev

DIST_DIR = dist

.PHONY: all build clean

all: build

build: clean
	@echo "🔧 Compilando binarios para múltiples plataformas..."
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64   go build -o $(DIST_DIR)/$(APP_NAME)-linux-amd64-$(VERSION)
	GOOS=linux GOARCH=arm64   go build -o $(DIST_DIR)/$(APP_NAME)-linux-arm64-$(VERSION)
	GOOS=darwin GOARCH=amd64  go build -o $(DIST_DIR)/$(APP_NAME)-darwin-amd64-$(VERSION)
	GOOS=darwin GOARCH=arm64  go build -o $(DIST_DIR)/$(APP_NAME)-darwin-arm64-$(VERSION)
	GOOS=windows GOARCH=amd64 go build -o $(DIST_DIR)/$(APP_NAME)-windows-amd64-$(VERSION).exe
	@echo "✅ Binarios generados en la carpeta $(DIST_DIR)"

clean:
	@echo "🧹 Limpiando binarios anteriores..."
	rm -rf $(DIST_DIR)/*
	mkdir -p $(DIST_DIR)
