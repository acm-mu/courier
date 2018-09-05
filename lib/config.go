package lib

import (
	"fmt"
	"os"
	"path"

	homedir "github.com/mitchellh/go-homedir"
)

// ConfigFolder is where the folder is stored in memory
var ConfigFolder string

func init() {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ConfigFolder = path.Join(home, ".config", "courier")
}
