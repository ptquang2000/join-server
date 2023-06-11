package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ptquang2000/lorawan-server/models"
)

var router *gin.Engine
var upgrader = websocket.Upgrader{}

type EndDeviceLiveData struct {
	ID      uint64
	FType   models.FrameType
	Time    time.Time
	Payload []byte
}

var edLiveDataChans = make(map[uint64](chan EndDeviceLiveData))

type GatewayLiveData struct {
	ID    uint64
	FType models.FrameType
	Time  time.Time
	Rssi  int8
	Snr   int16
}

var gwLiveDataChans = make(map[uint64](chan GatewayLiveData))

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

	router.GET("/frames/:limit", func(c *gin.Context) {
		limit := c.Param("limit")
		res, err := strconv.ParseUint(limit, 10, 32)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, err)
			return
		}
		c.JSONP(http.StatusOK, models.ReadLimitFrames(int(res)))
	})

	router.GET("/appkey", func(c *gin.Context) {
		c.JSONP(http.StatusOK, models.GenerateAppkey())
	})

	router.GET("/end-devices/:id/live", endDeviceLiveData)

	router.GET("/end-devices/:id/activity", func(c *gin.Context) {
		id := c.Param("id")
		res, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, err)
			return
		}
		actitivies := []EndDeviceLiveData{}
		for _, act := range models.GetEndDeviceActivities(res) {
			liveData := EndDeviceLiveData{
				ID:      uint64(act.EndDeviceID),
				Time:    act.CreatedAt,
				FType:   act.FType,
				Payload: act.Payload,
			}
			actitivies = append(actitivies, liveData)
		}
		c.JSONP(http.StatusOK, actitivies)
	})

	router.GET("/gateways/:id/live", gatewayLiveData)

	router.GET("/gateways/:id/activity", func(c *gin.Context) {
		id := c.Param("id")
		res, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, err)
		}
		actitivies := []GatewayLiveData{}
		for _, act := range models.GetGatewayActivities(res) {
			liveData := GatewayLiveData{
				ID:    uint64(act.GatewayID),
				FType: act.FType,
				Time:  act.CreatedAt,
				Rssi:  act.Rssi,
				Snr:   act.Snr,
			}
			actitivies = append(actitivies, liveData)
		}
		c.JSONP(http.StatusOK, actitivies)
	})

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

}

func endDeviceLiveData(c *gin.Context) {
	id := c.Param("id")
	res, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, err)
		return
	}
	writer, reader := c.Writer, c.Request
	connection, err := upgrader.Upgrade(writer, reader, nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	defer connection.Close()
	edLiveDataChans[res] = make(chan EndDeviceLiveData)
	go func() {
		_, _, err := connection.ReadMessage()
		if err != nil {
			close(edLiveDataChans[res])
		}
	}()

	for liveData := range edLiveDataChans[res] {
		message, err := json.Marshal(liveData)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			break
		}
		err = connection.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			break
		}
	}
	delete(edLiveDataChans, res)
}

func gatewayLiveData(c *gin.Context) {
	id := c.Param("id")
	res, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, err)
		return
	}
	writer, reader := c.Writer, c.Request
	connection, err := upgrader.Upgrade(writer, reader, nil)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	defer connection.Close()
	gwLiveDataChans[res] = make(chan GatewayLiveData)
	go func() {
		_, _, err := connection.ReadMessage()
		if err != nil {
			close(gwLiveDataChans[res])
		}
	}()

	for liveData := range gwLiveDataChans[res] {
		message, err := json.Marshal(liveData)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			break
		}
		err = connection.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			break
		}
	}
	delete(gwLiveDataChans, res)
}

func StartServer() {
	router = gin.Default()
	router.Delims("{[{", "}]}")

	router.Static("/static", "./static")
	router.LoadHTMLFiles("./templates/index.html")

	SetupDashboardAPI()

	router.Run(":8080")
}
