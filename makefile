.PHONY: build deploy destroy

build:
	rm -f deployment.zip
	rm -f bootstrap
	GOOS=linux GOARCH=arm64 go build -o bootstrap main.go
	zip deployment.zip bootstrap

deploy:
	cd pulumi-infra && pulumi up

destroy:
	cd pulumi-infra && pulumi destroy --yes