package backend

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const discoveryTTL = 5 * time.Minute

type Discovery struct {
	url    string
	client *http.Client

	mutex   sync.Mutex
	actions map[string]string
	fetched time.Time
}

func NewDiscovery(url string) *Discovery {
	return &Discovery{
		url:    url,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (d *Discovery) EditorURL(extension string) (string, error) {
	actions, err := d.load()
	if err != nil {
		return "", err
	}
	urlsrc, found := actions[strings.ToLower(extension)]
	if !found {
		return "", fmt.Errorf("no editor for .%s", extension)
	}
	return urlsrc, nil
}

func (d *Discovery) load() (map[string]string, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.actions != nil && time.Since(d.fetched) < discoveryTTL {
		return d.actions, nil
	}

	response, err := d.client.Get(d.url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery returned %s", response.Status)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var document discoveryDocument
	if err := xml.Unmarshal(body, &document); err != nil {
		return nil, err
	}

	actions := document.Actions()
	if len(actions) == 0 {
		return nil, fmt.Errorf("discovery listed no actions")
	}

	d.actions = actions
	d.fetched = time.Now()
	return actions, nil
}
