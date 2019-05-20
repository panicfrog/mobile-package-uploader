package api

type SucccStruct struct {
	Sc ApiStatus 		`json:"sc"`
	Message string 		`json:"message"`
	Data interface{} 	`json:"data"`
}

type UpdataResponse struct {
	Platform Platform 	`json:"platform"`
	Url string 			`json:"url"`
	PackageId string 	`json:"package_id"`
	Version string 		`json:"version"`
	Build string 		`json:"build"`
	Name string			`json:"name"`
}