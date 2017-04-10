package main

import (
	"fmt"
	"github.com/chen-zyc/gweb"
	"net/http"
)

func main() {
	s := gweb.NewServer()

	s.GET("/ping", func(c *gweb.Context) {
		c.Status(http.StatusOK)
		c.String("pong")
	})
	s.GET("/hello/:name", func(c *gweb.Context) {
		name := c.Param("name")
		c.Status(http.StatusOK)
		c.JSON(gweb.H{
			"name":    name,
			"message": "welcome to gweb!",
		})
	})
	s.GET("/gweb/query", func(c *gweb.Context) {
		name := c.Query("name")
		c.Status(http.StatusOK)
		c.String(name)
	})
	s.POST("/gweb/post", func(c *gweb.Context) {
		name := c.PostForm("name")
		c.Status(http.StatusOK)
		c.String(name)
	})

	s.Run(":8080", gweb.PanicHandlerOption(panicHandler))
}

func panicHandler(c *gweb.Context, err interface{}) {
	fmt.Println("panic message:", err)
}
