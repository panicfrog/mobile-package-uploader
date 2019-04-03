package config

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"qiniupkg.com/x/errors.v7"
)

type application struct {
	MaxMultipartMemory int
}

type aliyun struct {
	AliyunBucket string
	AccessKeyId string
	AccessKeySecret string
	EndPoint string
}

type filesPath struct {
	ApiPath string
	PlistsPath string
	TemPath string
}

type configuration struct {
	Application application
	Aliyun aliyun
	FilesPath filesPath
}

var Config configuration

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	check(err)

	viper.AddConfigPath(dir)
	err = viper.ReadInConfig()
	check(err)

	var ok bool

	config, ok :=  viper.Get("config").(map[string]interface{})
	checkOk(ok, "读取config出错")

	fpMap, ok := config["files_path"].(map[string]interface{})
	checkOk(ok, "读取filespath出错")

	aMap, ok := config["aliyun"].(map[string]interface{})
	checkOk(ok, "读取aliyun出错")

	appMap, ok := config["application"].(map[string]interface{})
	checkOk(ok, "读取application出错")

	fpApiPath, ok := fpMap["api_path"].(string)
	checkOk(ok, "api_path error")
	fpPlistPath, ok := fpMap["plists_path"].(string)
	checkOk(ok, "plists_path error")
	fpTemPath, ok := fpMap["tem_path"].(string)
	checkOk(ok, "tem_path error")

	fp := filesPath{
		ApiPath: fpApiPath,
		PlistsPath: fpPlistPath,
		TemPath: fpTemPath,
	}

	aliBucket, ok := aMap["aliyun_bucket"].(string)
	checkOk(ok, "aliyun_bucket error")
	aliAccessKeyId ,ok := aMap["access_key_id"].(string)
	checkOk(ok, "access_key_id error")
	aliAccessKeySecret, ok := aMap["access_key_secret"].(string)
	checkOk(ok, "access_key_secret error")
	aliEndPoint, ok := aMap["end_point"].(string)
	checkOk(ok, "end_point error")

	ali := aliyun{
		AliyunBucket: aliBucket,
		AccessKeyId: aliAccessKeyId,
		AccessKeySecret: aliAccessKeySecret,
		EndPoint: aliEndPoint,
	}

	appMaxMultipartMemory, ok := appMap["max_multipart_memory"].(int)
	checkOk(ok, "max_multipart_memory error")

	app := application{
		MaxMultipartMemory: appMaxMultipartMemory,
	}

	Config = configuration{
		Application: app,
		Aliyun: ali,
		FilesPath: fp,
	}

}

func check(err error){
	if err != nil {
		panic(err)
	}
}

func checkOk(ok bool, message string)  {
	if !ok {
		panic(errors.New(message))
	}
}