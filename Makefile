run:
	go run cmd/server/main.go

bundles:
	go run scripts/cacheBundles.go


style:
	npx tailwindcss -i ./assets/css/input.css -o ./assets/css/main.css --watch
