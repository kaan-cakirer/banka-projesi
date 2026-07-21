package utils

import (
	"banka-projesi/models"
	"encoding/json"
	"net/http"
)

// JSONResponse: Tüm handler'ların çıktı vermek için kullanacağı merkezi yardımcı fonksiyondur.
func JSONResponse(w http.ResponseWriter, statusCode int, success bool, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := models.APIResponse{
		Success: success,
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(resp)
}
