module github.com/jjonline/go-lib-backend/example/migrate

go 1.20

require (
	github.com/go-sql-driver/mysql v1.7.1
	github.com/jjonline/go-lib-backend/migrate v0.0.0-20220613140507-5de65279835c
	github.com/spf13/cobra v1.7.0
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/logrusorgru/aurora v2.0.3+incompatible // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

replace github.com/jjonline/go-lib-backend/migrate => ../../migrate
