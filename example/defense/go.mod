module github.com/jjonline/go-lib-backend/example/defense

go 1.24

require (
	github.com/jjonline/go-lib-backend/defense v0.0.0-20210720014741-ee748420702a
	github.com/redis/go-redis/v9 v9.10.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
)

replace github.com/jjonline/go-lib-backend/defense => ../../defense
