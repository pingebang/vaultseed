package database

import (
	"log"
	"vaultseed-backend/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() error {
	var err error
	DB, err = gorm.Open(sqlite.Open("vaultseed.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	// 自动迁移表结构
	err = DB.AutoMigrate(
		&models.User{},
		&models.EncryptedContent{},
	)
	if err != nil {
		return err
	}

	log.Println("Database connected and migrated successfully")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
