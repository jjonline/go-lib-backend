module github.com/jjonline/go-mod-library/example/crontab

go 1.16

require (
	github.com/jjonline/go-mod-library/crontab v1.1.0
	github.com/jjonline/go-mod-library/logger v1.8.0
	go.uber.org/zap v1.16.0
)

replace github.com/jjonline/go-mod-library/crontab => ../../crontab
