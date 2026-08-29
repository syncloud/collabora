package backend

type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	BaseURL      string
	AuthSocket   string
}
