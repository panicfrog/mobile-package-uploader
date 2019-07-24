package main

import (
	"encoding/base64"
	"errors"
	"github.com/gin-gonic/gin"
	"go_ipa_uploader/ipa"
	"go_ipa_uploader/others"
	"log"
	"strconv"

	//"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"go_ipa_uploader/config"
	"howett.net/plist"
	"io"
	"io/ioutil"
	_ "net/http"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/mholt/archiver"
	"github.com/shogo82148/androidbinary/apk"
	"go_ipa_uploader/api"
	neturl "net/url"
	"os"
)

func route(r *gin.Engine) {
	r.GET("/hello", hello)
	r.POST("/uploadIpa", uploadIpa)
	r.POST("/uploadApk", uploadApk)
	r.Static("/web","./www" )
}

func hello(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "world",
	})
}

// upload ipa
func uploadIpa(c *gin.Context) {
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
	_, err := io.Copy(out, w)
	if err != nil {
		log.Println(err)
		panic(err)
	}

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

	plistPath, plistDir,infoPlist, _, genErr := generatePlist(current, downloadPath)
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
	response := api.UpdataResponse{
		api.IOS,
		actionURL ,
		infoPlist.BundleId,
		infoPlist.BundleShortVersion,
		infoPlist.BundleVersion,
		infoPlist.BundleName,
	}
	api.SendSuccess("上传成功", response, c)
	_ = os.Remove(ipa_path)
	_ = os.RemoveAll(plistDir)
	_ = os.RemoveAll(tem)
}

// update apk
func uploadApk(c *gin.Context) {
	file, formErro := c.FormFile("file")

	if formErro != nil {
		api.SendFail(formErro.Error(), c)
		return
	}

	filename := file.Filename

	if !strings.HasSuffix(filename, ".apk") {
		api.SendFail("invalid format for apk file", c)
		return
	}
	current := time.Now().Format("20060102150405")
	apk_path := config.Config.FilesPath.ApkPath + "/" + current + " " + filename
	out, erro := os.Create(apk_path)
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
	_, err := io.Copy(out, w)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	pkg, err := apk.OpenFile(apk_path)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	version := pkg.Manifest().VersionName
	build := strconv.Itoa(pkg.Manifest().VersionCode)
	name, _ := pkg.Label(nil)

	icon, _ := pkg.Icon(nil)
	var iconB64 string
	iconB64, cErr := others.ConvPngToBase64String(&icon)
	log.Println(iconB64)

	if cErr != nil {
		log.Println(cErr, iconB64)
	}

	packageName := pkg.PackageName()

	downloadPath, aliUploaderErr := aliyunOSSUpload(current + filename, apk_path)
	if aliUploaderErr != nil {
		api.SendFail(aliUploaderErr.Error(), c)
		return
	}
	response := api.UpdataResponse{
		api.Android,
		downloadPath ,
		packageName,
		version,
		build,
		name,
	}
	api.SendSuccess("上传成功", response, c)
	_ = os.Remove(apk_path)
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
		prefix := strings.Replace(config.Config.Aliyun.EndPoint, "http://", "https://" + bucketName + ".", -1)
		p := prefix + "/" + neturl.PathEscape(filename)
		//p := "https://" + bucketName + ".oss-cn-shenzhen.aliyuncs.com" + "/" + neturl.PathEscape(filename)
		downloadURL = p
		return
	}
	return
}

func generatePlist(current string, downloadURL string) (plistPath string, plistDir string, infoplist ipa.InfoPlist, icon string,genErr error) {
	tem := config.Config.FilesPath.TemPath + "/" + current
	payloadDir := tem + "/Payload"
	dirs, _ := ioutil.ReadDir(payloadDir)
	infoPlistPath := ""
	applicationDir := ""
	if len(dirs) > 0 && dirs[0].IsDir() {
		infoPlistPath = payloadDir + "/" + dirs[0].Name() + "/info.plist"
		applicationDir = payloadDir + "/" + dirs[0].Name()
	} else {
		genErr = errors.New("读取文件夹出错")
		return
	}

	// got icon
	subs, e := ioutil.ReadDir(applicationDir)
	if e == nil {
		var buf []byte
		for i := 0; i < len(subs); i ++  {
			p := applicationDir + "/" + subs[i].Name()
			if strings.HasPrefix(subs[i].Name(), "AppIcon") && !subs[i].IsDir() {} else {
				continue
			}

			b, iconErr := ioutil.ReadFile(p)
			if iconErr != nil {
				continue
			}

			if buf == nil {
				buf = b
			} else {
				if len(buf) < len(b) {
					buf = b
				}
			}
		}

		icon = base64.StdEncoding.EncodeToString(buf)
	}

	log.Println("ios icon：", icon)

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
	buildVersion := v["CFBundleVersion"].(string)
	displayName := v["CFBundleName"].(string)

	infoplist = ipa.InfoPlist{version, buildVersion, displayName, bundleId }

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
