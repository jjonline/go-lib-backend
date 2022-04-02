module github.com/jjonline/go-lib-backend/example/queue

go 1.15

require (
	github.com/go-redis/redis/v8 v8.8.3
	github.com/jjonline/go-lib-backend/queue v0.0.0-20210720014741-ee748420702a
	go.uber.org/zap v1.21.0
)

replace github.com/jjonline/go-lib-backend/queue => ../../queue
