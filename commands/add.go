package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SwayKh/linksym/config"
	"github.com/SwayKh/linksym/link"
	"github.com/SwayKh/linksym/logger"
)

// Add function, which handles the Add subcommand and handles all scenarios of
// file paths provided.
// Handling one argument is simple enough, just Move the
// file to cwd and link it.
// Handling 2 arguments creates lots of different scenario of combination of
// files and directory, and handling the special scenario of a File/Dir which is
// already moved by the user, and just needs to be linked, Skipping the moving
// of file step of the Linking process
func (app *Application) Add(args []string, updateRecord bool) error {
	toMove := true

	switch len(args) {
	case 1:
		source, err := config.GetFileInfo(args[0])
		if err != nil {
			return err
		}

		if !source.Exists {
			return fmt.Errorf("File %s doesn't exist", source.AbsPath)
		}

		logger.VerboseLog(logger.SUCCESS, "Source path exists: %s", config.AliasPath(source.AbsPath, app.HomeDirectory, app.InitDirectory, true))

		sourcePath := source.AbsPath
		filename := filepath.Base(sourcePath)
		destinationPath := filepath.Join(app.InitDirectory, filename)

		logger.VerboseLog(logger.SUCCESS, "Destination path exists: %s", config.AliasPath(destinationPath, app.HomeDirectory, app.InitDirectory, true))

		paths := link.LinkPaths{
			SourcePath:      sourcePath,
			DestinationPath: destinationPath,
			HomeDir:         app.HomeDirectory,
			InitDir:         app.InitDirectory,
			IsDirectory:     source.IsDir,
		}

		err = paths.MoveAndLink()
		if err != nil {
			return err
		}
		if updateRecord {
			app.Configuration.AddRecord(sourcePath, destinationPath)
		}

	case 2:
		source, err := config.GetFileInfo(args[0])
		if err != nil {
			return err
		}

		destination, err := config.GetFileInfo(args[1])
		if err != nil {
			return err
		}

		// For Source and Destination paths, to Exist, !Exist, be a Dir or a File
		// respectively creates 16 different combination of booleans,
		isSourceDir := source.Exists && source.IsDir
		isSourceFile := source.Exists && !source.IsDir
		isDestinationDir := destination.Exists && destination.IsDir
		isDestinationFile := destination.Exists && !destination.IsDir

		sourcePath := source.AbsPath
		destinationPath := destination.AbsPath

		logger.VerboseLog(logger.SUCCESS, "Source path: %s", config.AliasPath(source.AbsPath, app.HomeDirectory, app.InitDirectory, true))
		logger.VerboseLog(logger.SUCCESS, "Destination path: %s", config.AliasPath(destination.AbsPath, app.HomeDirectory, app.InitDirectory, true))

		switch {
		// Link Source File to inside of Destination directory
		case isSourceFile && isDestinationDir:
			destinationPath = appendToDestinationPath(source.AbsPath, destination.AbsPath)

		case isSourceFile && isDestinationFile:
			return fmt.Errorf("Destination file %s already exists", destination.AbsPath)

		// Link Source file to Destination by using path as File or Directory based
		// on trailling / provided with argument
		case isSourceFile && !destination.Exists:
			if destination.HasSlash {
				err := os.MkdirAll(destinationPath, 0o755)
				if err != nil {
					return err
				}
				destinationPath = appendToDestinationPath(source.AbsPath, destination.AbsPath)
			}

		// Link Source Directory to inside of Destination directory
		case isSourceDir && isDestinationDir:
			destinationPath = appendToDestinationPath(source.AbsPath, destination.AbsPath)

		// Can't link a Directory to a File
		case isSourceDir && isDestinationFile:
			return fmt.Errorf("Can't link a Directory: %s to a File: %s", source.AbsPath, destination.AbsPath)

		// Link Source directory to Destination by using path as File or Directory
		// based on trailling / provided with argument. But can't link a Directory
		// to a File
		case isSourceDir && !destination.Exists:
			if destination.HasSlash {
				err := os.MkdirAll(destinationPath, 0o755)
				if err != nil {
					return err
				}
				destinationPath = appendToDestinationPath(source.AbsPath, destination.AbsPath)
			} else {
				return fmt.Errorf("Can't link a Directory: %s to a File: %s", source.AbsPath, destination.AbsPath)
			}

		// Source Doesn't exists, But Destination does, and is a file and the Source
		// can be a directory path or a file path
		case !source.Exists && isDestinationFile:
			if source.HasSlash {
				// Given Source path has a trailing /, hence it's a directory
				return fmt.Errorf("Can't Link a Directory %s to a File %s", source.AbsPath, destination.AbsPath)
			} else {
				// Source is a file which doesn't exist, Destination is a file
				toMove = false
			}

		// Source Doesn't exists(Can be file or dir), But Destination does, and is a directory
		case !source.Exists && isDestinationDir:
			if source.HasSlash {
				// Given Source path has a trailing /, hence it's a directory
				toMove = false
			} else {
				// Else Source is a file, and destination is a directory
				return fmt.Errorf("Can't link a File: %s to a Directory: %s", source.AbsPath, destination.AbsPath)
			}

		// Source and Destination Both Don't Exist
		case !source.Exists && !destination.Exists:
			return fmt.Errorf("Source and Destination paths don't exist, Nothing to Link")

		default:
			return fmt.Errorf("Invalid arguments provided")
		}

		paths := link.LinkPaths{
			SourcePath:      sourcePath,
			DestinationPath: destinationPath,
			HomeDir:         app.HomeDirectory,
			InitDir:         app.InitDirectory,
			IsDirectory:     source.IsDir,
		}

		if toMove {
			err = paths.MoveAndLink()
		} else {
			err = paths.Link()
		}
		if err != nil {
			return err
		}
		if updateRecord {
			app.Configuration.AddRecord(sourcePath, destinationPath)
		}

	default:
		return fmt.Errorf("Invalid number of arguments")
	}
	return nil
}

// Append filename from Source path to Destination path
func appendToDestinationPath(sourcePath, destinationPath string) string {
	filename := filepath.Base(sourcePath)
	destinationPath = filepath.Join(destinationPath, filename)

	return destinationPath
}
