package backend

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	sessionCookie   = "collabora_session"
	stateCookie     = "collabora_oidc_state"
	platformCACert  = "/var/snap/platform/current/syncloud.ca.crt"
	discoveryPeriod = 30 * time.Second
)

type Auth struct {
	config OIDCConfig
	secret []byte
	logger *zap.Logger

	mutex    sync.RWMutex
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	oauth    *oauth2.Config
	client   *http.Client
}

func NewAuth(config OIDCConfig, secret []byte, logger *zap.Logger) *Auth {
	return &Auth{config: config, secret: secret, logger: logger}
}

func (a *Auth) Ready() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.provider != nil
}

func (a *Auth) InitWithRetry(ctx context.Context, period time.Duration) {
	go func() {
		for {
			if err := a.discover(ctx); err != nil {
				a.logger.Warn("oidc discovery failed, retrying", zap.Error(err))
			} else {
				a.logger.Info("oidc discovery succeeded")
				return
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(period):
			}
		}
	}()
}

func (a *Auth) discover(ctx context.Context) error {
	var lastErr error
	for _, client := range []*http.Client{a.socketClient(), platformClient()} {
		if client == nil {
			continue
		}
		provider, err := oidc.NewProvider(oidc.ClientContext(ctx, client), a.config.Issuer)
		if err != nil {
			lastErr = err
			continue
		}
		a.mutex.Lock()
		a.provider = provider
		a.verifier = provider.Verifier(&oidc.Config{ClientID: a.config.ClientID})
		a.oauth = &oauth2.Config{
			ClientID:     a.config.ClientID,
			ClientSecret: a.config.ClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  a.config.RedirectURL,
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "groups"},
		}
		a.client = client
		a.mutex.Unlock()
		return nil
	}
	if lastErr == nil {
		lastErr = errors.New("no usable transport to the auth service")
	}
	return lastErr
}

func (a *Auth) socketClient() *http.Client {
	if a.config.AuthSocket == "" {
		return nil
	}
	issuer, err := url.Parse(a.config.Issuer)
	if err != nil {
		return nil
	}
	dialer := &net.Dialer{}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, "unix", a.config.AuthSocket)
		},
	}
	return &http.Client{Transport: &authHostTransport{
		host:      issuer.Host,
		viaSocket: transport,
		fallback:  platformTransport(),
	}}
}

func platformTransport() *http.Transport {
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	if pem, err := os.ReadFile(platformCACert); err == nil {
		pool.AppendCertsFromPEM(pem)
	}
	return &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}}
}

func platformClient() *http.Client {
	return &http.Client{Transport: platformTransport()}
}

func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		session, found := a.session(request)
		if !found {
			a.startLogin(writer, request)
			return
		}
		next.ServeHTTP(writer, request.WithContext(WithSession(request.Context(), session)))
	})
}

func (a *Auth) APIMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		session, found := a.session(request)
		if !found {
			writeError(writer, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(writer, request.WithContext(WithSession(request.Context(), session)))
	})
}

func (a *Auth) session(request *http.Request) (Session, bool) {
	cookie, err := request.Cookie(sessionCookie)
	if err != nil {
		return Session{}, false
	}
	session, err := VerifySession(a.secret, cookie.Value)
	if err != nil {
		return Session{}, false
	}
	return session, true
}

func (a *Auth) startLogin(writer http.ResponseWriter, request *http.Request) {
	a.mutex.RLock()
	oauthConfig := a.oauth
	a.mutex.RUnlock()
	if oauthConfig == nil {
		http.Error(writer, "auth service is not reachable yet", http.StatusServiceUnavailable)
		return
	}

	state, err := RandomID(16)
	if err != nil {
		http.Error(writer, "state generation failed", http.StatusInternalServerError)
		return
	}
	nonce, err := RandomID(16)
	if err != nil {
		http.Error(writer, "nonce generation failed", http.StatusInternalServerError)
		return
	}
	verifier, err := RandomID(32)
	if err != nil {
		http.Error(writer, "verifier generation failed", http.StatusInternalServerError)
		return
	}

	returnTo := request.URL.RequestURI()
	if request.Method != http.MethodGet || strings.HasPrefix(returnTo, "/oidc/") {
		returnTo = "/"
	}
	encoded, err := encodeState(a.secret, stateBlob{
		Nonce:            nonce,
		Verifier:         verifier,
		Return:           returnTo,
		RegisteredClaims: jwt.RegisteredClaims{ID: state},
	})
	if err != nil {
		http.Error(writer, "state encoding failed", http.StatusInternalServerError)
		return
	}
	http.SetCookie(writer, &http.Cookie{
		Name:     stateCookie,
		Value:    encoded,
		Path:     "/oidc",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(stateTTL.Seconds()),
	})

	challenge := sha256.Sum256([]byte(verifier))
	http.Redirect(writer, request, oauthConfig.AuthCodeURL(state,
		oidc.Nonce(nonce),
		oauth2.SetAuthURLParam("code_challenge", base64.RawURLEncoding.EncodeToString(challenge[:])),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	), http.StatusFound)
}

func (a *Auth) StartHandler(writer http.ResponseWriter, request *http.Request) {
	a.startLogin(writer, request)
}

func (a *Auth) CallbackHandler(writer http.ResponseWriter, request *http.Request) {
	a.mutex.RLock()
	oauthConfig, verifier, client := a.oauth, a.verifier, a.client
	a.mutex.RUnlock()
	if oauthConfig == nil {
		http.Error(writer, "auth service is not reachable yet", http.StatusServiceUnavailable)
		return
	}

	cookie, err := request.Cookie(stateCookie)
	if err != nil {
		http.Error(writer, "missing state cookie", http.StatusBadRequest)
		return
	}
	blob, err := decodeState(a.secret, cookie.Value)
	if err != nil {
		http.Error(writer, "bad state cookie", http.StatusBadRequest)
		return
	}
	if request.URL.Query().Get("state") != blob.ID {
		http.Error(writer, "state mismatch", http.StatusBadRequest)
		return
	}
	code := request.URL.Query().Get("code")
	if code == "" {
		http.Error(writer, "missing code: "+request.URL.Query().Get("error"), http.StatusBadRequest)
		return
	}

	ctx := oidc.ClientContext(request.Context(), client)
	token, err := oauthConfig.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", blob.Verifier))
	if err != nil {
		a.logger.Warn("token exchange failed", zap.Error(err))
		http.Error(writer, "token exchange failed", http.StatusBadGateway)
		return
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(writer, "no id_token in response", http.StatusBadGateway)
		return
	}
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		a.logger.Warn("id_token verification failed", zap.Error(err))
		http.Error(writer, "id_token verification failed", http.StatusBadGateway)
		return
	}
	if idToken.Nonce != blob.Nonce {
		http.Error(writer, "nonce mismatch", http.StatusBadRequest)
		return
	}

	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		http.Error(writer, "claims parsing failed", http.StatusBadGateway)
		return
	}
	session, err := SignSession(a.secret, Session{
		Username: claims.User(),
		Email:    claims.Email,
		Name:     claims.Name,
	}, SessionTTL)
	if err != nil {
		http.Error(writer, "session creation failed", http.StatusInternalServerError)
		return
	}

	http.SetCookie(writer, &http.Cookie{
		Name:     sessionCookie,
		Value:    session,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(SessionTTL.Seconds()),
	})
	http.SetCookie(writer, &http.Cookie{
		Name:     stateCookie,
		Path:     "/oidc",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	returnTo := blob.Return
	if !strings.HasPrefix(returnTo, "/") {
		returnTo = "/"
	}
	http.Redirect(writer, request, returnTo, http.StatusFound)
}

func (a *Auth) LogoutHandler(writer http.ResponseWriter, request *http.Request) {
	http.SetCookie(writer, &http.Cookie{
		Name:     sessionCookie,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	target := strings.TrimRight(a.config.Issuer, "/") + "/logout?rd=" + url.QueryEscape(a.config.BaseURL+"/")
	http.Redirect(writer, request, target, http.StatusFound)
}
