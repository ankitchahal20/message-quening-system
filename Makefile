GOCMD := go
GOBUILDCMD := ${GOCMD} build
GOINSTALLCMD := ${GOCMD} install
GOPROXY := direct

# The flags '-s -w' help slim down binary file size by removing debugger info (does not affect stack traces)
GO_LD_FLAGS = -s -w
GO_BUILD_FLAGS = -ldflags "${GO_LD_FLAGS}"

.PHONY: default
default: clean setup build run

.PHONY: test
test:
	${GOCMD} test ./... -cover

.PHONY: run
run:
	cd cmd;$(GOCMD) run main.go

.PHONY: setup
setup:
	GOPROXY=${GOPROXY} ${GOINSTALLCMD}

.PHONY: build
build: build-linux

.PHONY: build-linux
build-linux: out
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 \
		${GOBUILDCMD} ${GO_BUILD_FLAGS} \
			-o bin/message-queue-system cmd/main.go
	GOARCH=386 GOOS=linux CGO_ENABLED=0 \
		${GOBUILDCMD} ${GO_BUILD_FLAGS} \
			-o bin/message-queue-system-386 cmd/main.go

.PHONY: clean
clean:
	go clean -modcache

out:
	mkdir -p bin
