module github.com/jjonline/go-lib-backend/example/queue

go 1.24

require (
	github.com/jjonline/go-lib-backend/logger v0.0.0-20220613140507-5de65279835c
	github.com/jjonline/go-lib-backend/queue v0.0.0-20210720014741-ee748420702a
	github.com/redis/go-redis/v9 v9.13.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
)

replace (
	github.com/jjonline/go-lib-backend/logger => ../../logger
	github.com/jjonline/go-lib-backend/queue => ../../queue
)
