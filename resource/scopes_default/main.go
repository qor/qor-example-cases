package main

import (
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/qor/admin"
	"github.com/qor/qor"
	appkitlog "github.com/theplant/appkit/log"
	"github.com/theplant/appkit/server"
)

type Order struct {
	gorm.Model
	Num   string
	State string
	Price float64
}

// run with dummy data
// MODE=data go run main.go

func main() {
	db, err := gorm.Open("postgres", "user=qor_test password=123 dbname=qor_test sslmode=disable host=localhost port=6000")
	if err != nil {
		panic(err)
	}

	if os.Getenv("MODE") == "data" {
		db.DropTable(&Order{})
		db.AutoMigrate(&Order{})
		db.Create(&Order{Num: "T00001", State: "pending", Price: 1000})
		db.Create(&Order{Num: "T00002", State: "confirmed", Price: 1500})
		db.Create(&Order{Num: "T00003", State: "shipped", Price: 2000})
		db.Create(&Order{Num: "T00004", State: "cancelled", Price: 2500})
		db.Create(&Order{Num: "T00005", State: "pending", Price: 2000})
		db.Create(&Order{Num: "T00006", State: "confirmed", Price: 2000})
		db.Create(&Order{Num: "T00007", State: "shipped", Price: 2500})
		db.Create(&Order{Num: "T00008", State: "cancelled", Price: 1500})
		db.Create(&Order{Num: "T00009", State: "pending", Price: 2500})
		db.Create(&Order{Num: "T000010", State: "confirmed", Price: 2500})
		db.Create(&Order{Num: "T000020", State: "shipped", Price: 1000})
		db.Create(&Order{Num: "T000030", State: "cancelled", Price: 1000})
	} else {
		db.AutoMigrate(&Order{})
	}
	db.LogMode(true)

	adm := admin.New(&admin.AdminConfig{DB: db})
	orderR := adm.AddResource(&Order{})
	orderR.Scope(&admin.Scope{
		Name:    "Pending",
		Default: true,
		Group:   "State",
		Handler: func(d *gorm.DB, ctx *qor.Context) *gorm.DB {
			return d.Where("state = ?", "pending")
		},
	})
	orderR.Scope(&admin.Scope{
		Name:  "Confirmed",
		Group: "State",
		Handler: func(d *gorm.DB, ctx *qor.Context) *gorm.DB {
			return d.Where("state = ?", "confirmed")
		},
	})

	orderR.Scope(&admin.Scope{
		Name:  "Shipped",
		Group: "State",
		Handler: func(d *gorm.DB, ctx *qor.Context) *gorm.DB {
			return d.Where("state = ?", "shipped")
		},
	})

	orderR.Scope(&admin.Scope{
		Name:  "Cancelled",
		Group: "State",
		Handler: func(d *gorm.DB, ctx *qor.Context) *gorm.DB {
			return d.Where("state = ?", "cancelled")
		},
	})

	orderR.Scope(&admin.Scope{
		Name:    "Cheap",
		Group:   "Price",
		Default: true,
		Handler: func(d *gorm.DB, ctx *qor.Context) *gorm.DB {
			return d.Where("price <= 1500")
		},
	})
	orderR.Scope(&admin.Scope{
		Name:  "Middle",
		Group: "Price",
		Handler: func(d *gorm.DB, ctx *qor.Context) *gorm.DB {
			return d.Where("price > 1500 AND price <= 2000")
		},
	})
	orderR.Scope(&admin.Scope{
		Name:  "Expensive",
		Group: "Price",
		Handler: func(d *gorm.DB, ctx *qor.Context) *gorm.DB {
			return d.Where("price > 2000")
		},
	})

	mux := http.NewServeMux()
	adm.MountTo("/admin", mux)
	color.Green("URL: %v", "http://localhost:3000/admin/orders")
	server.ListenAndServe(server.Config{Addr: ":3000"}, appkitlog.Default(), mux)
}
