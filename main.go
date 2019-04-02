package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	// 设置最大上传为100Mib
	r.MaxMultipartMemory = 100 << 20
	route(r)
	r.Run()
}
