window.onload = function () {
    document.addEventListener("drop", function (e) {
        e.preventDefault()
    })

    document.addEventListener("dragleave", function (e) {
        e.preventDefault()
    })

    document.addEventListener("dragenter", function (e) {
        e.preventDefault()
    })

    document.addEventListener("dragover", function (e) {
        e.preventDefault()
    })

    function uploadFile(file) {
        return new Promise((resove, reject) => {
            let url
            if (file.name.endsWith(".apk")) {
                url = "http://localhost:8080/uploadApk"
            } else if (file.name.endsWith(".ipa")) {
                url = "http://localhost:8080/uploadIpa"
            } else {
                resove("上传文件格式不符")
            }
            let formData = new FormData()
            formData.append("file", file)

            fetch(url, {
                method: "POST",
                body: formData
            })
                .then((response) => response.json())
                .then(data => {
                    resove(data)
                })
                .catch((e) => {
                    reject(e)
                })
        })
    }

    function generInfo(packageSelector, qrcodeSelector, versionSelector, nameSelector, data) {
        document.querySelector(packageSelector)
            .style.display = "flex"
        while (document.querySelector(`${packageSelector} ${qrcodeSelector}`).hasChildNodes()) {
            document.querySelector(`${packageSelector} ${qrcodeSelector}`).childNodes.forEach( n => {
                document.querySelector(`${packageSelector} ${qrcodeSelector}`).removeChild(n)
            })
        }
        document.querySelector(`${packageSelector} ${nameSelector}`).innerText =  data.name
        document.querySelector(`${packageSelector} ${versionSelector}`).innerText = `version: ${data.version} (build ${data.build})`
        new QRCode(document.querySelector(`${packageSelector} ${qrcodeSelector}`), {
            text: data.url,
            width: 150,
            height: 150
        })
    }

    let areaBox = document.getElementById("drop-box")

    areaBox.addEventListener("drop", function (e) {
        var fileList = e.dataTransfer.files;
        if (fileList.length == 0) {
            return false
        } else if (fileList.length == 1) {
            let f = fileList[0]
            uploadFile(f)
                .then(response => {
                    if (response.data && response.data.platform == 1) { // iOS安装包
                        generInfo(".ios-package",
                            ".package-qrcode",
                            ".package-version",
                            ".package-name",
                            response.data)
                    } else if (response.data && response.data.platform == 2){ // 安卓安装包
                        generInfo(".android-package",
                            ".package-qrcode",
                            ".package-version",
                            ".package-name",
                            response.data)
                    }
                })
                .catch(e => {
                    alert(e)
                })
        } else {
            alert("只支持单个文件的上传")
        }
    })

}

