SERVICE_NAME = interceptor
TEST_SERVICE_NAME = example
DIST_DIR = dist
MAIN_FILE = cmd/interceptor/main.go
TEST_HTTP_MAIN_FILE = example/http/main.go
TEST_VALKEY_MAIN_FILE = example/valkey/main.go
TESTDATA_DIR = testdata
COMPILER = /opt/homebrew/opt/go/libexec/pkg/tool/darwin_arm64/compile
INTERCEPTOR_LOG_LEVEL ?= info

run:
	@go run ${MAIN_FILE}

lint:
	@golangci-lint run ./...

tests:
	@go test ./...

coverage:
	@go test -coverprofile=coverage.out ./...

clean:
	@rm -rf ${DIST_DIR}/

build:
	@echo "Building '${SERVICE_NAME}'..."
	INTERCEPTOR_LOG_LEVEL=${INTERCEPTOR_LOG_LEVEL} go build -o "${DIST_DIR}/${SERVICE_NAME}" ${MAIN_FILE}
	@chmod +x "${DIST_DIR}/${SERVICE_NAME}"
	@echo "Dist: ${DIST_DIR}/${SERVICE_NAME}"


test-valkey: build
	@echo "Building valkey module with '${SERVICE_NAME}'..."
	INTERCEPTOR_LOG_LEVEL=${INTERCEPTOR_LOG_LEVEL} go build -work -toolexec '${PWD}/${DIST_DIR}/${SERVICE_NAME}' -o "${DIST_DIR}/${TEST_SERVICE_NAME}" ${TEST_VALKEY_MAIN_FILE}
	@echo "--------------------------------"
	@echo "Run binary: ${DIST_DIR}/${TEST_SERVICE_NAME}"

test-http: build
	@echo "Building http module test with '${SERVICE_NAME}'..."
	INTERCEPTOR_LOG_LEVEL=${INTERCEPTOR_LOG_LEVEL} go build -work -toolexec '${PWD}/${DIST_DIR}/${SERVICE_NAME}' -o "${DIST_DIR}/${TEST_SERVICE_NAME}" ${TEST_HTTP_MAIN_FILE}
	@echo "--------------------------------"
	@echo "Run binary: ${DIST_DIR}/${TEST_SERVICE_NAME}"

local: export WORK = ${PWD}/${TESTDATA_DIR}
local: export TOOLEXEC_IMPORTPATH = net/http
local:
	@${PWD}/${DIST_DIR}/${SERVICE_NAME} \
	${COMPILER} -o $$WORK/b062/_pkg_.a -trimpath $$WORK/b062=> -p net/http \
	-lang=go1.26 -std -complete -buildid jqQFaCrtiZzSFoy6ICEi/jqQFaCrtiZzSFoy6ICEi \
	-goversion go1.26.0 -c=12 -shared -nolocalimports \
	-importcfg $$WORK/b062/importcfg -pack /opt/homebrew/opt/go/libexec/src/net/http/client.go