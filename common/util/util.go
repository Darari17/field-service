package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

type PaginationParam struct {
	Count int64       `json:"count"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
	Data  interface{} `json:"data"`
}

type PaginationResult struct {
	TotalPage int         `json:"totalPage"`
	TotalData int64       `json:"totalData"`
	NextPage  *int        `json:"nextPage"`
	PrevPage  *int        `json:"prevPage"`
	Page      int         `json:"page"`
	Limit     int         `json:"limit"`
	Data      interface{} `json:"data"`
}

func GeneratePagination(params PaginationParam) PaginationResult {
	totalPage := int(math.Ceil(float64(params.Count) / float64(params.Limit)))

	var (
		nextPage     int
		previousPage int
	)

	if params.Page < totalPage {
		nextPage = params.Page + 1
	}

	if params.Page > 1 {
		previousPage = params.Page - 1
	}

	result := PaginationResult{
		TotalPage: totalPage,
		TotalData: params.Count,
		NextPage:  &nextPage,
		PrevPage:  &previousPage,
		Page:      params.Page,
		Limit:     params.Limit,
		Data:      params.Data,
	}
	return result
}

func GenerateSHA256(inputString string) string {
	hash := sha256.New()
	hash.Write([]byte(inputString))
	hashBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	return hashString
}

func RupiahFormat(amount *float64) string {
	stringValue := "0"
	if amount != nil {
		humanizeValue := humanize.CommafWithDigits(*amount, 0)
		stringValue = strings.ReplaceAll(humanizeValue, ",", ".")
	}
	return fmt.Sprintf("Rp. %s", stringValue)
}

// BindFromJSON membaca file konfigurasi ber-format JSON
// dan melakukan unmarshal ke dalam struct tujuan (dest).
// Parameter:
// - dest: pointer ke struct tujuan tempat data disalin
// - filename: nama file konfigurasi (tanpa ekstensi .json)
// - path: lokasi direktori tempat file JSON berada
func BindFromJSON(dest any, filename, path string) error {
	v := viper.New() // Membuat instance baru dari Viper

	v.SetConfigType("json")   // Menentukan format file konfigurasi
	v.AddConfigPath(path)     // Menentukan direktori tempat file JSON berada
	v.SetConfigName(filename) // Menentukan nama file konfigurasi

	// Membaca file konfigurasi dari path yang ditentukan
	err := v.ReadInConfig()
	if err != nil {
		return err
	}

	// Melakukan unmarshal isi file ke dalam struct tujuan
	err = v.Unmarshal(&dest)
	if err != nil {
		logrus.Errorf("failed to unmarshal: %v", err)
		return err
	}

	return nil
}

// SetEnvFromConsulKV mengambil seluruh key-value dari Viper (yang berisi data dari Consul)
// kemudian menyimpannya ke dalam environment variable sistem.
func SetEnvFromConsulKV(v *viper.Viper) error {
	env := make(map[string]any)

	// Unmarshal semua key-value yang ada di Viper ke dalam map
	err := v.Unmarshal(&env)
	if err != nil {
		logrus.Errorf("failed to unmarshal: %v", err)
		return err
	}

	// Iterasi semua key-value dalam map
	for k, v := range env {
		var (
			valOf = reflect.ValueOf(v) // Mengambil refleksi tipe data value
			val   string               // Variabel untuk menyimpan nilai string akhir
		)

		// Cek tipe data dari value dan konversi ke string sesuai tipe-nya
		switch valOf.Kind() {
		case reflect.String:
			val = valOf.String()
		case reflect.Int:
			val = strconv.Itoa(int(valOf.Int()))
		case reflect.Uint:
			val = strconv.Itoa(int(valOf.Uint()))
		case reflect.Float32:
			val = strconv.Itoa(int(valOf.Float()))
		case reflect.Float64:
			val = strconv.Itoa(int(valOf.Float()))
		case reflect.Bool:
			val = strconv.FormatBool(bool(valOf.Bool()))
		}

		// Menyimpan nilai ke dalam environment variable
		err = os.Setenv(k, val)
		if err != nil {
			logrus.Errorf("failed to set env: %v", err)
			return err
		}
	}

	return nil
}

// BindFromConsul digunakan untuk membaca konfigurasi dari Consul KV Store,
// melakukan unmarshal hasilnya ke dalam struct tujuan, dan menyimpannya juga
// sebagai environment variable.
func BindFromConsul(dest any, endPoint, path string) error {
	v := viper.New() // Membuat instance baru dari Viper

	v.SetConfigType("json") // Menentukan format data remote (JSON)

	// Menambahkan remote provider Consul dengan endpoint dan path tertentu
	err := v.AddRemoteProvider("consul", endPoint, path)
	if err != nil {
		logrus.Errorf("failed to remote add provider: %v", err)
		return err
	}

	// Membaca konfigurasi dari Consul secara remote
	err = v.ReadRemoteConfig()
	if err != nil {
		logrus.Errorf("failed to read remote config: %v", err)
		return err
	}

	// Melakukan unmarshal ke struct tujuan
	err = v.Unmarshal(&dest)
	if err != nil {
		logrus.Errorf("failed to unmarshal: %v", err)
		return err
	}

	// Mengatur nilai-nilai dari Consul KV menjadi environment variable
	err = SetEnvFromConsulKV(v)
	if err != nil {
		logrus.Errorf("failed to set env from consul kv: %v", err)
		return err
	}

	return nil
}
