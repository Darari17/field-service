package gcs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage" // library resmi Google Cloud Storage
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option" // ini digunakan untuk mengatur opsi klien Google Cloud Storage
)

// struct ini mempresentasikan isi file service-acount-key.json yang biasa diunduh dari Google Cloud Console.
// field nya harus cocok dengan JSON di file asli agar bisa digunakan untuk autentikasi.
type ServiceAccountKeyJSON struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

// struct ini untuk menyimpan informasi klien GCS seperti kredensial dan nama bucket.
type GCSClient struct {
	ServiceAccountKeyJSON ServiceAccountKeyJSON
	BucketName            string
}

type IGCSClient interface {
	UploadFile(context.Context, string, []byte) (string, error)
}

func NewGCSClient(serviceAccountKeyJSON ServiceAccountKeyJSON, bucketName string) IGCSClient {
	return &GCSClient{
		ServiceAccountKeyJSON: serviceAccountKeyJSON,
		BucketName:            bucketName,
	}
}

func (g *GCSClient) createClient(ctx context.Context) (*storage.Client, error) {
	// membuat buffer kosong untuk menampung hasil encoding JSON
	reqBodyBytes := new(bytes.Buffer)

	// Encode struct ServiceAccountKeyJSON (berisi key, project ID, dsb) menjadi JSON
	// Contohnya akan seperti isi file service-account.json
	err := json.NewEncoder(reqBodyBytes).Encode(g.ServiceAccountKeyJSON)
	if err != nil {
		logrus.Errorf("failed to encode service account key json: %v", err)
		return nil, err
	}

	// Mengambil hasil encoding JSON dari buffer dalam bentuk []byte
	jsonByte := reqBodyBytes.Bytes()

	// Membuat client GCS dengan kredensial JSON di-memory (tanpa file)
	// option.WithCredentialsJSON(jsonByte) adalah cara resmi dari Google SDK
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(jsonByte))
	if err != nil {
		logrus.Errorf("failed to create client: %v", err)
		return nil, err
	}

	// Jika berhasil, kembalikan pointer *storage.Client
	return client, nil
}

func (g *GCSClient) UploadFile(ctx context.Context, filename string, data []byte) (string, error) {
	// Tentukan tipe konten default (jika tidak diketahui)
	// "application/octet-stream" = file biner umum
	var (
		contentType      = "application/octet-stream"
		timeoutInSeconds = 60
	)

	// Buat client GCS menggunakan kredensial di struct GCSClient
	client, err := g.createClient(ctx)
	if err != nil {
		logrus.Errorf("failed to create client: %v", err)
		return "", err
	}

	// Pastikan client ditutup setelah selesai (agar koneksi dilepaskan)
	defer func(client *storage.Client) {
		err := client.Close()
		if err != nil {
			logrus.Errorf("failed to close client: %v", err)
			return
		}
	}(client)

	// Membuat context baru dengan timeout (otomatis batal jika > 60 detik)
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutInSeconds)*time.Second)
	defer cancel() // pastikan context dibatalkan setelah selesai

	// Ambil referensi bucket berdasarkan nama yang sudah diset di GCSClient
	bucket := client.Bucket(g.BucketName)

	// Ambil referensi objek (file) di dalam bucket berdasarkan nama file yang diberikan
	object := bucket.Object(filename)

	// Bungkus data []byte ke dalam buffer agar bisa ditulis ke writer
	buffer := bytes.NewBuffer(data)

	// Buat writer untuk menulis data ke object GCS
	writer := object.NewWriter(ctx)

	// Set ChunkSize = 0 artinya upload akan dilakukan sekaligus (bukan bertahap)
	writer.ChunkSize = 0

	// Salin data dari buffer ke GCS writer (proses upload sebenarnya)
	_, err = io.Copy(writer, buffer)
	if err != nil {
		logrus.Errorf("failed to copy: %v", err)
		return "", err
	}

	// Tutup writer -> sangat penting karena menandakan upload selesai
	err = writer.Close()
	if err != nil {
		logrus.Errorf("failed to close: %v", err)
		return "", err
	}

	// Update metadata object (misalnya Content-Type)
	// tanpa ini file tetap bisa diakses tapi tidak punya tipe konten yang benar
	_, err = object.Update(ctx, storage.ObjectAttrsToUpdate{ContentType: contentType})
	if err != nil {
		logrus.Errorf("failed to update: %v", err)
		return "", err
	}

	// Buat URL publik untuk mengakses file (jika bucket publik)
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.BucketName, filename)

	// Kembalikan URL file hasil upload
	return url, nil
}
