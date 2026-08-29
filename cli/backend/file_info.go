package backend

import (
	"encoding/base64"
	"path/filepath"
	"strings"
	"time"
)

type FileInfo struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mtime"`
	Kind    string    `json:"kind"`
}

func NewFileInfo(name string, size int64, modTime time.Time) FileInfo {
	return FileInfo{
		ID:      FileID(name),
		Name:    name,
		Size:    size,
		ModTime: modTime,
		Kind:    Kind(name),
	}
}

func FileID(name string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(name))
}

func FileName(id string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func Kind(name string) string {
	switch strings.ToLower(strings.TrimPrefix(filepath.Ext(name), ".")) {
	case "docx", "doc", "odt", "rtf", "txt":
		return "document"
	case "xlsx", "xls", "ods", "csv":
		return "spreadsheet"
	case "pptx", "ppt", "odp":
		return "presentation"
	case "pdf":
		return "pdf"
	}
	return "unknown"
}
