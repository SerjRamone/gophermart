DSN = 'postgres://postgres:postgres@localhost:5432/gophermart?sslmode=disable'

build:
	go build -o cmd/gophermart/gophermart cmd/gophermart/*.go

run: build
	./cmd/gophermart/gophermart -d=$(DSN) -l=info -s=supersecret -e=86400
