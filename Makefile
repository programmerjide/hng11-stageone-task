server:
	nodemon --watch 'app/**/*.go' --signal SIGTERM --exec 'go run main.go'
