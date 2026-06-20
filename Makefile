build:
	rm -rf gwatch && go build -ldflags="-s -w" -o gwatch cmd/gwatch/main.go