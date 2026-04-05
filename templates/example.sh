#!/usr/bin/env bash
# Build and test script for crit-ide

set -euo pipefail

# Configuration
PROJECT_NAME="crit-ide"
BUILD_DIR="./build"
BINARY="${BUILD_DIR}/${PROJECT_NAME}"
GO_VERSION="1.24"
COVERAGE_THRESHOLD=70

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_go() {
    if ! command -v go &>/dev/null; then
        log_error "Go is not installed"
        exit 1
    fi

    local version
    version=$(go version | grep -oP '\d+\.\d+')
    log_info "Go version: ${version}"
}

# Clean build artifacts
clean() {
    log_info "Cleaning build directory..."
    rm -rf "${BUILD_DIR}"
    mkdir -p "${BUILD_DIR}"
}

# Build the project
build() {
    log_info "Building ${PROJECT_NAME}..."
    go build -o "${BINARY}" ./cmd/ide

    if [ -f "${BINARY}" ]; then
        local size
        size=$(du -h "${BINARY}" | cut -f1)
        log_info "Binary built: ${BINARY} (${size})"
    else
        log_error "Build failed!"
        return 1
    fi
}

# Run tests with coverage
test_all() {
    log_info "Running tests..."
    go test ./... -v -coverprofile="${BUILD_DIR}/coverage.out" 2>&1 | tee "${BUILD_DIR}/test.log"

    local coverage
    coverage=$(go tool cover -func="${BUILD_DIR}/coverage.out" | tail -1 | awk '{print $3}' | tr -d '%')

    if (( $(echo "${coverage} < ${COVERAGE_THRESHOLD}" | bc -l) )); then
        log_warn "Coverage ${coverage}% is below threshold ${COVERAGE_THRESHOLD}%"
    else
        log_info "Coverage: ${coverage}%"
    fi
}

# Run static analysis
lint() {
    log_info "Running go vet..."
    go vet ./...

    if command -v staticcheck &>/dev/null; then
        log_info "Running staticcheck..."
        staticcheck ./...
    fi
}

# Main
main() {
    local command="${1:-all}"

    check_go

    case "${command}" in
        clean)  clean ;;
        build)  clean && build ;;
        test)   test_all ;;
        lint)   lint ;;
        all)    clean && build && test_all && lint ;;
        *)
            echo "Usage: $0 {clean|build|test|lint|all}"
            exit 1
            ;;
    esac

    log_info "Done!"
}

main "$@"
