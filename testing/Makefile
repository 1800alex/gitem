
TARGETS:=all bootstrap build clean up down logs help testclient
.PHONY: $(TARGETS)

.DEFAULT_GOAL:=all

all: build

bootstrap:
	@echo "Bootstrapping test environment..."
	@./bootstrap.sh

build:
	@echo "Building test environment..."
	@cd multi && docker-compose build
	@cd multi && docker-compose pull

up: down
	@echo "Starting test environment..."
	@cd multi && docker-compose up -d

down:
	@echo "Stopping test environment..."
	-@cd multi && docker-compose down

clean: down
	@echo "Cleaning test environment..."
	-@cd multi && docker-compose rm -f
	-@cd multi && docker-compose down --rmi all --volumes --remove-orphans

logs:
	-@cd multi && docker-compose logs -f

testclient:
	@cd multi && docker-compose exec testclient bash

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  bootstrap   Bootstrap test environment"
	@echo "  build       Build test environment"
	@echo "  up          Start test environment"
	@echo "  down        Stop test environment"
	@echo "  clean       Clean test environment"
	@echo "  logs        Show test environment logs"
	@echo "  testclient  Connect to testclient container"
	@echo "  help        Show this help message"
