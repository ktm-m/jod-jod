secret:
	node -e "console.log(require('crypto').randomBytes(32).toString('hex'))"

up:
	docker-compose up -d

down:
	docker-compose down

run:
	go run main.go