run:
	npx nodemon --exec go run main.go --signal SIGTERM --ignore node_modules/ -e go,json,tmpl,js
