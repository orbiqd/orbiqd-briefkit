package briefkitctl

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/orbiqd/orbiqd-briefkit/internal/pkg/cli"
)

type ErrorEvent struct {
	Message string `json:"message"`
}

type Command struct {
	Log   cli.LogConfig   `embed:"" prefix:"log-"`
	Store cli.StoreConfig `embed:"" prefix:"store-"`

	Agent   AgentCmd    `cmd:"" help:"Manage agents"`
	State   StateCmd    `cmd:"" help:"Manage state"`
	Exec    ExecCmd     `cmd:"" help:"Run a prompt with specified model"`
	Setup   SetupCmd    `cmd:"" help:"Setup BriefKit environment"`
	Version VersionFlag `name:"version" help:"Print version information."`
}

type VersionFlag bool

func (v VersionFlag) BeforeApply(app *kong.Kong) error {
	fmt.Println(app.Model.Vars()["version"])
	app.Exit(0)
	return nil
}
