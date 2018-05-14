plugin_name = 'selectel'

clean:
	rm -rf bin/$(plugin_name)

build: clean
	GOGC=off go build -o bin/$(plugin_name) cmd/main.go

install: build
	cp bin/$(plugin_name) /usr/local/bin/docker-machine-driver-$(plugin_name)
