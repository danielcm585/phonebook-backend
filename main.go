package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Contact struct {
	gorm.Model
	Name  string `gorm:"column:name"`
	Phone string `gorm:"column:phone"`
}

var db *gorm.DB

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	var err error
	db, err = gorm.Open(postgres.Open(psqlInfo), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.AutoMigrate(&Contact{})
}

func GetAllContacts(c *gin.Context) {
	var contacts []Contact
	result := db.Find(&contacts)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(200, contacts)
}

func CreateNewContact(c *gin.Context) {
	var newContact Contact
	if err := c.ShouldBindJSON(&newContact); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	result := db.Create(&newContact)
	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(201, newContact)
}

func GetContactById(c *gin.Context) {
	var contact Contact
	id := c.Param("id")
	result := db.First(&contact, "id = ?", id)
	if result.Error != nil {
		c.JSON(404, gin.H{"error": "Contact not found"})
		return
	}
	c.JSON(200, contact)
}

func EditContact(c *gin.Context) {
	var contact Contact
	id := c.Param("id")

	if err := db.First(&contact, "id = ?", id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Contact not found"})
		return
	}

	var updatedContact Contact
	if err := c.ShouldBindJSON(&updatedContact); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	db.Model(&contact).Updates(updatedContact)

	c.JSON(200, contact)
}

func DeleteContact(c *gin.Context) {
	var contact Contact
	id := c.Param("id")

	result := db.Where("id = ?", id).Delete(&contact)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Contact not found"})
		return
	}

	c.JSON(200, gin.H{"message": "Contact deleted"})
}

func main() {
	app := gin.Default()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	}))

	app.GET("/contacts", GetAllContacts)
	app.POST("/contacts", CreateNewContact)

	app.GET("/contacts/:id", GetContactById)
	app.PUT("/contacts/:id", EditContact)
	app.DELETE("/contacts/:id", DeleteContact)

	app.Run(":8000")
}
