APP_NAME = azdevops

VERSION ?= dev

DIST_DIR = dist

LDFLAGS = -s -w -X azuredevops/azdevops.Version=$(VERSION)

.PHONY: all build clean test vet install

all: build

build: clean
	@echo "🔧 Compilando binarios para múltiples plataformas..."
	GOOS=linux GOARCH=amd64   go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(VERSION)/linux-amd64/$(APP_NAME)
	GOOS=linux GOARCH=arm64   go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(VERSION)/linux-arm64/$(APP_NAME)
	GOOS=darwin GOARCH=amd64  go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(VERSION)/darwin-amd64/$(APP_NAME)
	GOOS=darwin GOARCH=arm64  go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(VERSION)/darwin-arm64/$(APP_NAME)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(VERSION)/windows-amd64/$(APP_NAME).exe
	@echo "✅ Binarios generados en la carpeta $(DIST_DIR)"

install:
	go install -ldflags "$(LDFLAGS)" .

test:
	go test ./...

vet:
	go vet ./...

clean:
	@echo "🧹 Limpiando binarios anteriores..."
	rm -rf $(DIST_DIR)/*
	mkdir -p $(DIST_DIR)/$(VERSION)/linux-amd64
	mkdir -p $(DIST_DIR)/$(VERSION)/linux-arm64
	mkdir -p $(DIST_DIR)/$(VERSION)/darwin-amd64
	mkdir -p $(DIST_DIR)/$(VERSION)/darwin-arm64
	mkdir -p $(DIST_DIR)/$(VERSION)/windows-amd64
