package helpers

import (
	"net/http"
	"time"
)

type Error struct {
	Code       string   `json:"code"`
	StatusCode int      `json:"statusCode"`
	Reason     string   `json:"reason"`
	DateTime   DateTime `json:"dateTime"`
}

type DateTime struct {
	Value    string `json:"value"`
	TimeZone string `json:"timeZone"`
}

func (e *Error) Error() string {
	return e.Reason
}

func ParseError(err error) (int, interface{}) {
	if parsedErr, ok := err.(*Error); ok {
		parsedErr.setDateTime()
		return parsedErr.StatusCode, parsedErr
	}

	return http.StatusInternalServerError, err
}

func (e *Error) setDateTime() {
	istLocation, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		istLocation = time.UTC
	}

	currentTime := time.Now().In(istLocation)
	e.DateTime = DateTime{
		Value:    currentTime.Format(time.RFC3339),
		TimeZone: "IST",
	}
}
