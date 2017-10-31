package main

import (
	"net/http"

	"github.com/fatih/color"
	"github.com/jinzhu/configor"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/qor/admin"
	"github.com/qor/media"
	"github.com/qor/media/oss"
	"github.com/qor/oss/s3"
	appkitlog "github.com/theplant/appkit/log"
	"github.com/theplant/appkit/server"
)

type Order struct {
	gorm.Model
	File oss.OSS
}

// run with dummy data
// MODE=data go run main.go
type Config struct {
	AccessID  string `env:"QOR_AWS_ACCESS_KEY_ID"`
	AccessKey string `env:"QOR_AWS_SECRET_ACCESS_KEY"`
	Region    string `env:"QOR_AWS_REGION"`
	Bucket    string `env:"QOR_AWS_BUCKET"`
}

func main() {
	db, err := gorm.Open("postgres", "user=qor_test password=123 dbname=qor_test sslmode=disable host=localhost port=6000")
	if err != nil {
		panic(err)
	}
	config := Config{}
	err = configor.Load(&config)
	if err != nil {
		panic(err)
	}

	oss.Storage = s3.New(&s3.Config{AccessID: config.AccessID, AccessKey: config.AccessKey, Region: config.Region, Bucket: config.Bucket})

	media.RegisterCallbacks(db)

	db.AutoMigrate(&Order{})

	adm := admin.New(&admin.AdminConfig{DB: db})
	orderR := adm.AddResource(&Order{})
	_ = orderR
	mux := http.NewServeMux()
	adm.MountTo("/admin", mux)
	color.Green("URL: %v", "http://localhost:3000/admin/orders")
	server.ListenAndServe(server.Config{Addr: ":3000"}, appkitlog.Default(), mux)
}
