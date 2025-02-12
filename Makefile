# Makefile for the Go project

BUILD_OUTPUT=out/daiv

.PHONY: all build run

# Default target: build and run the project
all: run

# Build the project
build:
	go build -o $(BUILD_OUTPUT) .

# Run the built executable with the 'standup' argument
run: build
	$(BUILD_OUTPUT) standup
