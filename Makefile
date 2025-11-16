docker-up:
	docker-compose up -d --build

docker-remove:
	-docker stop $$(docker ps -q)             
	-docker rm -f $$(docker ps -aq)           
	-docker rmi -f $$(docker images -q)
	-docker image prune -f

docker-stop:
	docker compose down --volumes --remove-orphans --rmi all

clean:
	rm -rf mocks/ coverage.html *.out *.tmp coverage_percent.txt

generate-mocks:
	mockery

test:
	./scripts/test.sh

e2e:
	go test -count=1 -tags=e2e ./e2e

run_linter:
	golangci-lint run

run_format:
	go fmt ./...
	goimports -w .

.PHONY: docker-up docker-remove docker-stop clean test e2e run_format run_linter