.PHONY: build-cli

build-cli:
	cd apps/cli && go build -o ../../bin/interlock main.go