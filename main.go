package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/xujiajun/nutsdb"
)

var inMemory *nutsdb.DB
var err error

// Article is an entity.
type Article struct {
	//Parsing JSON into the struct
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {

	// setup persistent memory
	opt := nutsdb.DefaultOptions
	opt.Dir = "/tmp"
	inMemory, err = nutsdb.Open(opt)
	if err != nil {
		log.Fatal(err)
	}
	defer inMemory.Close()

	// Gin framework routes
	route := gin.Default()
	// route path eg:
	route.POST("/article", CreateArticle)
	route.GET("/articles/", GetAllArticles)
	route.GET("/article/:id", GetArticle) // path variable `id`
	route.PUT("/article/:id", UpdateArticle)
	route.DELETE("/article/:id", DeleteArticle)

	// listening at :8080
	route.Run(":8080")
}

//CreateArticle func creates an article
func CreateArticle(c *gin.Context) {
	var article Article
	var network bytes.Buffer
	enc := gob.NewEncoder(&network)

	c.BindJSON(&article)
	err := enc.Encode(article)
	if err != nil {
		c.JSON(400, err)
	}
	bucket := "article"
	err = inMemory.Update(func(tx *nutsdb.Tx) error {
		if err := tx.Put(bucket, []byte(fmt.Sprint(article.ID)), network.Bytes(), 0); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(404, err.Error())
	} else {
		c.JSON(200, article)
	}

}

//GetAllArticles func returns all articles
func GetAllArticles(c *gin.Context) {
	var articles []Article
	var article Article
	bucket := "article"
	err = inMemory.View(func(tx *nutsdb.Tx) error {
		if entries, err := tx.GetAll(bucket); err != nil {
			return err
		} else {
			for _, entry := range entries {
				newBuff := bytes.NewBuffer(entry.Value)
				dec := gob.NewDecoder(newBuff)
				dec.Decode(&article)
				articles = append(articles, article)
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(404, err.Error())
	} else {
		c.JSON(200, articles)
	}
}

//GetArticle func returns an article by ID
func GetArticle(c *gin.Context) {
	id := c.Params.ByName("id")
	var article Article
	bucket := "article"

	err = inMemory.View(func(tx *nutsdb.Tx) error {
		key := []byte(id)
		if e, err := tx.Get(bucket, key); err != nil {
			return err
		} else {
			newBuff := bytes.NewBuffer(e.Value)
			dec := gob.NewDecoder(newBuff)
			dec.Decode(&article)
		}
		return nil
	})

	if err != nil {
		c.JSON(404, "record not found")
	} else {
		c.JSON(200, article)
	}
}

//UpdateArticle func updates an article by ID
func UpdateArticle(c *gin.Context) {
	id := c.Params.ByName("id")
	var article Article
	c.BindJSON(&article)
	bucket := "article"

	err = inMemory.View(func(tx *nutsdb.Tx) error {
		key := []byte(id)
		if _, err := tx.Get(bucket, key); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(404, "record not found")
	} else {
		var network bytes.Buffer
		enc := gob.NewEncoder(&network)
		enc.Encode(article)
		err = inMemory.Update(func(tx *nutsdb.Tx) error {
			if err := tx.Delete(bucket, []byte(id)); err != nil {
				return err
			}
			if err := tx.Put(bucket, []byte(fmt.Sprint(article.ID)), network.Bytes(), 0); err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			c.JSON(404, err.Error())
		} else {
			c.JSON(200, article)
		}
	}
}

//DeleteArticle func deletes an article
func DeleteArticle(c *gin.Context) {
	id := c.Params.ByName("id")
	bucket := "article"

	err = inMemory.Update(func(tx *nutsdb.Tx) error {
		if err := tx.Delete(bucket, []byte(id)); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(404, err.Error())
	}
	if err == nil {
		c.JSON(200, "Article is deleted")
	}
}
