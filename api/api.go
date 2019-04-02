package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SendSuccess(message string, data *json.RawMessage, v interface{}, c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"sc": APISuccess,
		"message": message,
		"data": json.Unmarshal(*data, v),
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