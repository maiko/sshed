package commands

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/maiko/sshed/ssh"
	"github.com/mgutz/ansi"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"gopkg.in/AlecAivazis/survey.v1"
)

func (cmds *Commands) newAtCommand() cli.Command {
	return cli.Command{
		Name:      "at",
		Usage:     "Executes command on host(s)",
		ArgsUsage: "[key] [command]",
		Action:    cmds.atAction,
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
	}
}
func (cmds *Commands) atAction(c *cli.Context) (err error) {
	keys := []string{c.Args().First()}
	if keys[0] == "" {
		keys, err = cmds.askServersKeys()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to host: %s\n", err)
			return err
		}
	}

	command := c.Args().Get(1)
	if command == "" {

		err = survey.AskOne(&survey.Input{Message: "Command:"}, &command, nil)
		if err != nil {
			return err
		}

		fmt.Println("")
	}

	var wg sync.WaitGroup
	for _, key := range keys {
		var srv = ssh.Config.Get(key)
		if srv == nil {
			return errors.New("host not found")
		}

		if err != nil {
			return err
		}

		wg.Add(1)
		go (func() {
			defer wg.Done()

			cmd, err := cmds.createCommand(c, srv, &options{verbose: c.Bool("verbose")}, command)
			if err != nil {
				log.Panicln(err)
			}

			var buf []byte
			w := bytes.NewBuffer(buf)
			cmd.Stdout = w

			err = cmd.Run()
			if err != nil {
				log.Panicln(err)
			}

			sr, err := io.ReadAll(w)
			if err != nil {
				log.Panicln(err)
			}

			fmt.Printf("%s:\r\n", ansi.Color(srv.Key, "yellow"))
			fmt.Println(string(sr))
		})()
	}

	wg.Wait()

	return err
}
