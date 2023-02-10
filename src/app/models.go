package app

type Document struct {
	Title    string `json:"title" xml:"title" index:"title" binding:"required" `
	Url      string `json:"url" xml:"url" binding:"required"`
	Abstract string `json:"abstract" xml:"abstract" index:"abstract" binding:"required"`
}
