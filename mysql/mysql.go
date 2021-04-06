package mysql

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

var (
	DB  *gorm.DB
	err error
)

// 数据库配置
type MysqlConfig struct {
	Database     string // 数据库名
	User         string // 用户名
	Password     string // 密码
	Host         string // 主机IP
	Port         int    // 端口
	Charset      string // 编码
	ParseTime    bool   // 时区
	Loc          string // UTC
	MaxIdleConns int
	MaxOpenConns int
}

func New(config MysqlConfig) {
	url := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		config.User, config.Password,
		config.Host, config.Port,
		config.Database, config.Charset, config.ParseTime, config.Loc)
	DB, err = gorm.Open(mysql.Open(url), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	panicError(err)
	err = DB.Use(dbresolver.Register(dbresolver.Config{}).
		SetMaxOpenConns(config.MaxOpenConns).
		SetMaxIdleConns(config.MaxIdleConns))
	panicError(err)
}
