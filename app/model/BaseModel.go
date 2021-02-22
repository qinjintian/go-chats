package model

import (
	"fmt"
	"github.com/go-ini/ini"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
)

type BaseModel struct {}

var DB *gorm.DB

func InitDB(cfg *ini.File) (*gorm.DB, error) {
	driver := strings.ToLower(cfg.Section(ini.DefaultSection).Key("DB_CONNECTION").MustString("mysql"))
	host := cfg.Section(ini.DefaultSection).Key("DB_HOST").MustString("127.0.0.1")
	port := cfg.Section(ini.DefaultSection).Key("DB_PORT").MustString("3306")
	database := cfg.Section(ini.DefaultSection).Key("DB_DATABASE").MustString("forge")
	username := cfg.Section(ini.DefaultSection).Key("DB_USERNAME").MustString("forge")
	password := cfg.Section(ini.DefaultSection).Key("DB_PASSWORD").MustString("")

	switch driver {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, host, port, database)
		var err error
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("Database connection failed error:  %v", err)
		}

	case "sqlserver":

	case "postgres", "postgre", "postgresql":

	default:
		return nil, fmt.Errorf("%v SQL database type is not supported", driver)
	}

	return DB, nil
}