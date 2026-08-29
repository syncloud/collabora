package backend

type Claims struct {
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	Name              string `json:"name"`
	Subject           string `json:"sub"`
}

func (c Claims) User() string {
	if c.PreferredUsername != "" {
		return c.PreferredUsername
	}
	return c.Subject
}
