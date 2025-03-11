# Makefile for the Go project

BUILD_OUTPUT=out/daiv

.PHONY: all build build-worklog run

# Default target: build and run the project
all: run

# Build the project
build:
	go build -o $(BUILD_OUTPUT) .

build-worklog: build
	cd core_plugins/worklog && go build --buildmode=plugin -o out/daiv-worklog.so
	cd core_plugins/worklog && cp out/daiv-worklog.so ~/.daiv/plugins/

# Run the built executable with the 'standup' argument
run: build-worklog
	$(BUILD_OUTPUT) standup --prompt
