// Package testdb 提供按环境变量选择方言的测试数据库。
//
// 默认使用 SQLite 内存库（零依赖，本地开发直接 `go test`）；
// 设置 TEST_MYSQL_DSN 或 TEST_POSTGRES_DSN 时连接真实服务器，
// 同一套测试即可在三种数据库上运行。CI 的 database-matrix job
// 注入这两个变量完成 SQLite/MySQL/PostgreSQL 三库验证。
//
// 该包仅被测试二进制引用，不会链接进生产程序。
package testdb

import (
	"fmt"
	"os"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Open 打开测试数据库并同步 common 的主/日志库方言状态。
// 出错时 panic：仅用于 TestMain，初始化失败没有继续的意义。
//
// 方言选择：TEST_DB 显式指定（mysql/postgres/sqlite）时优先生效，
// 否则按 TEST_MYSQL_DSN / TEST_POSTGRES_DSN 是否存在选择，默认 SQLite。
// CI 的 database-matrix job 同时注入两个 DSN，靠步骤级 TEST_DB 切换方言。
func Open() *gorm.DB {
	dbType := Dialect()
	var dialector gorm.Dialector
	switch dbType {
	case common.DatabaseTypeMySQL:
		dialector = mysql.Open(os.Getenv("TEST_MYSQL_DSN"))
	case common.DatabaseTypePostgreSQL:
		dialector = postgres.Open(os.Getenv("TEST_POSTGRES_DSN"))
	default:
		// 共享缓存的内存库让多个连接看到同一份 schema；不能用裸 ":memory:"
		// 加单连接——业务代码里存在事务内二次查询（如 GetDBTimestamp），
		// 单连接会在这种模式下自死锁。
		dialector = sqlite.Open("file:ob_testdb?mode=memory&cache=shared&_busy_timeout=5000")
	}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("testdb: failed to open %s database: %v", dbType, err))
	}
	common.SetDatabaseTypes(dbType, dbType)
	return db
}

// Dialect 返回当前选择的方言，测试可据此断言或跳过方言特定用例。
func Dialect() common.DatabaseType {
	switch os.Getenv("TEST_DB") {
	case "mysql":
		return common.DatabaseTypeMySQL
	case "postgres":
		return common.DatabaseTypePostgreSQL
	default:
		switch {
		case os.Getenv("TEST_MYSQL_DSN") != "":
			return common.DatabaseTypeMySQL
		case os.Getenv("TEST_POSTGRES_DSN") != "":
			return common.DatabaseTypePostgreSQL
		default:
			return common.DatabaseTypeSQLite
		}
	}
}
