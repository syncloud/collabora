package backend

import (
	"net/http"
	"os"
	"strconv"

	"go.uber.org/zap"
)

type WopiHost struct {
	files             *FileStore
	locks             *LockStore
	serve             *FilesAPI
	secret            []byte
	postMessageOrigin string
	logger            *zap.Logger
}

func NewWopiHost(files *FileStore, locks *LockStore, serve *FilesAPI, secret []byte, postMessageOrigin string, logger *zap.Logger) *WopiHost {
	return &WopiHost{
		files:             files,
		locks:             locks,
		serve:             serve,
		secret:            secret,
		postMessageOrigin: postMessageOrigin,
		logger:            logger,
	}
}

func (h *WopiHost) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /wopi/files/{id}", h.authorised(h.CheckFileInfo))
	mux.Handle("POST /wopi/files/{id}", h.authorised(h.Override))
	mux.Handle("GET /wopi/files/{id}/contents", h.authorised(h.GetFile))
	mux.Handle("POST /wopi/files/{id}/contents", h.authorised(h.PutFile))
	return mux
}

func (h *WopiHost) authorised(next func(http.ResponseWriter, *http.Request, wopiFile)) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		id := request.PathValue("id")
		token, err := VerifyAccessToken(h.secret, request.URL.Query().Get("access_token"), id)
		if err != nil {
			h.logger.Warn("wopi access token rejected", zap.String("file", id), zap.Error(err))
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		path, name, err := h.files.PathByID(id)
		if err != nil {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		next(writer, request, wopiFile{ID: id, Name: name, Path: path, User: token.User})
	})
}

func (h *WopiHost) CheckFileInfo(writer http.ResponseWriter, _ *http.Request, file wopiFile) {
	info, err := os.Stat(file.Path)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	user := file.User
	if user == "" {
		user = "user"
	}
	writeJSON(writer, http.StatusOK, CheckFileInfo{
		BaseFileName:               file.Name,
		Size:                       info.Size(),
		Version:                    version(info),
		OwnerId:                    user,
		UserId:                     user,
		UserFriendlyName:           user,
		UserCanWrite:               true,
		UserCanNotWriteRelative:    true,
		SupportsUpdate:             true,
		SupportsLocks:              true,
		SupportsGetLock:            true,
		SupportsExtendedLockLength: true,
		PostMessageOrigin:          h.postMessageOrigin,
		EnableOwnerTermination:     true,
	})
}

func (h *WopiHost) GetFile(writer http.ResponseWriter, _ *http.Request, file wopiFile) {
	h.serve.Serve(writer, file.Name)
}

func (h *WopiHost) PutFile(writer http.ResponseWriter, request *http.Request, file wopiFile) {
	held := h.locks.Get(file.ID)
	requested := request.Header.Get("X-WOPI-Lock")
	if held != "" && requested != "" && requested != held {
		h.logger.Warn("wopi put file lock conflict",
			zap.String("file", file.Name), zap.String("held", held), zap.String("requested", requested))
		writer.Header().Set("X-WOPI-Lock", held)
		writer.WriteHeader(http.StatusConflict)
		return
	}
	if err := h.files.Write(file.Name, request.Body); err != nil {
		h.logger.Error("wopi put file failed", zap.String("file", file.Name), zap.Error(err))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if info, err := os.Stat(file.Path); err == nil {
		writer.Header().Set("X-WOPI-ItemVersion", version(info))
	}
	h.logger.Info("wopi saved", zap.String("file", file.Name))
	writer.WriteHeader(http.StatusOK)
}

func (h *WopiHost) Override(writer http.ResponseWriter, request *http.Request, file wopiFile) {
	override := request.Header.Get("X-WOPI-Override")
	requested := request.Header.Get("X-WOPI-Lock")
	held := h.locks.Get(file.ID)

	switch override {
	case "GET_LOCK":
		writer.Header().Set("X-WOPI-Lock", held)
		writer.WriteHeader(http.StatusOK)
	case "LOCK":
		if held != "" && held != requested {
			h.logger.Warn("wopi lock conflict",
				zap.String("file", file.Name), zap.String("held", held), zap.String("requested", requested))
			writer.Header().Set("X-WOPI-Lock", held)
			writer.WriteHeader(http.StatusConflict)
			return
		}
		h.locks.Set(file.ID, requested)
		if info, err := os.Stat(file.Path); err == nil {
			writer.Header().Set("X-WOPI-ItemVersion", version(info))
		}
		writer.WriteHeader(http.StatusOK)
	case "REFRESH_LOCK":
		if held == "" || held != requested {
			writer.Header().Set("X-WOPI-Lock", held)
			writer.WriteHeader(http.StatusConflict)
			return
		}
		h.locks.Set(file.ID, requested)
		writer.WriteHeader(http.StatusOK)
	case "UNLOCK":
		if held == "" || held != requested {
			writer.Header().Set("X-WOPI-Lock", held)
			writer.WriteHeader(http.StatusConflict)
			return
		}
		h.locks.Clear(file.ID)
		writer.WriteHeader(http.StatusOK)
	default:
		h.logger.Info("unsupported wopi override",
			zap.String("override", override), zap.String("file", file.Name))
		writer.WriteHeader(http.StatusNotImplemented)
	}
}

func version(info os.FileInfo) string {
	return strconv.FormatInt(info.ModTime().UnixNano(), 10)
}
