TARGETS:=all tidy build clean help
.PHONY: $(TARGETS)

.DEFAULT_GOAL:=all

all: help

tidy:
	@echo "Tidying go modules..."
	@go mod tidy

build:
	@echo "Building gitem..."
	@mkdir -p bin
	@go build -o bin/gitem

clean:
	@echo "Cleaning gitem..."
	@rm -rf bin

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  help  Show this help message"
	@echo "  tidy  Tidy go modules"
	@echo "  build Build gitem"
	@echo "  clean Clean gitem"
