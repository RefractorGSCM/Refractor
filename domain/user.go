package domain

// User represents a stored user of the system.
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Group    string `json:"group"`
}
