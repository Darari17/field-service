package error

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

// common error ini berfungsi untuk validasi ketika ada request masuk
// Struct ini digunakan untuk membentuk response JSON ketika ada error validasi.
// Field berisi nama field yang salah, Message berisi pesan kesalahan.
type ValidationResponse struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message,omitempty"`
}

// Map untuk menampung pesan error custom berdasarkan tag validasi.
// Misalnya: "min": "%s must be at least %s characters"
var ErrValidator = map[string]string{}

// Fungsi ini digunakan untuk mengubah error hasil dari validator menjadi slice ValidationResponse.
// Tujuannya agar error validasi dapat dikirim ke client dengan format JSON yang mudah dibaca.
func ErrValidationResponse(err error) (validationResponse []ValidationResponse) {
	var fieldErrors validator.ValidationErrors

	// Mengecek apakah error yang diterima merupakan bagian dari validator.ValidationErrors
	if errors.As(err, &fieldErrors) {

		// Melakukan looping terhadap semua field yang error
		for _, err := range fieldErrors {
			switch err.Tag() {

			// Jika tag validasinya adalah "required"
			case "required":
				validationResponse = append(validationResponse, ValidationResponse{
					Field:   err.Field(),
					Message: fmt.Sprintf("%s is required", err.Field()),
				})

			// Jika tag validasinya adalah "email"
			case "email":
				validationResponse = append(validationResponse, ValidationResponse{
					Field:   err.Field(),
					Message: fmt.Sprintf("%s is not a valid email address", err.Field()),
				})

			// Jika tag validasinya bukan "required" atau "email"
			default:

				// Mengecek apakah ada pesan custom di ErrValidator untuk tag ini
				errValidator, ok := ErrValidator[err.Tag()]
				if ok {
					// Menghitung jumlah placeholder "%s" pada string pesan
					count := strings.Count(errValidator, "%s")

					// Jika hanya ada satu placeholder, gunakan satu parameter (nama field)
					if count == 1 {
						validationResponse = append(validationResponse, ValidationResponse{
							Field:   err.Field(),
							Message: fmt.Sprintf(errValidator, err.Field()),
						})
					} else {
						// Jika ada dua placeholder, gunakan field dan parameter tambahan (misalnya nilai minimum)
						validationResponse = append(validationResponse, ValidationResponse{
							Field:   err.Field(),
							Message: fmt.Sprintf(errValidator, err.Field(), err.Param()),
						})
					}
				} else {
					// Jika tidak ada pesan custom, gunakan pesan default
					validationResponse = append(validationResponse, ValidationResponse{
						Field:   err.Field(),
						Message: fmt.Sprintf("something wrong on %s; %s", err.Field(), err.Tag()),
					})
				}
			}
		}
	}

	// Mengembalikan semua hasil validasi dalam bentuk slice ValidationResponse
	return validationResponse
}

// Fungsi ini digunakan untuk mencatat error menggunakan logrus,
// kemudian mengembalikan kembali error tersebut agar bisa diproses lebih lanjut.
func WrapError(err error) error {
	logrus.Errorf("error: %v", err)
	return err
}
