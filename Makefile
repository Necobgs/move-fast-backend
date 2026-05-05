include .env


start:
	@go run cmd/serve-api/main.go

migration-create:
	migrate create -ext=sql -dir=sql/migrations -seq $(name)

migration-up:
	migrate -path=sql/migrations -database "$(DB_URL)" -verbose up

migration-down:
	migrate -path=sql/migrations -database "$(DB_URL)" -verbose down -all
