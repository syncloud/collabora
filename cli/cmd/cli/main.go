package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/syncloud/golib/log"
	"hooks/installer"
	"os"
)

func main() {
	logger := log.Logger()

	var cmd = &cobra.Command{
		Use:          "cli",
		SilenceUsage: true,
	}

	commands := []struct {
		use string
		run func(*installer.Installer) error
	}{
		{"storage-change", func(i *installer.Installer) error { return i.StorageChange() }},
		{"access-change", func(i *installer.Installer) error { return i.AccessChange() }},
		{"backup-pre-stop", func(i *installer.Installer) error { return i.BackupPreStop() }},
		{"restore-pre-start", func(i *installer.Installer) error { return i.RestorePreStart() }},
		{"restore-post-start", func(i *installer.Installer) error { return i.RestorePostStart() }},
	}

	for _, command := range commands {
		use, run := command.use, command.run
		cmd.AddCommand(&cobra.Command{
			Use: use,
			RunE: func(cmd *cobra.Command, args []string) error {
				logger.Info(use)
				return run(installer.New(logger))
			},
		})
	}

	err := cmd.Execute()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}
