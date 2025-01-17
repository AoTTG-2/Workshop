###########################
# Docker / Local deploy
###########################

# (Re-)start container without rebuild and cleaning dependencies (volumes/images/networks)
restart: down up
# Use this to rebuild and clean old junk (volumes/images/networks)
rebuild: clean build

up:
	docker compose up -d
down:
	docker compose down
build:
	docker compose up -d --build
clean:
	docker compose down -v --remove-orphans
	docker image prune -f
	docker volume prune -f

###########################
# Documentation
###########################

# Generate go app env documentation
godoc:
	go run cmd\docs\main.go

# Generate go app swagger documentation
swag:
	swag fmt
	swag init --ot go,json --parseInternal -g ./cmd/workshop/main.go


###########################
# Tests and Benchmarks
###########################

DOCKER_IMAGE_NAME := workshop-test-runner
DOCKERFILE_PATH := ./docker/test/Dockerfile
DOCKER_MOUNTS := -v .:/workshop

TEST_CMD := go test
TEST_ARGS := -v

BENCH_CMD := go test -bench=. -run=^\#
BENCH_ARGS := -test.count 5 -test.benchmem -test.v

# Init docker image before usage
test_build:
	docker build -t $(DOCKER_IMAGE_NAME) -f $(DOCKERFILE_PATH) .

# Run all tests
test_all: test_build test_postgres

# Tests
test_postgres:
	docker run --rm --network host --env-file .env.postgres_test $(DOCKER_MOUNTS) $(DOCKER_IMAGE_NAME) sh -c "$(TEST_CMD) $(TEST_ARGS) ./internal/repository/driver/postgres"
