module github.com/jjonline/go-lib-backend/example/crontab

go 1.16

require (
	github.com/jjonline/go-lib-backend/crontab v1.1.0
	github.com/jjonline/go-lib-backend/logger v1.8.0
	go.uber.org/zap v1.16.0
)

replace github.com/jjonline/go-lib-backend/crontab => ../../crontab
