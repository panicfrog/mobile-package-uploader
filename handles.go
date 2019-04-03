package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go_ipa_uploader/config"
	"howett.net/plist"
	"io"
	"io/ioutil"
	_ "net/http"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/mholt/archiver"
	"go_ipa_uploader/api"
	neturl "net/url"
	"os"
)

func route(r *gin.Engine) {
	r.GET("/hello", hello)
	r.POST("/uploadIpa", upload_ipa)
}

func hello(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "world",
	})
}

func upload_ipa(c *gin.Context) {
	file, formErro := c.FormFile("file")

	if formErro != nil {
		api.SendFail(formErro.Error(), c)
		return
	}

	filename := file.Filename

	if !strings.HasSuffix(filename, ".ipa") {
		api.SendFail("invalid format for ipa file", c)
		return
	}

	current := time.Now().Format("20060102150405")
	ipa_path := config.Config.FilesPath.IpaPath + "/" + current + " " + filename + ".zip"
	out, erro := os.Create(ipa_path)
	if erro != nil {
		api.SendFail(erro.Error(), c)
		return
	}
	defer out.Close()
	w, openErr := file.Open()

	if openErr != nil {
		api.SendFail(openErr.Error(), c)
		return
	}
	io.Copy(out, w)

	downloadPath, aliUploaderErr := aliyunOSSUpload(current+filename, ipa_path)
	if aliUploaderErr != nil {
		api.SendFail(aliUploaderErr.Error(), c)
		return
	}

	tem := config.Config.FilesPath.TemPath + "/" + current
	//tem := dir + "/tem/" + current
	unarchiveErr := archiver.Unarchive(ipa_path, tem)

	if unarchiveErr != nil {
		api.SendFail(unarchiveErr.Error(), c)
		return
	}

	plistPath, plistDir, genErr := generatePlist(current, downloadPath)
	if genErr != nil {
		api.SendFail(genErr.Error(), c)
		return
	}

	plistDownload, plistUploadErr := aliyunOSSUpload(current + "manifest.plist", plistPath)
	if plistUploadErr != nil {
		api.SendFail(plistUploadErr.Error(), c)
		return
	}
	actionURL := "itms-services://?action=download-manifest&url=" + plistDownload
	api.SendSuccessString("上传成功", actionURL, c)
	_ = os.Remove(ipa_path)
	_ = os.RemoveAll(plistDir)
	_ = os.RemoveAll(tem)
}

func aliyunOSSUpload(filename string, localPath string) (downloadURL string, err error) {
	accessKeyId := config.Config.Aliyun.AccessKeyId
	accessKeySecret := config.Config.Aliyun.AccessKeySecret
	endPoint := config.Config.Aliyun.EndPoint
	bucketName := config.Config.Aliyun.AliyunBucket
	client, err := oss.New(endPoint, accessKeyId, accessKeySecret)
	downloadURL = ""
	if err != nil {
		return
	}

	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return
	}

	err = bucket.PutObjectFromFile(filename, localPath)
	if err == nil {
		p := "https://" + bucketName + ".oss-cn-shenzhen.aliyuncs.com" + "/" + neturl.PathEscape(filename)
		downloadURL = p
		return
	}
	return
}

func generatePlist(current string, downloadURL string) (plistPath string, plistDir string,genErr error) {
	tem := config.Config.FilesPath.TemPath + "/" + current
	payloadDir := tem + "/Payload"
	dirs, _ := ioutil.ReadDir(payloadDir)
	infoPlistPath := ""
	if len(dirs) > 0 && dirs[0].IsDir() {
		infoPlistPath = payloadDir + "/" + dirs[0].Name() + "/info.plist"
	} else {
		genErr = errors.New("读取文件夹出错")
		return
	}
	bplist, readErr := ioutil.ReadFile(infoPlistPath)
	if readErr != nil {
		genErr = readErr
		return
	}
	v := make(map[string]interface{})
	_, pUnmaErro := plist.Unmarshal(bplist, v)
	if pUnmaErro != nil {
		genErr = pUnmaErro
		return
	}
	bundleId := v["CFBundleIdentifier"].(string)
	version := v["CFBundleShortVersionString"].(string)
	//buildVersion := v["CFBundleVersion"].(string)
	displayName := v["CFBundleName"].(string)

	metadata := map[string]string{
		"bundle-identifier": bundleId,
		"bundle-version":    version,
		"kind":              "software",
		"title":             displayName,
	}

	asset := map[string]string{
		"kind": "software-package",
		"url":  downloadURL,
	}

	manifast := map[string]interface{}{
		"items": []map[string]interface{}{
			map[string]interface{}{
				"assets": []map[string]string{
					asset,
				},
				"metadata": metadata,
			},
		},
	}
	bp, marshalErr := plist.Marshal(manifast, plist.BinaryFormat)
	if marshalErr != nil {
		genErr = marshalErr
		return
	}
	bpDir := config.Config.FilesPath.PlistsPath + "/" + current
	createErr := os.Mkdir(bpDir, os.ModePerm)
	if createErr != nil {
		genErr = createErr
		return
	}
	bpPath := config.Config.FilesPath.PlistsPath + "/" + current + "/" + "manifast.plist"
	ioErr := ioutil.WriteFile(bpPath, bp, os.ModePerm)
	if ioErr != nil {
		genErr = ioErr
		return
	}
	plistPath = bpPath
	plistDir = bpDir
	return
}
