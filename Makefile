run:
	go run cmd/server/main.go -port 3000

build:
	go build -o bin/server cmd/server/main.go && systemctl restart gates.service && systemctl status gates.service

bundles:
	go run scripts/cacheBundles.go


style:
	npx tailwindcss -i ./assets/css/input.css -o ./assets/css/main.css --watch
