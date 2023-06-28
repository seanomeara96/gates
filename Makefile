run:
	npx nodemon --exec go run server/main.go --signal SIGTERM --ignore node_modules/ -e go,json,tmpl,js


bundles:
	go run scripts/cacheBundles.go


style:
	npx tailwindcss -i ./assets/css/input.css -o ./assets/css/main.css --watch
