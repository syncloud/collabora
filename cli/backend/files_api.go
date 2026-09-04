package backend

import (
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type FilesAPI struct {
	files  *FileStore
	locks  *LockStore
	logger *zap.Logger
}

func NewFilesAPI(files *FileStore, locks *LockStore, logger *zap.Logger) *FilesAPI {
	return &FilesAPI{files: files, locks: locks, logger: logger}
}

func (a *FilesAPI) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/files", a.List)
	mux.HandleFunc("POST /api/files", a.Create)
	mux.HandleFunc("PUT /api/files/{id}", a.Upload)
	mux.HandleFunc("DELETE /api/files/{id}", a.Delete)
	mux.HandleFunc("GET /api/files/{id}/contents", a.Contents)
}

func (a *FilesAPI) List(writer http.ResponseWriter, _ *http.Request) {
	files, err := a.files.List()
	if err != nil {
		writeError(writer, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, files)
}

func (a *FilesAPI) Create(writer http.ResponseWriter, request *http.Request) {
	var document NewDocument
	if err := json.NewDecoder(request.Body).Decode(&document); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	name := document.FileName()
	if err := a.files.CreateFromTemplate(name, document.Kind); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(writer, http.StatusCreated, map[string]string{"id": FileID(name), "name": name})
}

func (a *FilesAPI) Upload(writer http.ResponseWriter, request *http.Request) {
	id := request.PathValue("id")
	name, err := FileName(id)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid file id")
		return
	}
	if err := a.files.Write(name, request.Body); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, map[string]string{"id": id, "name": name})
}

func (a *FilesAPI) Delete(writer http.ResponseWriter, request *http.Request) {
	id := request.PathValue("id")
	name, err := FileName(id)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid file id")
		return
	}
	if err := a.files.Delete(name); err != nil {
		writeError(writer, http.StatusBadRequest, err.Error())
		return
	}
	a.locks.Clear(id)
	writeJSON(writer, http.StatusOK, map[string]string{"id": id, "name": name})
}

func (a *FilesAPI) Contents(writer http.ResponseWriter, request *http.Request) {
	name, err := FileName(request.PathValue("id"))
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid file id")
		return
	}
	a.Serve(writer, name)
}

func (a *FilesAPI) Serve(writer http.ResponseWriter, name string) {
	file, err := a.files.Open(name)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	defer file.Close()
	writer.Header().Set("Content-Type", "application/octet-stream")
	if _, err := io.Copy(writer, file); err != nil {
		a.logger.Warn("serving file failed", zap.String("file", name), zap.Error(err))
	}
}
