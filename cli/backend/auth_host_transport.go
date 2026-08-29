package backend

import "net/http"

type authHostTransport struct {
	host      string
	viaSocket http.RoundTripper
	fallback  http.RoundTripper
}

func (t *authHostTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Host != t.host {
		return t.fallback.RoundTrip(request)
	}
	routed := request.Clone(request.Context())
	routed.URL.Scheme = "http"
	routed.Header.Set("X-Forwarded-Proto", "https")
	routed.Header.Set("X-Forwarded-Host", t.host)
	return t.viaSocket.RoundTrip(routed)
}
