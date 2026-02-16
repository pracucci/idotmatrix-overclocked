.PHONY: build generate-previews

build:
	go build -o idm-cli ./cmd/cli

generate-previews:
	go run ./tools/preview-generator
