package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/SwayKh/linksym/pkg/config"
)

// Load config, Setup up Global variables and handle all subcommand switching
func Run() error {
	CreateFlags()
	flag.Parse()

	configName := ".linksym.yaml"

	err := config.InitialiseHomePath()
	if err != nil {
		return err
	}

	// Since the Init Command creates the config file, the LoadConfig function
	// can't be called before handling the init subcommand.
	// But Init function calls aliasPath, which requires HomeDirectory variable,
	// and hence function SetupDirectories was split up
	if flag.Arg(0) == "init" {
		return Init(configName)
	}

	configuration, err := config.LoadConfig(configName)
	if err != nil {
		return err
	}
	config.SetupDirectories(configuration, configName)

	if *HelpFlag {
		Help()
		os.Exit(0)
	}

	switch flag.Arg(0) {
	case "":
		Help()
	case "init":
		break
	case "add":
		err = Add(configuration, os.Args[2:])
	case "remove":
		err = Remove(configuration, os.Args[2:])
	case "source":
		err = Source()
	default:
		err = fmt.Errorf("Invalid Command. Please use -h or --help flags to see available commands.")
	}

	if err != nil {
		return err
	}

	if err := config.WriteConfig(configuration); err != nil {
		return err
	}
	return nil
}
