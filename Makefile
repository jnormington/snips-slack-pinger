# PROFILE_TYPE can be of func or html
PROFILE_TYPE ?=func
COVERAGE_FILE=cover.out

test: unit-test coverage

unit-test:
	@go test ./... -coverprofile ${COVERAGE_FILE} -tags test

coverage:
	@echo "\n*** Coverage output ***"
	@go tool cover -${PROFILE_TYPE}=${COVERAGE_FILE}
