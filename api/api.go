package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SendSuccess(message string, data *json.RawMessage, c *gin.Context) (err error) {

	v := make(map[string]interface{})

	err = json.Unmarshal(*data, v)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sc": APISuccess,
		"message": message,
		"data": v,
	})
	return
}

func SendSuccessString(message string, data string, c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"sc": APISuccess,
		"message": message,
		"data": data,
	})
}

func SendSuccessNoData(message string, c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"sc": APISuccess,
		"message": message,
		"data": "",
	})
}

func SendFail(message string, c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"sc": APIFailed,
		"message": message,
	})
}