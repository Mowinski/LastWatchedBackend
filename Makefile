test:
	go test ./... -cover

build:
	go build

run: LastWatchedBackend
	./LastWatchedBackend

cover:
	go test ./$$APP -coverprofile=c.out
	go tool cover -html=c.out
	rm c.out

LastWatchedBackend: build

clean:
	rm LastWatchedBackend
	rm server.log
	rm c.out
