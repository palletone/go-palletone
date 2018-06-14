# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: gpan android ios gpan-cross swarm evm all test clean
.PHONY: gpan-linux gpan-linux-386 gpan-linux-amd64 gpan-linux-mips64 gpan-linux-mips64le
.PHONY: gpan-linux-arm gpan-linux-arm-5 gpan-linux-arm-6 gpan-linux-arm-7 gpan-linux-arm64
.PHONY: gpan-darwin gpan-darwin-386 gpan-darwin-amd64
.PHONY: gpan-windows gpan-windows-386 gpan-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest

gpan:
	build/env.sh go run build/ci.go install ./cmd/gpan
	@echo "Done building."
	@echo "Run \"$(GOBIN)/gpan\" to launch gpan."

swarm:
	build/env.sh go run build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

all:
	build/env.sh go run build/ci.go install

android:
	build/env.sh go run build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/gpan.aar\" to use the library."

ios:
	build/env.sh go run build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Geth.framework\" to use the library."

test: all
	build/env.sh go run build/ci.go test

lint: ## Run linters.
	build/env.sh go run build/ci.go lint

clean:
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

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

gpan-cross: gpan-linux gpan-darwin gpan-windows gpan-android gpan-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/gpan-*

gpan-linux: gpan-linux-386 gpan-linux-amd64 gpan-linux-arm gpan-linux-mips64 gpan-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/gpan-linux-*

gpan-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/gpan
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/gpan-linux-* | grep 386

gpan-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/gpan
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gpan-linux-* | grep amd64

gpan-linux-arm: gpan-linux-arm-5 gpan-linux-arm-6 gpan-linux-arm-7 gpan-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/gpan-linux-* | grep arm

gpan-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/gpan
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/gpan-linux-* | grep arm-5

gpan-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/gpan
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/gpan-linux-* | grep arm-6

gpan-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/gpan
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/gpan-linux-* | grep arm-7

gpan-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/gpan
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/gpan-linux-* | grep arm64

gpan-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/gpan
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/gpan-linux-* | grep mips

gpan-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/gpan
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/gpan-linux-* | grep mipsle

gpan-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/gpan
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/gpan-linux-* | grep mips64

gpan-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/gpan
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/gpan-linux-* | grep mips64le

gpan-darwin: gpan-darwin-386 gpan-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/gpan-darwin-*

gpan-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/gpan
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/gpan-darwin-* | grep 386

gpan-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/gpan
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gpan-darwin-* | grep amd64

gpan-windows: gpan-windows-386 gpan-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/gpan-windows-*

gpan-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/gpan
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/gpan-windows-* | grep 386

gpan-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/gpan
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/gpan-windows-* | grep amd64
