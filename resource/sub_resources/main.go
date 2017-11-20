package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jinzhu/configor"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/qor/admin"
	"github.com/qor/media"
	"github.com/qor/media/oss"
	"github.com/qor/oss/s3"
	"github.com/qor/qor-example-cases/config"
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
	Name       string
	OrderID    uint
	File       oss.OSS
	CustomFile OSS
}

type OSS struct {
	oss.OSS
}

func (o OSS) GetURLTemplate(option *media.Option) (url string) {
	fmt.Println("option = ", option, "oss = ", o)
	if url = option.Get("URL"); url == "" {
		url = fmt.Sprintf("/%s/%d.{{extension}}", time.Now().Format("20060102"), time.Now().UnixNano())
	}

	return
}

var domain = "//qor3-agc-develop.s3.ap-northeast-1.amazonaws.com"

func (o *OSS) Scan(data interface{}) (err error) {
	switch values := data.(type) {
	case []byte:
		if strings.HasPrefix(string(values), "{") && strings.HasSuffix(string(values), "}") {
			return json.Unmarshal(values, o)
		}
		if string(values) != "" {
			o.Url = domain + string(values)
		}
		fmt.Printf("[]byte %+v\n", string(values))
	case string:
		fmt.Printf("string %+v\n", string(values))
		return o.Scan([]byte(values))
	case []string:
		for _, str := range values {
			fmt.Printf("[]string %+v\n", str)
			if err := o.Scan(str); err != nil {
				return err
			}
		}
	default:
		fmt.Printf("default %+v\n", data)
		return o.OSS.Scan(data)
	}
	return
}

func (o OSS) Value() (driver.Value, error) {
	if o.Delete {
		return nil, nil
	}
	fmt.Printf("Value %+v\n", o.Url)
	return strings.Replace(o.Url, domain, "", -1), nil
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
	media.RegisterCallbacks(db)

	if os.Getenv("DATA") != "" {
		db.DropTable(&Order{}, &OrderItem{})
	}

	db.AutoMigrate(&Order{}, &OrderItem{})

	if os.Getenv("DATA") != "" {
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
	}

	oss.Storage = s3.New(&s3.Config{AccessID: config.AccessID, AccessKey: config.AccessKey, Region: config.Region, Bucket: config.Bucket})

	adm := admin.New(&admin.AdminConfig{DB: db})
	orderR := adm.AddResource(&Order{})
	// orderR := adm.AddResource(&Order{}, &admin.Config{Permission: roles.Deny(roles.Create, roles.Anyone)})
	_ = orderR

	orderItemR, err := orderR.AddSubResource("OrderItems")
	orderItemR.UseTheme("grid")
	orderItemR.IndexAttrs("CustomFile", "Name")
	if err != nil {
		panic(err)
	}

	_ = orderItemR

	mux := http.NewServeMux()
	adm.MountTo("/admin", mux)
	color.Green("URL: %v", "http://localhost:3000/admin/orders")
	server.ListenAndServe(server.Config{Addr: ":3000"}, appkitlog.Default(), mux)
}
