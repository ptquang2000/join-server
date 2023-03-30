package controllers

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
	"strconv"
	
	"github.com/gin-gonic/gin"
	"github.com/ptquang2000/lorawan-server/models"
)

var router *gin.Engine

func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d%02d/%02d", year, month, day)
}

func SetupDashboardAPI() {

	router.GET("/gateways", func(c *gin.Context) {
		c.JSONP(http.StatusOK, models.ReadGateways())
	})
	
	router.POST("/gateways", func(c *gin.Context) {
		var gateway models.Gateway
		if err := c.BindJSON(&gateway); err != nil {
			return
		}
		tx := gateway.Create()
		if tx.Error != nil {
			c.AbortWithStatusJSON(http.StatusConflict, tx.Error)
			return
		}
		c.IndentedJSON(http.StatusCreated, gateway)
	})

	router.DELETE("/gateways/:id", func(c *gin.Context) {
		id := c.Param("id")
		res, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, err)
			return
		}
		tx := models.DeleteGatewayById(uint32(res))

		if tx.RowsAffected == 0 {
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			c.AbortWithStatus(http.StatusNoContent)
		}
	})

	router.GET("/end-devices", func(c *gin.Context) {
		c.JSONP(http.StatusOK, models.ReadEndDevices())
	})

	router.POST("/end-devices", func(c *gin.Context) {
		var endDevice models.EndDevice
		if err := c.BindJSON(&endDevice); err != nil {
			fmt.Println(err)
			return
		}		
		tx := endDevice.Create()
		if tx.Error != nil {
			c.AbortWithStatusJSON(http.StatusConflict, tx.Error)
			return
		}
		c.IndentedJSON(http.StatusCreated, endDevice)
	})

	router.DELETE("/end-devices/:id", func(c *gin.Context) {
		id := c.Param("id")
		res, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, err)
			return
		}
		tx := models.DeleteEndDeviceById(uint32(res))

		if tx.RowsAffected == 0 {
			c.AbortWithStatus(http.StatusNotFound)
		} else {
			c.AbortWithStatus(http.StatusNoContent)
		}
	})


	router.GET("/frames", func(c *gin.Context) {
		c.JSONP(http.StatusOK, models.ReadFrames())
	})

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"now": time.Date(2017, 0o7, 0o1, 0, 0, 0, 0, time.UTC),
		})
	})

	router.GET("/appkey", func(c *gin.Context) {
		c.JSONP(http.StatusOK, models.GenerateAppkey())
	})

}

func StartServer() {
	router = gin.Default()
	router.Delims("{[{", "}]}")
	router.SetFuncMap(template.FuncMap{
		"formatAsDate": formatAsDate,
	})

	router.Static("/static", "./static")
	router.LoadHTMLFiles("./templates/index.html")

	SetupDashboardAPI()

	router.Run(":8080")
}