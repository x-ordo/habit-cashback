package httpx

import (
	"encoding/json"
	"net/http"
)

type ErrorResp struct {
	Error   string `json:"error"`
	Details any    `json:"details,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func BadRequest(w http.ResponseWriter, msg string, details any) {
	WriteJSON(w, http.StatusBadRequest, ErrorResp{Error: msg, Details: details})
}

func Unauthorized(w http.ResponseWriter, msg string) {
	WriteJSON(w, http.StatusUnauthorized, ErrorResp{Error: msg})
}

func ServerError(w http.ResponseWriter, msg string) {
	WriteJSON(w, http.StatusInternalServerError, ErrorResp{Error: msg})
}
