package app

type Document struct {
	Title    string `json:"title" xml:"title" binding:"required" index:"true"`
	Url      string `json:"url" xml:"url" binding:"required"`
	Abstract string `json:"abstract" xml:"abstract" binding:"required" index:"true"`
}
