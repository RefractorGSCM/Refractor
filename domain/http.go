package domain

type Response struct {
	Success bool              `json:"success"`
	Message string            `json:"message,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
	Payload interface{}       `json:"payload,omitempty"`
}
