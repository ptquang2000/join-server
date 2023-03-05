package controllers

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ptquang2000/join-server/models"
)

func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d%02d/%02d", year, month, day)
}

func StartServer() {
	router := gin.Default()
	router.Delims("{[{", "}]}")
	router.SetFuncMap(template.FuncMap{
		"formatAsDate": formatAsDate,
	})

	router.Static("/static", "./static")
	router.LoadHTMLFiles("./templates/index.html")

	router.GET("/gateways", func(c *gin.Context) {
		c.JSONP(http.StatusOK, models.ReadGateways())
	})

	router.GET("/end-devices", func(c *gin.Context) {
		c.JSONP(http.StatusOK, models.ReadEndevices())
	})

	router.GET("/frames", func(c *gin.Context) {
		c.JSONP(http.StatusOK, models.ReadGateways())
	})

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"now": time.Date(2017, 0o7, 0o1, 0, 0, 0, 0, time.UTC),
		})
	})

	router.Run(":8080")
}