package api

type SucccStruct struct {
	Sc ApiStatus `json:"sc"`
	Message string `json:"message"`
	Data interface{} `json:"data"`
}

type UpdataResponse struct {
	Platform Platform `json:"platform"`
	Url string `json:"url"`
}