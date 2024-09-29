package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SwayKh/linksym/pkg/config"
)

// Loop over the configuration []Records, for each entry get the source and
// destination paths. Run the Link command for each entry.
func Source(configuration *config.AppConfig) error {
	fmt.Println("Creating Symlinks from .linksym.yaml Records...")
	for _, record := range configuration.Records {
		sourcePath := record.Paths[0]
		destinationPath := filepath.Dir(record.Paths[1])

		err := os.MkdirAll(destinationPath, 0o755)
		if err != nil {
			return err
		}

		pathArgs := []string{sourcePath, destinationPath}

		err = Add(configuration, pathArgs, false)
		if err != nil {
			return err
		}
	}
	return nil
}
