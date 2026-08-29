package backend

import "path/filepath"

type NewDocument struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

func (d NewDocument) FileName() string {
	if filepath.Ext(d.Name) == "" {
		return d.Name + "." + d.Kind
	}
	return d.Name
}
