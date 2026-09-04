package backend

import (
	"regexp"
	"strings"
)

var placeholders = regexp.MustCompile(`<[^>]*>`)

type discoveryDocument struct {
	NetZone discoveryNetZone `xml:"net-zone"`
}

func (d discoveryDocument) Actions() map[string]string {
	actions := map[string]string{}
	for _, app := range d.NetZone.Apps {
		for _, action := range app.Actions {
			if action.Ext == "" {
				continue
			}
			extension := strings.ToLower(action.Ext)
			if _, taken := actions[extension]; taken && action.Name != "edit" {
				continue
			}
			actions[extension] = placeholders.ReplaceAllString(action.URLSrc, "")
		}
	}
	return actions
}
