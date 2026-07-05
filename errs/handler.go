package errs

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrsResponse struct {
	Status bool   `json:"status"`
	Desc   string `json:"desc"`
}

func HandleError(c *gin.Context, err error) {
	switch e := err.(type) {
	case AppError:
		c.JSON(e.Code, ErrsResponse{Status: false, Desc: e.Message})
	case error:
		c.JSON(http.StatusInternalServerError, ErrsResponse{Status: false, Desc: e.Error()})
	}
}
