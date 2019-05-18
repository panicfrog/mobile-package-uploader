package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SendSuccess(message string, v interface{}, c *gin.Context) {

	res := SucccStruct{APISuccess, message, v}
	response, err := json.Marshal(res)
	if err != nil {
		panic(err)
	}

	var m map[string]interface{}

	err = json.Unmarshal(response, &m)

	if err != nil {
		panic(err)
	}

	c.JSON(http.StatusOK, m)
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