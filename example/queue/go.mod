module github.com/jjonline/go-lib-backend/example/queue

go 1.18

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/jjonline/go-lib-backend/logger v0.0.0-20220613140507-5de65279835c
	github.com/jjonline/go-lib-backend/queue v0.0.0-20210720014741-ee748420702a
	go.uber.org/zap v1.21.0
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gin-gonic/gin v1.7.7 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742 // indirect
	github.com/ugorji/go/codec v1.1.7 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	google.golang.org/protobuf v1.23.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/gorm v1.23.5 // indirect
)

replace github.com/jjonline/go-lib-backend/queue => ../../queue
