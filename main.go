package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ساختار داده‌ی نود (شامل ID، نام و اولویت)
type Node struct {
	ID       int    `json:"id" gorm:"primaryKey"`
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

var db *gorm.DB

func main() {
	var err error

	// اتصال به دیتابیس SQLite
	db, err = gorm.Open(sqlite.Open("nodes.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("خطا در اتصال به دیتابیس:", err)
	}

	// ایجاد جدول در صورت عدم وجود
	db.AutoMigrate(&Node{})

	// راه‌اندازی Gin برای API
	r := gin.Default()

	// API برای افزودن نود جدید
	r.POST("/add", func(c *gin.Context) {
		var newNode Node
		if err := c.ShouldBindJSON(&newNode); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ورودی نامعتبر"})
			return
		}

		result := db.Create(&newNode)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "نود با موفقیت اضافه شد", "node": newNode})
	})

	// API برای جستجوی نود بر اساس شناسه
	r.GET("/find/:id", func(c *gin.Context) {
		var node Node
		id := c.Param("id")

		result := db.First(&node, id)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "نود پیدا نشد"})
			return
		}

		c.JSON(http.StatusOK, node)
	})

	// API برای حذف نود بر اساس شناسه
	r.DELETE("/delete/:id", func(c *gin.Context) {
		var node Node
		id := c.Param("id")

		result := db.Delete(&node, id)
		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "نودی با این شناسه یافت نشد"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "نود حذف شد"})
	})

	// API برای افزایش اولویت نود مشخص‌شده
	r.POST("/boost", func(c *gin.Context) {
		type BoostInput struct {
			ID           int `json:"id"`
			NewPriority int `json:"new_priority"`
		}

		var input BoostInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ورودی نامعتبر"})
			return
		}

		var node Node
		result := db.First(&node, input.ID)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "نود پیدا نشد"})
			return
		}

		if input.NewPriority > node.Priority {
			node.Priority = input.NewPriority
			db.Save(&node)
			c.JSON(http.StatusOK, gin.H{"message": "اولویت به‌روزرسانی شد", "node": node})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "اولویت جدید باید بزرگ‌تر باشد"})
		}
	})

	// API برای رسیدگی به نود دارای بالاترین اولویت
	r.POST("/handle", func(c *gin.Context) {
		var top Node
		result := db.Order("priority DESC").First(&top)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "صف خالی است"})
			return
		}

		db.Delete(&top)
		c.JSON(http.StatusOK, gin.H{"message": "نود با اولویت بالا رسیدگی شد و حذف شد", "node": top})
	})

	// API برای نمایش صف اولویت (مرتب‌سازی بر اساس priority)
	r.GET("/queue", func(c *gin.Context) {
		var nodes []Node
		db.Order("priority DESC").Find(&nodes)
		c.JSON(http.StatusOK, nodes)
	})

	// API برای نمایش درخت (مرتب‌سازی بر اساس ID شبیه in-order BST)
	r.GET("/tree", func(c *gin.Context) {
		var nodes []Node
		db.Order("id ASC").Find(&nodes)
		c.JSON(http.StatusOK, nodes)
	})

	// اجرای سرور روی پورت 8080
	r.Run(":8080")
}