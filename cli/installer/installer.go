package installer

import (
	"fmt"
	cp "github.com/otiai10/copy"
	"github.com/syncloud/golib/config"
	"github.com/syncloud/golib/linux"
	"github.com/syncloud/golib/platform"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"path"
)

type Variables struct {
	SnapData string
	Domain   string
}

const (
	App = "collabora"
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

	err = i.StorageChange()
	if err != nil {
		return err
	}

	err = i.UpdateConfigs()
	if err != nil {
		return err
	}

	err = linux.CreateMissingDirs(
		path.Join(i.commonDir, "log"),
		path.Join(i.commonDir, "nginx"),
		path.Join(i.dataDir, "coolwsd"),
		path.Join(i.dataDir, "systemplate"),
		path.Join(i.dataDir, "child-roots"),
	)
	if err != nil {
		return err
	}

	coolFileserverPath := path.Join(i.dataDir, "coolwsd")
	err = cp.Copy(path.Join(i.dataDir, "config", "discovery.xml"), path.Join(coolFileserverPath, "discovery.xml"))
	if err != nil {
		return err
	}

	err = i.CopyBrowserAssets(coolFileserverPath)
	if err != nil {
		return err
	}

	err = i.FixPermissions()
	if err != nil {
		return err
	}
	return nil
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
	err := i.UpdateConfigs()
	if err != nil {
		return err
	}

	coolFileserverPath := path.Join(i.dataDir, "coolwsd")
	err = os.MkdirAll(coolFileserverPath, 0755)
	if err != nil {
		return err
	}

	err = cp.Copy(path.Join(i.dataDir, "config", "discovery.xml"), path.Join(coolFileserverPath, "discovery.xml"))
	if err != nil {
		return err
	}

	err = i.CopyBrowserAssets(coolFileserverPath)
	if err != nil {
		return err
	}

	err = i.ClearVersion()
	if err != nil {
		return err
	}

	err = i.FixPermissions()
	if err != nil {
		return err
	}
	return nil
}

func (i *Installer) StorageChange() error {
	storageDir, err := i.platformClient.InitStorage(App, App)
	if err != nil {
		return err
	}

	err = linux.Chown(storageDir, App)
	if err != nil {
		return err
	}

	return nil
}

func (i *Installer) AccessChange() error {
	return i.UpdateConfigs()
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

	variables := Variables{
		SnapData: i.dataDir,
		Domain:   domain,
	}

	err = config.Generate(
		path.Join(i.appDir, "config"),
		path.Join(i.dataDir, "config"),
		variables,
	)
	if err != nil {
		return err
	}
	return nil
}

func (i *Installer) CopyBrowserAssets(coolFileserverPath string) error {
	command := exec.Command(
		"cp", "-r",
		path.Join(i.appDir, "app", "usr", "share", "coolwsd", "browser"),
		coolFileserverPath,
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
	err = linux.Chown(i.commonDir, App)
	if err != nil {
		return err
	}
	return nil
}
