
build:
	GOOS=linux GOARCH=arm64 go build -o bootstrap main.go
	zip deployment.zip bootstrap

deploy:
	cd pulumi-infra && pulumi up