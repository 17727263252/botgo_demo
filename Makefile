PROJECT_NAME=botgo_mo

build:
	go build -tags=jsoniter -o ./${PROJECT_NAME} ./cmd/server

run:
	go run ./server/comet/main.go

stop:
	pkill -f target/logic
	pkill -f target/comet

lint:
	golangci-lint run ./... --skip-dirs="docs" --build-tags="!apitest"


test:
