module github.com/seanomeara96/gates

go 1.23.1

toolchain go1.24.1

require github.com/mattn/go-sqlite3 v1.14.27

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/sessions v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/seanomeara96/auth v0.0.0-20250311125829-3f40be05d59a
	github.com/stripe/stripe-go/v82 v82.1.0
	golang.org/x/text v0.24.0
)

require (
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	golang.org/x/crypto v0.37.0 // indirect
)
