package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SwayKh/linksym/pkg/config"
	"github.com/SwayKh/linksym/pkg/linker"
)

// Initialise and empty config with cwd as init directory
func Init() error {
	err := config.InitialiseConfig()
	if err != nil {
		return err
	}
	return nil
}

func Add(args []string) error {
	var sourcePath, destinationPath string
	var err error
	var isDirectory bool

	switch len(args) {

	case 1:
		// Set first arg source path, get absolute path, check if it exists, set the
		// destination path as cwd+filename of source path

		sourcePath, err = filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("Error getting absolute path of file %s: \n%w", sourcePath, err)
		}

		fileExists, fileInfo, err := config.CheckFile(sourcePath)
		if err != nil {
			return err
		} else if fileInfo.IsDir() {
			isDirectory = true
		} else if !fileExists {
			return fmt.Errorf("File %s doesn't exist", sourcePath)
		}

		filename := filepath.Base(sourcePath)
		destinationPath = filepath.Join(config.InitDirectory, filename)

	case 2:
		// set first and second args as source and destination path, get absolute
		// paths, check if the paths exist, plus handle the special case of source
		// path not existing but destination path exists, hence creating a link
		// without the moving the files

		sourcePath, err = filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("Error getting absolute path of file %s: \n%w", sourcePath, err)
		}

		destinationPath, err = filepath.Abs(args[1])
		if err != nil {
			return fmt.Errorf("Error getting absolute path of file %s: \n%w", destinationPath, err)
		}

		sourceFileExists, sourceFileInfo, err := config.CheckFile(sourcePath)
		if err != nil {
			return err
		}

		destinationFileExists, DestinationFileInfo, err := config.CheckFile(destinationPath)
		if err != nil {
			return err
		}

		if destinationFileExists && DestinationFileInfo.IsDir() {
			filename := filepath.Base(sourcePath)
			destinationPath = filepath.Join(destinationPath, filename)
			isDirectory = true
		}

		if sourceFileExists && sourceFileInfo.IsDir() && destinationFileExists {
			filename := filepath.Base(destinationPath)
			sourcePath = filepath.Join(sourcePath, filename)

			err := linker.Link(sourcePath, destinationPath)
			if err != nil {
				return err
			}
			return nil
		}

		if destinationFileExists && !sourceFileExists {
			err := linker.Link(sourcePath, destinationPath)
			if err != nil {
				return err
			}
			return nil
		}

	default:
		return fmt.Errorf("Invalid number of arguments")
	}

	err = linker.MoveAndLink(sourcePath, destinationPath, isDirectory)
	if err != nil {
		return err
	}
	return nil
}

func Remove(linkName string) error {
	var linkPath string
	var sourcePath, destinationPath string
	var err error
	var isDirectory bool

	configuration, err := config.LoadConfig(config.ConfigPath)
	if err != nil {
		return err
	}

	linkPath, err = filepath.Abs(linkName)
	if err != nil {
		return fmt.Errorf("Error getting absolute path of file %s: \n%w", linkPath, err)
	}

	fileExists, fileInfo, err := config.CheckFile(linkPath)
	if err != nil {
		return err
	} else if fileInfo.IsDir() {
		isDirectory = true
	} else if !fileExists {
		return fmt.Errorf("File %s doesn't exist", linkPath)
	}

	recordPathName := filepath.Join(filepath.Base(filepath.Dir(linkPath)), filepath.Base(linkPath))

	for i := range configuration.Records {
		if configuration.Records[i].Name == recordPathName {
			sourcePath = configuration.Records[i].Paths[0]
			destinationPath = configuration.Records[i].Paths[1]
			err = config.RemoveRecord(i)
			if err != nil {
				return nil
			}
		}
	}

	err = linker.UnLink(sourcePath, destinationPath, isDirectory)
	if err != nil {
		return nil
	}
	return nil
}

func Source() error {
	return nil
}

func Help() {
	fmt.Println("Usage: linksym [subcommand] [flags]")

	fmt.Println("\n Subcommands:")
	fmt.Println("   add [Path] [(optional) Destination]:")
	fmt.Println("     Create a symlink for given path, optionally define a destination for symlink")
	fmt.Println("   remove [Path]")
	fmt.Println("     Remove the symlink and move the file to the original path")

	fmt.Println("\n Flags:")
	fmt.Println("   -h, --help")
	fmt.Println("     Print this help message")
	os.Exit(0)
}
