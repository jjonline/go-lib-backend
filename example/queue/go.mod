module github.com/jjonline/go-mod-library/example/queue

go 1.15

require (
	github.com/go-redis/redis/v8 v8.8.3
	github.com/jjonline/go-mod-library/queue v0.8.0
	github.com/kr/pretty v0.1.0 // indirect
	go.uber.org/zap v1.17.0
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace github.com/jjonline/go-mod-library/queue => ../../queue
