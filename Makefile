# Makefile for the Go project

BUILD_OUTPUT=out/daiv

.PHONY: all build build-worklog run delete-test-plugin create-test-plugin build-test-plugin test-plugin

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
	$(BUILD_OUTPUT)

standup: build
	$(BUILD_OUTPUT) standup --prompt

delete-test-plugin: build
	rm -rf plugins/daiv-test

create-test-plugin: delete-test-plugin
	$(BUILD_OUTPUT) plugin create test --dir ./plugins

build-test-plugin: create-test-plugin
	cd ./plugins/daiv-test && go mod tidy && go build --buildmode=plugin -o ./out/daiv-test.so

test-plugin: build-test-plugin
	$(BUILD_OUTPUT) plugin install ./plugins/daiv-test/out/daiv-test.so
