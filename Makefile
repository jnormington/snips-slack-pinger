# PROFILE_TYPE can be of func or html
PROFILE_TYPE ?=func
COVERAGE_FILE=cover.out

BIN_DIR=./bin
APP_NAME=ssp
APP_VERSION=0.0.1
PREFIX=${APP_NAME}-${APP_VERSION}

export GOOS ?=linux
export GOARCH ?=arm

test: unit-test coverage

unit-test:
	@go test ./... -coverprofile ${COVERAGE_FILE} -tags test

coverage:
	@echo "\n*** Coverage output ***"
	@go tool cover -${PROFILE_TYPE}=${COVERAGE_FILE}

build:
	@go build

build-all:
	GOOS=linux GOARCH=arm go build -o ${BIN_DIR}/${PREFIX}-linux-arm
	GOOS=linux GOARCH=amd64 go build -o ${BIN_DIR}/${PREFIX}-linux64
	GOOS=linux GOARCH=386 go build -o ${BIN_DIR}/${PREFIX}-linux386
	GOOS=darwin GOARCH=amd64 go build -o ${BIN_DIR}/${PREFIX}-darwinx64
