package backend

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileStore struct {
	root      string
	templates string
}

func NewFileStore(root, templates string) *FileStore {
	return &FileStore{root: root, templates: templates}
}

func (s *FileStore) Init() error {
	return os.MkdirAll(s.root, 0o770)
}

func (s *FileStore) path(name string) (string, error) {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\") {
		return "", errors.New("invalid file name")
	}
	return filepath.Join(s.root, name), nil
}

func (s *FileStore) PathByID(id string) (string, string, error) {
	name, err := FileName(id)
	if err != nil {
		return "", "", err
	}
	path, err := s.path(name)
	if err != nil {
		return "", "", err
	}
	return path, name, nil
}

func (s *FileStore) List() ([]FileInfo, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		return nil, err
	}
	files := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, NewFileInfo(entry.Name(), info.Size(), info.ModTime()))
	}
	sort.Slice(files, func(i, j int) bool { return files[i].ModTime.After(files[j].ModTime) })
	return files, nil
}

func (s *FileStore) Stat(name string) (os.FileInfo, error) {
	path, err := s.path(name)
	if err != nil {
		return nil, err
	}
	return os.Stat(path)
}

func (s *FileStore) Open(name string) (*os.File, error) {
	path, err := s.path(name)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (s *FileStore) Write(name string, src io.Reader) error {
	path, err := s.path(name)
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(s.root, ".upload-*")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if _, err := io.Copy(temp, src); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(temp.Name(), 0o660); err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}

func (s *FileStore) Delete(name string) error {
	path, err := s.path(name)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func (s *FileStore) CreateFromTemplate(name, kind string) error {
	path, err := s.path(name)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return errors.New("file already exists")
	}
	if Kind("x."+kind) == "unknown" {
		return errors.New("unsupported document kind")
	}
	template, err := os.Open(filepath.Join(s.templates, "blank."+kind))
	if err != nil {
		return err
	}
	defer template.Close()
	return s.Write(name, template)
}
