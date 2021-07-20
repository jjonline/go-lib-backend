module github.com/jjonline/go-lib-backend/example/crontab

go 1.16

require (
	github.com/jjonline/go-lib-backend/crontab v0.0.0-20210720014741-ee748420702a
	github.com/jjonline/go-lib-backend/logger v0.0.0-20210720014741-ee748420702a
	go.uber.org/zap v1.18.1
)

replace github.com/jjonline/go-lib-backend/crontab => ../../crontab
