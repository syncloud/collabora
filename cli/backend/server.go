package backend

import (
	"context"
	"net"
	"net/http"
	"os"

	"go.uber.org/zap"
)

type Server struct {
	config  Config
	files   *FileStore
	auth    *Auth
	session *SessionAPI
	api     *FilesAPI
	editor  *EditorAPI
	wopi    *WopiHost
	logger  *zap.Logger
}

func NewServer(
	config Config,
	files *FileStore,
	auth *Auth,
	session *SessionAPI,
	api *FilesAPI,
	editor *EditorAPI,
	wopi *WopiHost,
	logger *zap.Logger,
) *Server {
	return &Server{
		config:  config,
		files:   files,
		auth:    auth,
		session: session,
		api:     api,
		editor:  editor,
		wopi:    wopi,
		logger:  logger,
	}
}

func (s *Server) Run(ctx context.Context) error {
	err := s.files.Init()
	if err != nil {
		return err
	}

	s.auth.InitWithRetry(ctx, discoveryPeriod)

	errors := make(chan error, 2)
	go func() { errors <- s.serveAPI() }()
	go func() { errors <- s.serveWopi() }()
	return <-errors
}

func (s *Server) serveAPI() error {
	if err := os.Remove(s.config.Socket); err != nil && !os.IsNotExist(err) {
		return err
	}
	listener, err := net.Listen("unix", s.config.Socket)
	if err != nil {
		return err
	}
	if err := os.Chmod(s.config.Socket, 0o660); err != nil {
		return err
	}
	s.logger.Info("api listening", zap.String("socket", s.config.Socket))
	return http.Serve(listener, s.apiRoutes())
}

func (s *Server) serveWopi() error {
	listener, err := net.Listen("tcp", s.config.WopiListen)
	if err != nil {
		return err
	}
	s.logger.Info("wopi host listening", zap.String("address", s.config.WopiListen))
	return http.Serve(listener, s.wopi.Routes())
}

func (s *Server) apiRoutes() http.Handler {
	authorised := http.NewServeMux()
	authorised.HandleFunc("GET /api/session", s.session.Current)
	authorised.HandleFunc("GET /api/editor", s.editor.Open)
	s.api.Routes(authorised)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /oidc/start", s.auth.StartHandler)
	mux.HandleFunc("GET /oidc/callback", s.auth.CallbackHandler)
	mux.HandleFunc("GET /oidc/logout", s.auth.LogoutHandler)
	mux.HandleFunc("GET /api/health", s.session.Health)
	mux.Handle("/api/", s.auth.APIMiddleware(authorised))
	return mux
}
