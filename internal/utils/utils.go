package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type Envelope map[string]interface{}

func WriteJSON(w http.ResponseWriter, status int, data Envelope) error {
	js, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}
	js = append(js, '\n')
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func ReadIDParam(r *http.Request) (int64, error) {
	path := r.URL.Path
	pathParts := strings.Split(path, "/")

	// Find the ID part (usually the last non-empty segment)
	var idParam string
	for i := len(pathParts) - 1; i >= 0; i-- {
		if pathParts[i] != "" {
			idParam = pathParts[i]
			break
		}
	}

	if idParam == "" {
		return 0, errors.New("invalid id param")
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return 0, errors.New("invalid id param")
	}

	return id, nil
}
