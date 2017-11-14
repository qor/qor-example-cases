package config

import (
	"os"

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
	dialect := os.Getenv("DB_DIALECT")
	source := os.Getenv("DB_SOURCE")
	if dialect == "" {
		dialect = "postgres"
	}
	if source == "" {
		source = "user=qor_test password=123 dbname=qor_test sslmode=disable host=localhost port=6000"
	}

	DB, err = gorm.Open(dialect, source)

	if err != nil {
		panic(err)
	}

	DB.LogMode(true)

	Admin = admin.New(&admin.AdminConfig{DB: DB})
}
