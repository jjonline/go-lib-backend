package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jjonline/go-lib-backend/example/migrate/migrations"
	"github.com/jjonline/go-lib-backend/migrate"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "migrate-test",
	Short: "migrate-test",
	Long:  "migrate-test",
	Run: func(cmd *cobra.Command, args []string) {

	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		return nil
	},
}

func init() {
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "数据库迁移相关",
		Long:  `数据库迁移相关`,
	}
	migrateCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "查看迁移文件列表和状态",
		Long:  `查看迁移文件列表和状态`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return getMigration().Status()
		},
	})
	migrateCmd.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "创建迁移文件",
		Long:  `创建迁移文件`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return getMigration().Create(args[0])
		},
		Args: cobra.ExactArgs(1),
	})
	migrateCmd.AddCommand(&cobra.Command{
		Use:   "up",
		Short: "执行迁移文件",
		Long:  `执行迁移文件`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return getMigration().ExecUp()
		},
	})
	migrateCmd.AddCommand(&cobra.Command{
		Use:   "down",
		Short: "迁移文件回滚：不指定文件名，默认回滚migrations表最后一条迁移记录",
		Long:  `迁移文件回滚：不指定文件名，默认回滚migrations表最后一条迁移记录`,
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := "" // 文件名需包含后缀
			if len(args) > 0 {
				filename = args[0]
			}
			return getMigration().ExecDown(filename)
		},
	})
	rootCmd.AddCommand(migrateCmd)
}

func getMigration() *migrate.Migrate {
	// create MySQL link
	// todo please change to your MySQL userName/Password/DatabaseName
	open, err := sql.Open("mysql", "test:test@tcp(127.0.0.1:3306)/test?parseTime=true")
	if err != nil {
		panic("mysql link error")
	}

	// init migrate Instance
	return migrate.New(migrate.Config{
		Dir:       "migrations",
		Fs:        &migrations.Sql, // fs.FS
		TableName: "migrations",
		DB:        open,
	})
}

func main() {
	// go run main.go migrate --help
	err := rootCmd.Execute()
	if err != nil {
		return
	}
}
