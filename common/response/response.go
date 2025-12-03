package response

import (
	"field-service/constants"
	errConstnt "field-service/constants/error"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response merepresentasikan struktur standar response JSON yang dikirim ke client.
// Token bersifat opsional dan hanya muncul jika nilainya tidak nil.
type Response struct {
	Status  string      `json:"status"`
	Message any         `json:"message"`
	Data    interface{} `json:"data"`
	Token   *string     `json:"token,omitempty"`
}

// ParamHTTPResp adalah parameter input untuk fungsi HttpResponse.
// Tujuannya agar lebih mudah mengirim response tanpa perlu menulis banyak argumen.
type ParamHTTPResp struct {
	Code    int
	Err     error
	Message *string
	Gin     *gin.Context
	Data    interface{}
	Token   *string
}

// HttpResponse digunakan untuk mengirim response JSON standar ke client.
// Jika tidak ada error (param.Err == nil), maka akan mengirim response sukses.
// Jika ada error, akan mengirim response dengan status "error".
func HttpResponse(param ParamHTTPResp) {
	if param.Err == nil {
		param.Gin.JSON(param.Code, Response{
			Status:  constants.Success,
			Message: http.StatusText(http.StatusOK),
			Data:    param.Data,
			Token:   param.Token,
		})
		return
	}

	message := errConstnt.ErrInternalServerError.Error()
	if param.Message != nil {
		message = *param.Message
	} else if param.Err != nil {
		if errConstnt.ErrMapping(param.Err) {
			message = param.Err.Error()
		}
	}

	param.Gin.JSON(param.Code, Response{
		Status:  constants.Error,
		Message: message,
		Data:    param.Data,
	})
	return
}
