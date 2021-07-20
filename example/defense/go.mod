module github.com/jjonline/go-lib-backend/example/defense

go 1.16

require (
	github.com/go-redis/redis/v7 v7.4.1
	github.com/go-redis/redis/v8 v8.8.3 // indirect
	github.com/jjonline/go-lib-backend/defense v0.0.0-20210720014741-ee748420702a
)

replace github.com/jjonline/go-lib-backend/defense => ../../defense
