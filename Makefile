DSN = 'postgres://postgres:postgres@localhost:5432/gophermart?sslmode=disable'

build:
	go build -o cmd/gophermart/gophermart cmd/gophermart/*.go

run: build
	./cmd/gophermart/gophermart -d=$(DSN) -l=info -s=supersecret -e=86400 -r=http://localhost:8008

run-accrual:
	./cmd/accrual/accrual_darwin_amd64 -d=$(DSN) -a=localhost:8008	

stattest:
	go vet -vettool=statictest ./...

autotest: build
	./gophermarttest -test.v -test.run=^TestGophermart$$ -gophermart-binary-path=cmd/gophermart/gophermart -gophermart-host=localhost -gophermart-port=8080 -gophermart-database-uri=$(DSN) -accrual-binary-path=cmd/accrual/accrual_darwin_amd64 -accrual-host=localhost -accrual-port=8008 -accrual-database-uri=$(DSN)

mocks:
	mockgen -destination=internal/server/handlers/mocks/mock_storage.go -source=internal/server/handlers/base_handler.go -package=mocks Storage,Hasher

