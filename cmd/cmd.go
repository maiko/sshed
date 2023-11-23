package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/maiko/sshed/commands"
	"github.com/maiko/sshed/keychain"
	"github.com/maiko/sshed/ssh"
	"github.com/mgutz/ansi"
	"github.com/urfave/cli"
)

var version, build string

func main() {

	app := cli.NewApp()

	app.Name = "sshed"
	app.Usage = "SSH config editor and hosts manager"
	app.Authors = []cli.Author{
		{
			Name:  "Eugene Terentev",
			Email: "eugene@terentev.net",
		},
		{
			Name:  "Maiko BOSSUYT",
			Email: "hello@maiko-bossuyt.eu",
		},
	}

	if version != "" && build != "" {
		app.Version = fmt.Sprintf("%s (build %s)", version, build)
	}

	usr, _ := user.Current()
	homeDir := usr.HomeDir

	app.HelpName = "help"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "keychain",
			EnvVar: "SSHED_KEYCHAIN",
			Value:  filepath.Join(homeDir, ".sshed", "keychain"),
			Usage:  "path to keychain database",
		},
		cli.StringFlag{
			Name:   "config",
			EnvVar: "SSHED_CONFIG_FILE",
			Value:  filepath.Join(homeDir, ".ssh", "config"),
			Usage:  "path to SSH config file",
		},
		cli.StringFlag{
			Name:   "backup-dir",
			EnvVar: "SSHED_BACKUP_DIR",
			Value:  filepath.Join(homeDir, ".sshed", "backup"),
			Usage:  "path to backup directory",
		},
		cli.StringFlag{
			Name:   "ssh-path",
			EnvVar: "SSHED_SSH_BIN",
			Value:  "ssh",
			Usage:  "path to SSH binary",
		},
		cli.StringFlag{
			Name:   "scp-path",
			EnvVar: "SSHED_SCP_BIN",
			Value:  "scp",
			Usage:  "path to SCP binary",
		},
	}

	app.EnableBashCompletion = true

	app.Before = func(context *cli.Context) error {
		if context.Command.Name == "help" {
			return nil
		}

		var err error
		ssh.Config, err = ssh.Parse(context.String("config"))
		if err != nil {
			return err
		}

		dbpath := context.String("keychain")

		err = keychain.Open(dbpath)
		return err
	}

	commands.RegisterCommands(app)

	err := app.Run(os.Args)

	if err != nil {
		fmt.Println(ansi.Red, fmt.Sprintf("Error: %s", err))
		os.Exit(1)
	}
}
