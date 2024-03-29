EXECS   := $(wildcard examples/*)
TARGETS := ${EXECS:examples/%=%}

TESTA   := ${shell go list ./... | grep -v /examples/ }

BRANCH   := ${shell git branch --show-current}
REVCNT   := ${shell git rev-list --count $(BRANCH)}
REVHASH  := ${shell git log -1 --format="%h"}

LDFLAGS  := -X main.version=${BRANCH}.${REVCNT}.${REVHASH}

all: check build

check: gen lint test

cover:
	go test -coverprofile=cover.out ${TESTA} && \
	go tool cover -func=cover.out

gen:
	go generate ./...

lint:
	golangci-lint run ./...

test:
	go test -count 1 ${TESTA}

build: ${TARGETS}
	@echo ":: Done"

${TARGETS}:
	@echo ":: Building $@"
	go build -ldflags '${LDFLAGS}' -o bin/$@ examples/$@/main.go

.PHONY:  test
