package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Material struct {
	gorm.Model
	Name     string
	ML_label string
}

type Craft struct {
	gorm.Model
	Name        string
	Description string
	ImageURL    string
	materials   []Material `gorm:"many2many:crafts_materials;"`
}

func main() {
	dsn := "host=localhost user=postgres password=root dbname=capstone port=5432 sslmode=disable TimeZone=Asia/Bangkok"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		fmt.Println("database connection error")
	} else {
		fmt.Println("database connection success")
	}

	db.AutoMigrate(&Craft{}, &Material{})

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
