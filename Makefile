test:
	go test ./... -cover

build:
	go build

run: LastWatchedBackend
	./LastWatchedBackend

LastWatchedBackend: build

clean:
	rm LastWatchedBackend
	rm server.log