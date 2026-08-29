package backend

type Config struct {
	Socket      string
	WopiListen  string
	WopiBaseURL string
	BaseURL     string
	Secret      []byte
}
