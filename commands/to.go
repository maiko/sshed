package commands

import (
	"fmt"
	"os"

	"github.com/maiko/sshed/host"
	"github.com/maiko/sshed/ssh"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func (cmds *Commands) newToCommand() cli.Command {
	return cli.Command{
		Name:      "to",
		Usage:     "Connects to host",
		ArgsUsage: "<key>",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "verbose, v",
				Usage: "verbose ssh output",
			},
		},
		BashComplete: func(c *cli.Context) {
			// This will complete if no args are passed
			if c.NArg() > 0 {
				return
			}
			cmds.completeWithServers()
		},
		Action: cmds.toAction,
	}
}

func (cmds *Commands) toAction(c *cli.Context) (err error) {
	var key string
	var srv *host.Host

	if c.NArg() == 0 {
		key, err = cmds.askServerKey()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to host: %s\n", err)
			return err
		}
	} else {
		key = c.Args().First()
	}

	srv = ssh.Config.Get(key)
	if srv == nil {
		return errors.New("host not found")
	}

	cmd, err := cmds.createCommand(c, srv, &options{verbose: c.Bool("verbose")}, "")
	if err != nil {
		return err
	}

	return cmd.Run()
}
