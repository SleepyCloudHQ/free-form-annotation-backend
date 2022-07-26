package utils

import (
	"backend/app/models"
	"errors"
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Init_db() (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	if sqliteFilePath := os.Getenv("SQLITE_FILE_PATH"); sqliteFilePath != "" {
		db, err = gorm.Open(sqlite.Open(sqliteFilePath))

	} else if mysqlHost := os.Getenv("MYSQL_HOST"); mysqlHost != "" {
		user := os.Getenv("MYSQL_USER")
		pass := os.Getenv("MYSQL_PASS")
		dbName := os.Getenv("MYSQL_DBNAME")
		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, mysqlHost, dbName)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	} else {
		return nil, errors.New("DB not configured")
	}

	if err != nil {
		return nil, err
	}

	if jointTableErr := db.SetupJoinTable(&models.User{}, "Datasets", &models.UserDataset{}); jointTableErr != nil {
		return nil, jointTableErr
	}

	return db, nil
}
