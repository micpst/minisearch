package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/micpst/full-text-search-engine/db"
)

type DocumentBody struct {
	Content string `json:"content" binding:"required"`
}

type SearchQueryParams struct {
	Q string `form:"q" binding:"required"`
}

func CreateDocument(c *gin.Context) {
	var documentBody DocumentBody
	if err := c.BindJSON(&documentBody); err != nil {
		return
	}

	id := db.AddDocument(documentBody.Content)
	db.IndexDocument(documentBody.Content)

	c.JSON(http.StatusCreated, gin.H{
		id: id,
	})
}

func UpdateDocument(c *gin.Context) {
	//id := c.Param("id")
	//db.
	//for _, a := range albums {
	//	if a.ID == id {
	//		c.JSON(http.StatusOK, a)
	//		return
	//	}
	//}
	//c.Inden	tedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
	//var data UpdateDataV1
	//err := c.Bind(&data)
	//if err != nil {
	//	fmt.Println(err)
	//	c.JSON(500, gin.H{
	//		"error": "an error occurred",
	//	})
	//	return
	//}
	//
	//updateError := UpdateDocument(data.ID, data.Content)
	//if updateError != nil {
	//	fmt.Println(err)
	//	c.JSON(500, gin.H{
	//		"error": "an error occurred",
	//	})
	//	return
	//}
	//
	//c.JSON(200, gin.H{
	//	"result": fmt.Sprintf("Document %s updated successfully", data.ID),
	//})
}

func DeleteDocument(c *gin.Context) {
	//var data DeleteDataV1
	//err := c.Bind(&data)
	//if err != nil {
	//	fmt.Println(err)
	//	c.JSON(500, gin.H{
	//		"error": "an error occurred",
	//	})
	//	return
	//}
	//
	//deleteError := DeleteDocument(data.ID)
	//if deleteError != nil {
	//	fmt.Println(err)
	//	c.JSON(500, gin.H{
	//		"error": "an error occurred",
	//	})
	//	return
	//}
	//
	//c.JSON(200, gin.H{
	//	"result": fmt.Sprintf("Document %s deleted successfully", data.ID),
	//})
}

func SearchDocument(c *gin.Context) {
	//var query SearchDataV1
	//err := c.Bind(&query)
	//if err != nil {
	//	fmt.Println(err)
	//	c.JSON(500, gin.H{
	//		"error": "an error occurred",
	//	})
	//	return
	//}
	//
	//result := Search(query.Q)
	//
	//c.JSON(200, gin.H{
	//	"docs": result,
	//})
}
