run:
	npx nodemon --exec go run server/main.go --signal SIGTERM --ignore node_modules/ -e go,json,tmpl,js


bundles:
	go run scripts/cacheBundles.go
