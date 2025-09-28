package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator"
)

type errorMessage struct {
	Message string `json:"message"`
}

func makeErrorResponse(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errResponse := &errorMessage{Message: err.Error()}
	if errE := json.NewEncoder(w).Encode(errResponse); errE != nil {
		fmt.Println(errE)
		return
	}
}

func formatValidationErrors(errs validator.ValidationErrors) string {
	var messages []string
	for _, e := range errs {
		var msg string
		switch e.Tag() {
		case "gt":
			msg = fmt.Sprintf("поле '%s' должно быть больше %s", e.Field(), e.Param())
		case "required":
			msg = fmt.Sprintf("поле '%s' является обязательным", e.Field())
		default:
			msg = fmt.Sprintf("поле '%s' не предусмотрено в теле запроса", e.Field())
		}
		messages = append(messages, msg)
	}
	return strings.Join(messages, ", ")
}
