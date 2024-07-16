run:
	go run ./cmd/server/ -port 3000

build:
	go build -o bin/server ./cmd/server/

bundles:
	go run scripts/cacheBundles.go


style:
	npx tailwindcss -i ./assets/css/input.css -o ./assets/css/main.css --watch
