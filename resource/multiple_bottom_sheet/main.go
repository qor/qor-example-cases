package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/mattn/go-sqlite3"

	"github.com/qor/admin"
	"github.com/qor/qor-example-cases/config"
)

// Create a GORM-backend model
type User struct {
	gorm.Model
	Name      string
	Product   Product
	ProductID uint
	Another   Another
	AnotherID uint
}

// Create another GORM-backend model
type Product struct {
	gorm.Model
	Name        string
	Description string
}

type Another struct {
	gorm.Model
	Name string
}

func main() {
	db := config.DB
	if os.Getenv("MODE") == "data" {
		db.DropTable(&User{}, &Product{}, &Another{})
		db.AutoMigrate(&User{}, &Product{}, &Another{})
		db.Create(&User{Name: "user1"})
		db.Create(&User{Name: "user2"})
		db.Create(&Product{Name: "product1"})
		db.Create(&Product{Name: "product2"})
		db.Create(&Another{Name: "another1"})
		db.Create(&Another{Name: "another2"})
	} else {
		db.AutoMigrate(&User{}, &Product{}, &Another{})
	}

	// Initalize
	Admin := config.Admin
	user := Admin.AddResource(&User{})
	product := Admin.AddResource(&Product{})
	user.Meta(&admin.Meta{Name: "Product", Type: "select_one",
		Config: &admin.SelectOneConfig{SelectMode: "bottom_sheet", RemoteDataResource: product},
	})

	another := Admin.AddResource(&Another{})
	user.Meta(&admin.Meta{Name: "Another", Type: "select_one",
		Config: &admin.SelectOneConfig{SelectMode: "bottom_sheet", RemoteDataResource: another},
	})
	// Register route

	mux := http.NewServeMux()
	// amount to /admin, so visit `/admin` to view the admin interface
	Admin.MountTo("/admin", mux)
	fmt.Println("started")
	http.ListenAndServe(":9000", mux)
}
