.PHONY: build deploy destroy test test-verbose test-coverage test-handlers test-services test-utils test-repository test-logger test-config test-main test-utils-test

build:
	rm -f deployment.zip
	rm -f bootstrap
	GOOS=linux GOARCH=arm64 go build -o bootstrap main.go
	zip deployment.zip bootstrap

deploy:
	cd pulumi-infra && pulumi up

destroy:
	cd pulumi-infra && pulumi destroy --yes

test:
	@go test ./... 2>/dev/null

test-verbose:
	@go test ./... -v 2>/dev/null

test-coverage:
	@go test ./... -cover 2>/dev/null