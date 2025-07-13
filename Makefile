APP_NAME = azdevops

VERSION ?= dev

DIST_DIR = dist

.PHONY: all build clean

all: build

build: clean
	@echo "🔧 Compilando binarios para múltiples plataformas..."
	GOOS=linux GOARCH=amd64   go build -o $(DIST_DIR)/$(VERSION)/linux-amd64/$(APP_NAME)
	GOOS=linux GOARCH=arm64   go build -o $(DIST_DIR)/$(VERSION)/linux-arm64/$(APP_NAME)
	GOOS=darwin GOARCH=amd64  go build -o $(DIST_DIR)/$(VERSION)/darwin-amd64/$(APP_NAME)
	GOOS=darwin GOARCH=arm64  go build -o $(DIST_DIR)/$(VERSION)/darwin-arm64/$(APP_NAME)
	GOOS=windows GOARCH=amd64 go build -o $(DIST_DIR)/$(VERSION)/windows-amd64/$(APP_NAME).exe
	@echo "✅ Binarios generados en la carpeta $(DIST_DIR)"

clean:
	@echo "🧹 Limpiando binarios anteriores..."
	rm -rf $(DIST_DIR)/*
	mkdir -p $(DIST_DIR)/$(VERSION)/linux-amd64
	mkdir -p $(DIST_DIR)/$(VERSION)/linux-arm64
	mkdir -p $(DIST_DIR)/$(VERSION)/darwin-amd64
	mkdir -p $(DIST_DIR)/$(VERSION)/darwin-arm64
	mkdir -p $(DIST_DIR)/$(VERSION)/windows-amd64
