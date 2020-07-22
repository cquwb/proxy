all: server client
server:
	cd app/server && GOOS=linux GOARCH=amd64 go build -o ../../bin/proxyserver
client:
	cd app/client && GOOS=linux GOARCH=amd64 go build -o ../../bin/proxyclient
