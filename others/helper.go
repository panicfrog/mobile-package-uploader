package others

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
)

func ConvPngToBase64String(img *image.Image) (s string, e error)  {
	buf := new(bytes.Buffer)
	e = png.Encode(buf, *img)
	if e != nil {
		return
	}
	//log.Println("buf:", buf.Bytes())
	s = base64.StdEncoding.EncodeToString(buf.Bytes())
	return
}