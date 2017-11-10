package config

import (
	"github.com/jinzhu/gorm"
	"github.com/qor/admin"
)

// DB gorm DB instance
var (
	DB    *gorm.DB
	Admin *admin.Admin
)

func init() {
	var err error
	DB, err = gorm.Open("postgres", "user=qor_test password=123 dbname=qor_test sslmode=disable host=localhost port=6000")

	if err != nil {
		panic(err)
	}

	DB.LogMode(true)

	Admin = admin.New(&admin.AdminConfig{DB: DB})
}
