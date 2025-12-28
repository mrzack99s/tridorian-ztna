package common

import (
	"encoding/json"
	"net/http"
)

type JSONResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, success bool, data interface{}, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(JSONResponse{
		Success: success,
		Data:    data,
		Error:   errMsg,
	})
}

func Error(w http.ResponseWriter, status int, errMsg string) {
	WriteJSON(w, status, false, nil, errMsg)
}

func Success(w http.ResponseWriter, status int, data interface{}) {
	WriteJSON(w, status, true, data, "")
}
