SERVICE_NAME = interceptor
TEST_SERVICE_NAME = example
DIST_DIR = dist
MAIN_FILE = cmd/interceptor/main.go
TEST_MAIN_FILE = example/main.go
TESTDATA_DIR = testdata
COMPILER = /opt/homebrew/opt/go/libexec/pkg/tool/darwin_arm64/compile
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

build: clean
	@echo "Building '${SERVICE_NAME}'..."
	@go build -o "${DIST_DIR}/${SERVICE_NAME}" ${MAIN_FILE}
	@chmod +x "${DIST_DIR}/${SERVICE_NAME}"
	@echo "Dist: ${DIST_DIR}/${SERVICE_NAME}"

test:
	@echo "Testing '${SERVICE_NAME}'..."
	@go build -work -toolexec '${PWD}/${DIST_DIR}/${SERVICE_NAME}' -o "${DIST_DIR}/${TEST_SERVICE_NAME}" ${TEST_MAIN_FILE}
	

local: export WORK = ${PWD}/${TESTDATA_DIR}
local: export TOOLEXEC_IMPORTPATH = net/http
local:
	@${PWD}/${DIST_DIR}/${SERVICE_NAME} \
	${COMPILER} -o '$$WORK/b061/_pkg_.a' -trimpath '$$WORK/b061=>' -p net/http \
	-lang=go1.26 -std -complete -buildid jqQFaCrtiZzSFoy6ICEi/jqQFaCrtiZzSFoy6ICEi \
	-goversion go1.26.0 -c=12 -shared -nolocalimports \
	-importcfg '$$WORK/b061/importcfg' -pack /opt/homebrew/opt/go/libexec/src/net/http/client.go