package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/michaelolof/gofi-test-utils/internal/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	port := "8083"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	gin.SetMode(gin.ReleaseMode)

	data := make([]utils.SmallPayload, 100)
	for i := 0; i < 100; i++ {
		data[i] = utils.SmallPayload{ID: i, Name: fmt.Sprintf("Item %d", i)}
	}
	largeData := utils.GenerateLargeData()
	largeDataValidate := utils.GenerateLargeDataValidate()

	r := gin.New()

	r.GET("/", func(c *gin.Context) {
		c.String(200, "OK")
	})

	r.GET("/user/:name", func(c *gin.Context) {
		io.WriteString(c.Writer, c.Param("name"))
	})

	r.GET("/users/:userID/posts/:postID", func(c *gin.Context) {
		c.String(200, c.Param("userID")+":"+c.Param("postID"))
	})

	r.POST("/json", func(c *gin.Context) {
		var p utils.SmallPayload
		if err := c.ShouldBindJSON(&p); err != nil {
			c.String(400, "Bad")
			return
		}
		c.String(200, "OK")
	})

	r.GET("/json-response", func(c *gin.Context) {
		c.JSON(200, data)
	})

	r.POST("/json-large", func(c *gin.Context) {
		var p []utils.LargePayload
		if err := c.ShouldBindJSON(&p); err != nil {
			c.String(400, "Bad")
			return
		}
		c.String(200, "OK")
	})

	r.GET("/json-response-large", func(c *gin.Context) {
		c.JSON(200, largeData)
	})

	r.POST("/json-validate-small", func(c *gin.Context) {
		var p utils.SmallPayloadValidate
		if err := c.ShouldBindJSON(&p); err != nil {
			c.String(400, "Bad")
			return
		}
		c.String(200, "OK")
	})

	r.GET("/json-response-validate-small", func(c *gin.Context) {
		c.JSON(200, utils.SmallPayloadValidate{ID: 1, Name: "test"})
	})

	r.POST("/json-validate-large", func(c *gin.Context) {
		var p []utils.LargePayloadValidate
		if err := c.ShouldBindJSON(&p); err != nil {
			c.String(400, "Bad")
			return
		}
		c.String(200, "OK")
	})

	r.GET("/json-response-validate-large", func(c *gin.Context) {
		c.JSON(200, largeDataValidate)
	})

	r.POST("/multipart", func(c *gin.Context) {
		var p utils.SmallPayloadValidate
		if err := c.ShouldBind(&p); err != nil {
			c.String(400, "Bad")
			return
		}
		c.String(200, "OK")
	})

	r.POST("/formdata", func(c *gin.Context) {
		var p utils.SmallPayloadValidate
		if err := c.ShouldBind(&p); err != nil {
			c.String(400, "Bad")
			return
		}
		c.String(200, "OK")
	})

	mw := func(c *gin.Context) {
		c.Next()
	}
	r.GET("/middlewares", mw, mw, mw, mw, mw, func(c *gin.Context) {
		c.String(200, "OK")
	})

	r.GET("/query", func(c *gin.Context) {
		q := c.Query("q")
		limit := c.Query("limit")
		c.String(200, q+limit)
	})

	log.Printf("Gin listening on :%s\n", port)
	log.Fatal(r.Run(":" + port))
}
