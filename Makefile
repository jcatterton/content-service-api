run:
	go run cmd/svr/main.go
run-with-docker:
	docker-compose -f ./docker/docker-compose.yaml up -d --build --force-recreate
test:
	go test ./...
coverage:
	go test -failfast=true ./... -coverprofile cover.out
	go tool cover -html=cover.out
	rm cover.out
mocks:
	mockery --name=DBHandler --recursive=true --case=underscore --output=./pkg/testhelper/mocks;
	mockery --name=ExtHandler --recursive=true --case=underscore --output=./pkg/testhelper/mocks;
	mockery --name=Requestor --recursive=true --case=underscore --output=./pkg/testhelper/mocks;
