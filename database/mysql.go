package database

import (
	"fmt"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	db   *gorm.DB
	once sync.Once
)

// MysqlConfig mysql 配置
type MysqlConfig struct {
	Host              string
	Port              int
	User              string
	Password          string
	Dbname            string
	Enable            bool
	MaxIdleConnection int
	MaxOpenConnection int
}

// InitDB 初始化数据库
func InitDB(config MysqlConfig) *gorm.DB {
	once.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.User,
			config.Password,
			config.Host,
			config.Port,
			config.Dbname,
		)
		var err error
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info), // 设置日志级别
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   "forum_", // 全局表前缀
				SingularTable: false,    // false → 表名复数，true → 表名单数
			},
			// Logger: logger.Default.LogMode(logger.Error), // 设置日志级别
			// Logger: &logc.CustomLogger{}, // 设置日志级别
		})
		if err != nil {
			logx.Errorf("failed to connect database: %+v", err)
			panic(any(fmt.Errorf("failed to connect database: %w", err).Error()))
		}

		sqlDB, err := db.DB()
		if err != nil {
			logx.Errorf("Failed to get sql.DB from gorm.DB: %+v", err)
			panic(any(fmt.Errorf("failed to get sql.DB from gorm.DB: %w", err).Error()))
		}

		// 设置连接池
		sqlDB.SetMaxOpenConns(config.MaxOpenConnection) // 最大连接数
		sqlDB.SetMaxIdleConns(config.MaxIdleConnection) // 最大空闲连接数
		sqlDB.SetConnMaxLifetime(10 * time.Minute)      // 连接最大生命周期
	})

	return db
}
