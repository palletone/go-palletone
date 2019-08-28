# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gptn android ios gptn-cross swarm evm all test clean
.PHONY: gptn-linux gptn-linux-386 gptn-linux-amd64 gptn-linux-mips64 gptn-linux-mips64le
.PHONY: gptn-linux-arm gptn-linux-arm-5 gptn-linux-arm-6 gptn-linux-arm-7 gptn-linux-arm64
.PHONY: gptn-darwin gptn-darwin-386 gptn-darwin-amd64
.PHONY: gptn-windows gptn-windows-386 gptn-windows-amd64

# Define variables needed for mirroring
DOCKER_TAG = 0.6
BASE_DOCKER_TAG = $(DOCKER_TAG)
DOCKER_NS = palletone
BASE_DOCKER_NS = $(DOCKER_NS)

GOBIN = $(shell pwd)/build/bin
GO ?= latest
BUILD_DIR = $(shell pwd)/build

gptn:
	go run -mod=vendor build/ci.go install ./cmd/gptn
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gptn\" to launch gptn."

mainnet:
	go build -mod=vendor -tags "mainnet" ./cmd/gptn
	@echo "Done building."
	@echo "Run \"./gptn\" to launch mainnet node."

all:
	build/env.sh go run -mod=vendor build/ci.go install

	
golang-baseimage: 
	docker pull palletone/goimg
golang-baseimage-dev:
	vm/baseimages/dev/tarPro.sh
	docker build -t palletone/goimg vm/baseimages/dev/
	vm/baseimages/dev/del.sh
docker:
	@mkdir -p $(BUILD_DIR)/images/gptn
	@cat images/gptn/Dockerfile.in \
                | sed -e 's|_BASE_NS_|$(BASE_DOCKER_NS)|g' \
                | sed -e 's|_NS_|$(DOCKER_NS)|g' \
                | sed -e 's|_BASE_TAG_|$(BASE_DOCKER_TAG)|g' \
                | sed -e 's|_TAG_|$(DOCKER_TAG)|g' \
                | sed -e 's|_GPTN_PATH_|$(GOBIN)|g' \
                > $(BUILD_DIR)/images/gptn/Dockerfile
	@cp -rf $(GOBIN)/gptn $(BUILD_DIR)/images/gptn
	@cp -rf images/gptn/entrypoint.sh $(BUILD_DIR)/images/gptn
	@echo "Successful generation of mirror files"

android:
	build/env.sh go run build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/gptn.aar\" to use the library."

ios:
	build/env.sh go run build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Geth.framework\" to use the library."

test: all
	build/env.sh go run -mod=vendor build/ci.go test

lint: ## Run linters.
	build/env.sh go run -mod=vendor build/ci.go lint

clean:
	rm -fr build/_workspace $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

gptn-cross: gptn-linux gptn-darwin gptn-windows gptn-android gptn-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/gptn-*

gptn-linux: gptn-linux-386 gptn-linux-amd64 gptn-linux-arm gptn-linux-mips64 gptn-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/gptn-linux-*

gptn-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/gptn
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/gptn-linux-* | grep 386

gptn-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/gptn
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gptn-linux-* | grep amd64

gptn-linux-arm: gptn-linux-arm-5 gptn-linux-arm-6 gptn-linux-arm-7 gptn-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/gptn-linux-* | grep arm

gptn-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/gptn
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/gptn-linux-* | grep arm-5

gptn-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/gptn
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/gptn-linux-* | grep arm-6

gptn-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/gptn
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/gptn-linux-* | grep arm-7

gptn-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/gptn
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/gptn-linux-* | grep arm64

gptn-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/gptn
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/gptn-linux-* | grep mips

gptn-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/gptn
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/gptn-linux-* | grep mipsle

gptn-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/gptn
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/gptn-linux-* | grep mips64

gptn-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/gptn
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/gptn-linux-* | grep mips64le

gptn-darwin: gptn-darwin-386 gptn-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/gptn-darwin-*

gptn-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/gptn
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/gptn-darwin-* | grep 386

gptn-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/gptn
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gptn-darwin-* | grep amd64

gptn-windows: gptn-windows-386 gptn-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/gptn-windows-*

gptn-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/gptn
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/gptn-windows-* | grep 386

gptn-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/gptn
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gptn-windows-* | grep amd64
