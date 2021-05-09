module github.com/jjonline/go-mod-library/example/queue

go 1.15

require (
	github.com/go-redis/redis/v7 v7.4.0
	github.com/jjonline/go-mod-library/queue v1.2.3
	go.uber.org/zap v1.16.0
)

replace github.com/jjonline/go-mod-library/queue => ../../queue
