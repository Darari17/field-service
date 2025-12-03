package constants

import "net/textproto"

// textproto.CanonicalMIMEHeaderKey(...) ==> digunakan untuk Membuat format header jadi baku (canonical form) — yaitu huruf pertama setiap kata kapital. contoh: "x-api-key" menjadi "X-Api-Key"

// keempat variable ini digunakan untuk keamanan jika ada frontend atau service lain yang ingin mengakses field-service harus ada 4 header dibawah ini.

var (
	XServiceName  = textproto.CanonicalMIMEHeaderKey("x-service-name")
	XApiKey       = textproto.CanonicalMIMEHeaderKey("x-api-key")
	XRequestAt    = textproto.CanonicalMIMEHeaderKey("x-request-at")
	Authorization = textproto.CanonicalMIMEHeaderKey("authorization")
)
