VERSION=v1.0.0
build:
	@cd cmd && go mod tidy && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ../bin/gateway-server.bin

build-linux:
	@cd cmd && go mod tidy && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ../bin/gateway-server.bin

docker-build:
	@docker build -t dm/gateway-server:${VERSION} -f ./docker/Dockerfile .

docker-build-linux:
	@cd cmd && go mod tidy && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ../bin/gateway-server.bin
	@docker build -t dm/gateway-server:${VERSION} -f ./docker/Dockerfile-linux .
