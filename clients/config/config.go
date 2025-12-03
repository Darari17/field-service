package config

import "github.com/parnurzeal/gorequest"

// Tujuan kode ini adalah untuk mengatur konfigurasi klien HTTP yang dapat digunakan untuk melakukan permintaan ke server tertentu.

// ClientConfig menyimpan konfigurasi klien HTTP seperti baseURL dan signatureKey.
type ClientConfig struct {
	client       *gorequest.SuperAgent
	baseURL      string
	signatureKey string
}

// IClientConfig mendefinisikan antarmuka untuk konfigurasi klien HTTP.
type IClientConfig interface {
	Client() *gorequest.SuperAgent
	BaseURL() string
	SignatureKey() string
}

// Option adalah fungsi yang mengonfigurasi ClientConfig.
type Option func(*ClientConfig)

// NewClientConfig membuat instance baru dari ClientConfig dengan opsi yang diberikan.
func NewClientConfig(options ...Option) IClientConfig {
	// Inisialisasi klien HTTP dengan header default
	clientConfig := &ClientConfig{
		client: gorequest.New().
			Set("Content-Type", "application/json").
			Set("Accept", "application/json"),
	}

	// Terapkan setiap opsi konfigurasi
	for _, option := range options {
		option(clientConfig)
	}

	// Kembalikan instance ClientConfig yang telah dikonfigurasi
	return clientConfig
}

// getters untuk Client HTTP
func (c *ClientConfig) Client() *gorequest.SuperAgent {
	return c.client
}

// getters untuk baseURL
func (c *ClientConfig) BaseURL() string {
	return c.baseURL
}

// getters untuk signatureKey
func (c *ClientConfig) SignatureKey() string {
	return c.signatureKey
}

// WithBaseURL mengatur baseURL untuk ClientConfig.
func WithBaseURL(baseURL string) Option {
	return func(c *ClientConfig) {
		c.baseURL = baseURL
	}
}

// WithSignatureKey mengatur signatureKey untuk ClientConfig.
func WithSignatureKey(signatureKey string) Option {
	return func(c *ClientConfig) {
		c.signatureKey = signatureKey
	}
}
