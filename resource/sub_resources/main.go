package main

import (
	"fmt"
	"net/http"

	"github.com/fatih/color"
	"github.com/jinzhu/configor"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/qor/admin"
	"github.com/qor/media"
	"github.com/qor/media/oss"
	"github.com/qor/oss/s3"
	"github.com/qor/qor-example-cases/config"
	"github.com/qor/roles"
	appkitlog "github.com/theplant/appkit/log"
	"github.com/theplant/appkit/server"
)

type Order struct {
	gorm.Model
	Name       string
	OrderItems []OrderItem
}

type OrderItem struct {
	gorm.Model
	Name    string
	OrderID uint
	File    oss.OSS
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
	db := config.DB
	config := Config{}
	err := configor.Load(&config)
	if err != nil {
		panic(err)
	}

	db.LogMode(true)
	db.DropTable(&Order{}, &OrderItem{})
	db.AutoMigrate(&Order{}, &OrderItem{})
	order := &Order{}
	err = db.Create(order).Error
	if err != nil {
		panic(err)
	}
	for i := 0; i < 5; i++ {
		err = db.Create(&OrderItem{OrderID: order.ID, Name: fmt.Sprintf("Order Item %d", i)}).Error
		if err != nil {
			panic(err)
		}
	}

	oss.Storage = s3.New(&s3.Config{AccessID: config.AccessID, AccessKey: config.AccessKey, Region: config.Region, Bucket: config.Bucket})

	media.RegisterCallbacks(db)

	db.AutoMigrate(&Order{})

	adm := admin.New(&admin.AdminConfig{DB: db})
	orderR := adm.AddResource(&Order{}, &admin.Config{Permission: roles.Deny(roles.Create, roles.Anyone)})
	_ = orderR

	orderItemR, err := orderR.AddSubResource("OrderItems")
	if err != nil {
		panic(err)
	}

	_ = orderItemR

	mux := http.NewServeMux()
	adm.MountTo("/admin", mux)
	color.Green("URL: %v", "http://localhost:3000/admin/orders")
	server.ListenAndServe(server.Config{Addr: ":3000"}, appkitlog.Default(), mux)
}
