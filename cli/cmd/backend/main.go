package main

import (
	"context"
	"fmt"
	"hooks/backend"
	"hooks/installer"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/syncloud/golib/log"
)

func main() {
	logger := log.Logger()

	cmd := &cobra.Command{
		Use:          "backend",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			appDir := fmt.Sprintf("/snap/%s/current", installer.App)
			dataDir := fmt.Sprintf("/var/snap/%s/current", installer.App)

			secret, err := installer.New(logger).WopiSecret()
			if err != nil {
				return err
			}

			baseURL := os.Getenv("COLLABORA_BASE_URL")
			auth := backend.NewAuth(backend.OIDCConfig{
				Issuer:       os.Getenv("COLLABORA_AUTH_URL"),
				ClientID:     os.Getenv("COLLABORA_OIDC_CLIENT_ID"),
				ClientSecret: os.Getenv("COLLABORA_OIDC_CLIENT_SECRET"),
				RedirectURL:  baseURL + installer.OIDCRedirect,
				BaseURL:      baseURL,
				AuthSocket:   os.Getenv("COLLABORA_AUTH_SOCKET"),
			}, []byte(secret), logger)

			files := backend.NewFileStore(os.Getenv("COLLABORA_FILES_DIR"), path.Join(appDir, "samples"))
			locks := backend.NewLockStore()
			filesAPI := backend.NewFilesAPI(files, locks, logger)

			server := backend.NewServer(
				backend.Config{
					Socket:      path.Join(dataDir, "run", "backend.sock"),
					WopiListen:  installer.WopiListen,
					WopiBaseURL: installer.WopiBaseURL,
					BaseURL:     baseURL,
					Secret:      []byte(secret),
				},
				files,
				auth,
				backend.NewSessionAPI(auth),
				filesAPI,
				backend.NewEditorAPI(files, backend.NewDiscovery(installer.DiscoveryURL),
					[]byte(secret), baseURL, installer.WopiBaseURL),
				backend.NewWopiHost(files, locks, filesAPI, []byte(secret), baseURL, logger),
				logger,
			)
			return server.Run(context.Background())
		},
	}

	if err := cmd.Execute(); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
