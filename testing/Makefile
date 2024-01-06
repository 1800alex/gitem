
TARGETS:=all build clean up down logs help
.PHONY: $(TARGETS)

.DEFAULT_GOAL:=all

all: build

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

