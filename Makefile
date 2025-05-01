tag = latest

gen:
	@oapi-codegen --config=config.yaml ./contracts/tester/v1/openapi.yaml
dev: gen
	@go run main.go
build: gen
	@docker build . -t ms-tester:${tag}
	@#docker push ms-tester:${tag}
