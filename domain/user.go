package domain

import kratos "github.com/ory/kratos-client-go"

// User represents a stored user of the system.
type User struct {
	Traits *Traits
	*kratos.Identity
}
