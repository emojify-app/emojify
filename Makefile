test_unit:
	@echo "Execute unit tests"
	@go test -v -race `go list ./... | grep -v functional_tests`
	@echo ""

test_functional:
	@echo "Execute functional tests"
	@cd functional_tests && docker-compose up -d 
	@echo ""
	docker run -it --network functional_tests_default \
		-e GO111MODULE=on -v "${GOPATH}:/go" \
		-w /go/src/github.com/emojify-app/emojify golang:latest \
		/bin/bash -c 'go get ./... && cd functional_tests && go test -v --godog.format=pretty --godog.random'
	@echo ""
	@cd functional_tests && docker-compose down
	@echo ""

# go get ./... && cd functional_tests && go test -v --godog.format=pretty --godog.random
