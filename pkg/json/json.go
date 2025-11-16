package json

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"maps"
	"net/http"

	"github.com/Mockird31/avito_tech/internal/entity"
)

const (
	MaxBytes = 1024 * 1024
)

var (
	ErrMultipleJSONValues = errors.New("body must only contain a single JSON value")
)

func ReadJSON(w http.ResponseWriter, r *http.Request, v interface{}) error {
	maxBytes := int64(MaxBytes)
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(v); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return ErrMultipleJSONValues
	}

	return nil
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}, headers http.Header) {
	var jsonData []byte
	var err error

	jsonData, err = json.Marshal(data)
	if err != nil {
		log.Printf("json.WriteJSON: %v", err)
	}

	maps.Copy(w.Header(), headers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(jsonData)
	if err != nil {
		log.Printf("json.WriteJSON: %v", err)
	}
}

func WriteErrorJson(w http.ResponseWriter, status int, errorMessage string) {
	errorResponse := &entity.ErrorResponse{
		Code:    status,
		Message: errorMessage,
	}
	WriteJSON(w, status, errorResponse, nil)
}
