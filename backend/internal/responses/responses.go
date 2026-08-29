package responses

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type List struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Body{Success: true, Message: message, Data: data})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Body{Success: true, Message: message, Data: data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Error(c *gin.Context, status int, message string, errs interface{}) {
	c.JSON(status, Body{Success: false, Message: message, Errors: errs})
}

func CreatedNoDataGuard(data interface{}) interface{} {
	if data == nil {
		return gin.H{}
	}
	return data
}
