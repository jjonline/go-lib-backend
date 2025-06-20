module github.com/jjonline/go-lib-backend/example/crontab

go 1.24

require (
	github.com/jjonline/go-lib-backend/crond v0.0.0-20220401021235-e9ae05f536b4
	github.com/jjonline/go-lib-backend/logger v0.0.0-20220401021235-e9ae05f536b4
)

require github.com/robfig/cron/v3 v3.0.1 // indirect

replace (
	github.com/jjonline/go-lib-backend/crond => ../../crond
	github.com/jjonline/go-lib-backend/logger => ../../logger
)
