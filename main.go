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
	Materials   []Material `gorm:"many2many:crafts_materials;"`
}

type NewMaterial struct {
	Name     string `json:"name" binding:"required"`
	ML_label string `json:"MLLabel" binding:"required"`
}

type NewCraft struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	ImageURL    string `json:"imageUrl" binding:"required"`
	MaterialIds []uint `json:"materialIds" binding:"required"`
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

	r.POST("/materials", func(c *gin.Context) {
		var toCreate NewMaterial

		if err := c.BindJSON(&toCreate); err != nil {
			return
		}

		material := Material{Name: toCreate.Name, ML_label: toCreate.ML_label}

		if result := db.Create(&material); result.Error != nil {
			c.Status(http.StatusInternalServerError)
		} else {
			c.Status(http.StatusOK)
		}

	})

	// r.POST("/crafts", func(c *gin.Context) {
	// 	var jsonBody NewCraft

	// 	if err := c.BindJSON(&jsonBody); err != nil {
	// 		fmt.Println(err)
	// 		return
	// 	}
	// 	fmt.Println(jsonBody)
	// 	toCreate := Craft{
	// 		Name:        jsonBody.Name,
	// 		Description: jsonBody.Description,
	// 		ImageURL:    jsonBody.ImageURL,
	// 	}

	// 	fmt.Println(toCreate)

	// 	c.Status(http.StatusOK)
	// })
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
