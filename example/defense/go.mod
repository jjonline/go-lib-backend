module github.com/jjonline/go-mod-library/example/defense

go 1.16

require (
	github.com/go-redis/redis/v7 v7.4.0
	github.com/jjonline/go-mod-library/defense v1.0.0
)

replace github.com/jjonline/go-mod-library/defense => ../../defense
