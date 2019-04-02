package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"go/types"
	"io"
	"io/ioutil"
	"log"
	_ "net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/mholt/archiver"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	"go_ipa_uploader/api"
	"howett.net/plist"
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

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		api.SendFail("invalid content for ipa file", c)
		return
	}
	current := time.Now().Format("20060102150405")
	ipa_path := dir + "/ipas/" + current + " " + filename + ".zip"
	out, erro := os.Create(ipa_path)
	defer out.Close()
	if erro != nil {
		api.SendFail(erro.Error(), c)
		return
	}
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

	tem := dir + "/tem/" + current
	unarchiveErr := archiver.Unarchive(ipa_path, tem)

	if unarchiveErr != nil {
		api.SendFail(unarchiveErr.Error(), c)
		return
	}

	plistPath, genErr := generatePlist(current, downloadPath)
	if genErr != nil {
		api.SendFail(genErr.Error(), c)
		return
	}

	plistDownload, plistUploadErr := aliyunOSSUpload("manifast.plist", plistPath)
	if plistUploadErr != nil {
		api.SendFail(plistUploadErr.Error(), c)
		return
	}
	//var p string
	//api.SendSuccess("上传成功", *[]byte(plistDownload), &p, c)
}

func aliyunOSSUpload(filename string, localPath string) (downloadURL string, err error) {
	accessKeyId := "LTAILE7u7V2G6LxX"
	accessKeySecret := "ItjQm5myPQClvqvpX1QnqraKl12GtX"
	endPoint := "http://oss-cn-shenzhen.aliyuncs.com"
	bucketName := "ipa-uploader"
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
		p := "https://ipa-uploader.oss-cn-shenzhen.aliyuncs.com" + "/" + neturl.PathEscape(filename)
		log.Println(p)
		downloadURL = p
		return
	}
	return
}

func generatePlist(current string, downloadURL string) (plistPath string, genErr error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		genErr = err
		return
	}
	tem := dir + "/tem/" + current
	payloadDir := tem + "/Payload"
	dirs, _ := ioutil.ReadDir(payloadDir)
	infoPlistPath := ""
	if len(dir) > 0 && dirs[0].IsDir() {
		infoPlistPath = payloadDir + "/" + dirs[0].Name() + "/info.plist"
	} else {
		genErr = errors.New("读取文件夹出错")
		return
	}
	//infoPlistPath := tem + "/Payload"
	bplist, readErr := ioutil.ReadFile(infoPlistPath)
	if readErr != nil {
		genErr = readErr
		return
	}
	v := make(map[string]interface{})
	plist.Unmarshal(bplist, v)
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
				"asserts": []map[string]string{
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
	bpDir := dir + "/plists/" + current
	createErr := os.Mkdir(bpDir, os.ModePerm)
	if createErr != nil {
		genErr = createErr
		return
	}
	bpPath := dir + "/plists/" + current + "/" + "manifast.plist"
	ioErr := ioutil.WriteFile(bpPath, bp, os.ModePerm)
	if ioErr != nil {
		genErr = ioErr
		return
	}
	plistPath = bpPath
	return
}

func qiniuUpload(filename string, localPath string) (err error) {
	bucket := "ipa_uploader"
	key := filename
	accessKey := "MGcrk96J5WU5NMPvNP0ZBl1vzvZPSQfS5e_R3-H-"
	secretKey := "xT9htZWcqGT4cJWBRyWW3WjJoxry3IrDVhYbe76C"
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}

	mac := qbox.NewMac(accessKey, secretKey)
	localfile := localPath
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	cfg.Zone = &storage.ZoneHuanan
	cfg.UseHTTPS = true
	cfg.UseCdnDomains = false

	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{}

	err = formUploader.PutFile(context.Background(), &ret, upToken, key, localfile, &putExtra)
	if err != nil {
		log.Println(err)
	}
	log.Println("hash: ", ret.Hash, "key: ", ret.Key)
	return
}
