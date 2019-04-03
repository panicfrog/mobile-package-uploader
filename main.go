package main

import (
	"github.com/gin-gonic/gin"
	"go_ipa_uploader/config"
)

func main() {
	r := gin.Default()
	// 设置最大上传为100Mib
	r.MaxMultipartMemory = int64(config.Config.Application.MaxMultipartMemory) << 20
	route(r)
	r.Run()
}
