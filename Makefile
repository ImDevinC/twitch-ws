build-windows-x64:
	mkdir -p out/windows-x64
	GOOS=windows GOARCH=amd64 go build -o out/windows-x64/twitch-ws.exe -ldflags "-s -w" ./cmd/server.go
build-linux-x64:
	mkdir -p out/linux-x64
	GOOS=linux GOARCH=amd64 go build -o out/linux-x64/twitch-ws -ldflags "-s -w" ./cmd/server.go

clean:
	rm -rf out

release-all: clean build-windows-x64 build-linux-x64
	cd out/windows-x64 && \
	zip ../twitch-ws-windows-x64-${RELEASE_VERSION}.zip twitch-ws.exe
	cd  out/linux-x64 && \
	zip ../twitch-ws-linux-x64-${RELEASE_VERSION}.zip twitch-ws

