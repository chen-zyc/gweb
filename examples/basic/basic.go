package main

import (
	"fmt"
	"github.com/chen-zyc/gweb"
	"net/http"
)

func main() {
	s := gweb.NewServer()

	s.GET("/ping", func(c *gweb.Context) {
		c.String(http.StatusOK, "pong")
	})
	s.GET("/hello/:name", func(c *gweb.Context) {
		name := c.Param("name")
		c.JSON(http.StatusOK, gweb.H{
			"name":    name,
			"message": "welcome to gweb!",
		})
	})
	s.GET("/gweb/query", func(c *gweb.Context) {
		name := c.Query("name")
		c.String(http.StatusOK, name)
	})
	s.POST("/gweb/post", func(c *gweb.Context) {
		name := c.PostForm("name")
		c.String(http.StatusOK, name)
	})

	v1 := s.Group("/v1", func(c *gweb.Context) {
		c.SetUserData("version", "v1")
	})
	{
		v1.GET("/test", func(c *gweb.Context) {
			version, exist := c.UserData("version")
			if !exist {
				c.String(http.StatusOK, "user data not exist")
				return
			}
			c.JSON(http.StatusOK, gweb.H{
				"version": version,
			})
		})
	}

	s.Run(":8080",
		gweb.NameOption("test"),
		gweb.PanicHandlerOption(panicHandler),
	)
}

func panicHandler(c *gweb.Context, err interface{}) {
	fmt.Println("panic message:", err)
}
