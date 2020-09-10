PROJECT = jetson-monitor

BASE_PATH := $(shell pwd)
BUILD_PATH := $(BASE_PATH)/bin

build:
	go build -o $(BUILD_PATH)/$(PROJECT) main.go

clean:
	rm -rf $(BUILD_PATH)