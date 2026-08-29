package backend

import "net/http"

type SessionAPI struct {
	auth *Auth
}

func NewSessionAPI(auth *Auth) *SessionAPI {
	return &SessionAPI{auth: auth}
}

func (a *SessionAPI) Health(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]bool{"auth": a.auth.Ready()})
}

func (a *SessionAPI) Current(writer http.ResponseWriter, request *http.Request) {
	session, _ := SessionFrom(request.Context())
	writeJSON(writer, http.StatusOK, map[string]string{
		"username": session.Username,
		"name":     session.Name,
		"email":    session.Email,
	})
}
