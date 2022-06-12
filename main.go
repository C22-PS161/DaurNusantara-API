package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

type MLResponse struct {
	Labels []string `json:"objects" binding:"required"`
}

type CraftJson struct {
	Id          uint   `json:"ID"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
}

func main() {
	visionAPI, found := os.LookupEnv("VISION_URL")
	if !found {
		visionAPI = "http://localhost:8081"
	}
	dsn, found := os.LookupEnv("DB_URI")
	if !found {
		dsn = "host=localhost user=postgres password=root dbname=capstone port=5432 sslmode=disable TimeZone=Asia/Bangkok"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

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

	r.POST("/crafts", func(c *gin.Context) {
		var jsonBody NewCraft

		if err := c.BindJSON(&jsonBody); err != nil {
			fmt.Println(err)
			return
		}

		toCreate := Craft{
			Name:        jsonBody.Name,
			Description: jsonBody.Description,
			ImageURL:    jsonBody.ImageURL,
		}

		if result := db.Create(&toCreate); result.Error != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		var neededMaterials []Material

		result := db.Where("id IN ?", jsonBody.MaterialIds).Find(&neededMaterials)

		if result.Error != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		err := db.Model(&toCreate).Association("Materials").Append(neededMaterials)
		if err != nil {
			fmt.Println("DB error")
			fmt.Println(err)
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Status(http.StatusOK)
	})

	r.POST("/vision", func(c *gin.Context) {
		img, err := c.FormFile("photo")
		if err != nil {
			fmt.Println(err)
			c.Status(http.StatusBadRequest)
			return
		}

		imgBytes, _ := img.Open()
		defer imgBytes.Close()
		reqBody := &bytes.Buffer{}
		writer := multipart.NewWriter(reqBody)
		writer.WriteField("threshold", "0.4")
		part, _ := writer.CreateFormFile("img", img.Filename)
		io.Copy(part, imgBytes)
		writer.Close()

		r, _ := http.NewRequest("POST", visionAPI, reqBody)
		r.Header.Add("Content-Type", writer.FormDataContentType())
		client := &http.Client{}

		resp, err := client.Do(r)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		defer resp.Body.Close()

		var jsonBody MLResponse
		json.NewDecoder(resp.Body).Decode(&jsonBody)

		fmt.Println(jsonBody)

		var queriedCrafts []Craft

		db.Model(&Craft{}).Select("crafts.*").
			Joins("JOIN crafts_materials ON crafts.id = crafts_materials.craft_id").
			Joins("JOIN materials ON crafts_materials.material_id = materials.id").
			Where("materials.ml_label IN ?", jsonBody.Labels).
			Group("crafts.id").
			Having("COUNT(*) = ?", len(jsonBody.Labels)).
			Find(&queriedCrafts)

		var responseCrafts []CraftJson

		for _, v := range queriedCrafts {
			responseCrafts = append(responseCrafts, CraftJson{
				Name:        v.Name,
				Description: v.Description,
				Id:          v.ID,
				ImageURL:    v.ImageURL,
			})
		}

		c.JSON(http.StatusOK, responseCrafts)
	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
