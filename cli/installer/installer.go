package installer

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	cp "github.com/otiai10/copy"
	"github.com/syncloud/golib/config"
	"github.com/syncloud/golib/linux"
	"github.com/syncloud/golib/platform"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
)

const (
	App            = "collabora"
	OIDCRedirect   = "/oidc/callback"
	OIDCAuthMethod = "client_secret_basic"
	WopiListen     = "127.0.0.1:9981"
	WopiBaseURL    = "http://127.0.0.1:9981"
	DiscoveryURL   = "http://127.0.0.1:9980/hosting/discovery"
	AdminUser      = "admin"
)

type Installer struct {
	newVersionFile     string
	currentVersionFile string
	platformClient     *platform.Client
	appDir             string
	dataDir            string
	commonDir          string
	logger             *zap.Logger
}

func New(logger *zap.Logger) *Installer {
	appDir := fmt.Sprintf("/snap/%s/current", App)
	dataDir := fmt.Sprintf("/var/snap/%s/current", App)
	commonDir := fmt.Sprintf("/var/snap/%s/common", App)

	return &Installer{
		newVersionFile:     path.Join(appDir, "version"),
		currentVersionFile: path.Join(dataDir, "version"),
		platformClient:     platform.New(),
		appDir:             appDir,
		dataDir:            dataDir,
		commonDir:          commonDir,
		logger:             logger,
	}
}

func (i *Installer) Install() error {
	err := linux.CreateUser(App)
	if err != nil {
		return err
	}

	err = linux.CreateMissingDirs(
		path.Join(i.dataDir, "nginx"),
		path.Join(i.dataDir, "run"),
		path.Join(i.dataDir, "secret"),
		path.Join(i.dataDir, "coolwsd"),
		path.Join(i.dataDir, "systemplate"),
		path.Join(i.dataDir, "child-roots"),
	)
	if err != nil {
		return err
	}

	err = i.StorageChange()
	if err != nil {
		return err
	}

	err = i.UpdateConfigs()
	if err != nil {
		return err
	}

	err = i.CopyFileserverAssets()
	if err != nil {
		return err
	}

	return i.FixPermissions()
}

func (i *Installer) Configure() error {
	err := i.StorageChange()
	if err != nil {
		return err
	}
	return i.UpdateVersion()
}

func (i *Installer) PreRefresh() error {
	return nil
}

func (i *Installer) PostRefresh() error {
	err := linux.CreateMissingDirs(
		path.Join(i.dataDir, "nginx"),
		path.Join(i.dataDir, "run"),
		path.Join(i.dataDir, "secret"),
		path.Join(i.dataDir, "coolwsd"),
	)
	if err != nil {
		return err
	}

	err = i.UpdateConfigs()
	if err != nil {
		return err
	}

	err = i.CopyFileserverAssets()
	if err != nil {
		return err
	}

	err = i.RemoveLegacyCommonDirs()
	if err != nil {
		return err
	}

	err = i.ClearVersion()
	if err != nil {
		return err
	}

	return i.FixPermissions()
}

func (i *Installer) StorageChange() error {
	storageDir, err := i.platformClient.InitStorage(App, App)
	if err != nil {
		return err
	}

	err = linux.CreateMissingDirs(path.Join(storageDir, "files"))
	if err != nil {
		return err
	}

	return linux.Chown(storageDir, App)
}

func (i *Installer) AccessChange() error {
	err := i.UpdateConfigs()
	if err != nil {
		return err
	}
	return i.RestartServices()
}

func (i *Installer) RestartServices() error {
	for _, service := range []string{"server", "backend"} {
		err := i.platformClient.RestartService(fmt.Sprintf("%s.%s", App, service))
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Installer) BackupPreStop() error {
	return nil
}

func (i *Installer) RestorePreStart() error {
	return nil
}

func (i *Installer) RestorePostStart() error {
	err := i.StorageChange()
	if err != nil {
		return err
	}
	return i.FixPermissions()
}

func (i *Installer) RemoveLegacyCommonDirs() error {
	for _, dir := range []string{"log", "nginx"} {
		err := os.RemoveAll(path.Join(i.commonDir, dir))
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Installer) ClearVersion() error {
	return os.RemoveAll(i.currentVersionFile)
}

func (i *Installer) UpdateVersion() error {
	return cp.Copy(i.newVersionFile, i.currentVersionFile)
}

func (i *Installer) UpdateConfigs() error {
	domain, err := i.platformClient.GetAppDomainName(App)
	if err != nil {
		return err
	}

	deviceDomain, err := i.platformClient.GetDeviceDomainName()
	if err != nil {
		return err
	}

	appUrl, err := i.platformClient.GetAppUrl(App)
	if err != nil {
		return err
	}

	authUrl, err := i.platformClient.GetAppUrl("auth")
	if err != nil {
		return err
	}

	storageDir, err := i.platformClient.InitStorage(App, App)
	if err != nil {
		return err
	}

	adminPassword, err := i.AdminPassword()
	if err != nil {
		return err
	}

	oidcSecret, err := i.OIDCClientSecret(strings.TrimRight(appUrl, "/"))
	if err != nil {
		return err
	}

	variables := Variables{
		SnapData:            i.dataDir,
		SnapApp:             i.appDir,
		Domain:              domain,
		DeviceDomainPattern: regexp.QuoteMeta(deviceDomain),
		AuthLocalSocket:     i.platformClient.GetAuthLocalSocket(),
		AuthSocketPath:      AuthSocketPath(i.platformClient.GetAuthLocalSocket()),
		AdminUser:           AdminUser,
		AdminPassword:       adminPassword,
		AdminAuthorization:  basicAuthorization(AdminUser, adminPassword),
		WopiHost:            WopiBaseURL,
		FilesDir:            path.Join(storageDir, "files"),
		AppUrl:              strings.TrimRight(appUrl, "/"),
		AuthUrl:             strings.TrimRight(authUrl, "/"),
		OIDCClientID:        App,
		OIDCClientSecret:    oidcSecret,
	}

	return config.Generate(
		path.Join(i.appDir, "config"),
		path.Join(i.dataDir, "config"),
		variables,
	)
}

func (i *Installer) AdminPassword() (string, error) {
	return i.secret("admin.password")
}

func (i *Installer) WopiSecret() (string, error) {
	return i.secret("wopi.key")
}

func (i *Installer) OIDCClientSecret(appUrl string) (string, error) {
	secretFile := path.Join(i.dataDir, "secret", "oidc.secret")
	urlFile := path.Join(i.dataDir, "secret", "oidc.url")

	secret := readSecret(secretFile)
	if secret != "" && readSecret(urlFile) == appUrl {
		return secret, nil
	}

	secret, err := i.platformClient.RegisterOIDCClient(
		App,
		[]string{OIDCRedirect},
		true,
		OIDCAuthMethod,
	)
	if err != nil {
		return "", err
	}
	err = writeSecret(secretFile, secret)
	if err != nil {
		return "", err
	}
	err = writeSecret(urlFile, appUrl)
	if err != nil {
		return "", err
	}
	return secret, nil
}

func readSecret(file string) string {
	content, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(content))
}

func (i *Installer) secret(name string) (string, error) {
	file := path.Join(i.dataDir, "secret", name)
	if existing := readSecret(file); existing != "" {
		return existing, nil
	}
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	secret := base64.RawURLEncoding.EncodeToString(buffer)
	if err := writeSecret(file, secret); err != nil {
		return "", err
	}
	return secret, nil
}

func writeSecret(file, secret string) error {
	if err := os.MkdirAll(path.Dir(file), 0o750); err != nil {
		return err
	}
	return os.WriteFile(file, []byte(secret), 0o640)
}

func AuthSocketPath(localSocket string) string {
	return strings.TrimSuffix(strings.TrimPrefix(localSocket, "http://unix:"), ":")
}

func basicAuthorization(user, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+password))
}

func (i *Installer) CopyFileserverAssets() error {
	fileserver := path.Join(i.dataDir, "coolwsd")
	err := os.MkdirAll(fileserver, 0o755)
	if err != nil {
		return err
	}

	err = cp.Copy(
		path.Join(i.dataDir, "config", "discovery.xml"),
		path.Join(fileserver, "discovery.xml"),
	)
	if err != nil {
		return err
	}

	command := exec.Command(
		"cp", "-r",
		path.Join(i.appDir, "app", "usr", "share", "coolwsd", "browser"),
		fileserver,
	)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

func (i *Installer) FixPermissions() error {
	err := linux.Chown(i.dataDir, App)
	if err != nil {
		return err
	}
	return linux.Chown(i.commonDir, App)
}
