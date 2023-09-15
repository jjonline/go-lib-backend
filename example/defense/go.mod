module github.com/jjonline/go-lib-backend/example/defense

go 1.18

require (
	github.com/go-redis/redis/v7 v7.4.1
	github.com/jjonline/go-lib-backend/defense v0.0.0-20210720014741-ee748420702a
)

require github.com/go-redis/redis/v8 v8.11.5 // indirect

replace github.com/jjonline/go-lib-backend/defense => ../../defense
