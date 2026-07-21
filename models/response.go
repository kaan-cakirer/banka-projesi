package models

// APIResponse: Tüm HTTP yanıtlarının standart formatıdır.
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"` // any (interface{}): İstediğimiz her veriyi taşıyabilir
}
