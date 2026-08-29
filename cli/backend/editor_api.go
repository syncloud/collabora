package backend

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

type EditorAPI struct {
	files       *FileStore
	discovery   *Discovery
	secret      []byte
	baseURL     string
	wopiBaseURL string
}

func NewEditorAPI(files *FileStore, discovery *Discovery, secret []byte, baseURL, wopiBaseURL string) *EditorAPI {
	return &EditorAPI{
		files:       files,
		discovery:   discovery,
		secret:      secret,
		baseURL:     strings.TrimRight(baseURL, "/"),
		wopiBaseURL: wopiBaseURL,
	}
}

func (a *EditorAPI) Open(writer http.ResponseWriter, request *http.Request) {
	id := request.URL.Query().Get("id")
	name, err := FileName(id)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid file id")
		return
	}
	if _, err := a.files.Stat(name); err != nil {
		writeError(writer, http.StatusNotFound, "no such file")
		return
	}

	extension := strings.TrimPrefix(filepath.Ext(name), ".")
	urlsrc, err := a.discovery.EditorURL(extension)
	if err != nil {
		writeError(writer, http.StatusServiceUnavailable, err.Error())
		return
	}

	session, _ := SessionFrom(request.Context())
	token, err := SignAccessToken(a.secret, id, session.Username, AccessTokenTTL)
	if err != nil {
		writeError(writer, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{
		"url":   appendWopiSrc(rebaseURL(urlsrc, a.baseURL), wopiSource(a.wopiBaseURL, id)),
		"token": token,
		"ttl":   time.Now().Add(AccessTokenTTL).UnixMilli(),
		"name":  name,
	})
}

func wopiSource(wopiBaseURL, id string) string {
	return fmt.Sprintf("%s/wopi/files/%s", strings.TrimRight(wopiBaseURL, "/"), id)
}

func rebaseURL(target, base string) string {
	parsedTarget, err := url.Parse(target)
	if err != nil {
		return target
	}
	parsedBase, err := url.Parse(base)
	if err != nil {
		return target
	}
	parsedTarget.Scheme = parsedBase.Scheme
	parsedTarget.Host = parsedBase.Host
	return parsedTarget.String()
}

func appendWopiSrc(urlsrc, source string) string {
	separator := "?"
	if strings.Contains(urlsrc, "?") {
		separator = "&"
	}
	if strings.HasSuffix(urlsrc, "?") || strings.HasSuffix(urlsrc, "&") {
		separator = ""
	}
	return urlsrc + separator + "WOPISrc=" + url.QueryEscape(source)
}
