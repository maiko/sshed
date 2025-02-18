package commands

import (
	"os/user"

	"github.com/maiko/sshed/host"
	"github.com/maiko/sshed/keychain"
	"github.com/maiko/sshed/ssh"
	"github.com/urfave/cli"
	"gopkg.in/AlecAivazis/survey.v1"
)

type answers struct {
	Key            string
	Host           string
	Port           string
	User           string
	Password       string
	KeyFile        string
	KeyFileContent string
	JumpHost       string
}

func (cmds *Commands) newAddCommand() cli.Command {
	return cli.Command{
		Name:      "add",
		Usage:     "Add or edit host",
		ArgsUsage: "[key]",
		Action:    cmds.addAction,
		BashComplete: func(c *cli.Context) {
			// This will complete if no args are passed
			if c.NArg() > 0 {
				return
			}
			cmds.completeWithServers()
		},
	}
}

func (cmds *Commands) addAction(c *cli.Context) error {
	var h *host.Host
	var err error
	var usr, _ = user.Current()
	var key = c.Args().First()

	if key != "" {
		h = ssh.Config.Get(key)
	}

	if h == nil {
		h = &host.Host{
			Key:  key,
			Port: "22",
			User: usr.Username,
		}
	}

	var qs = []*survey.Question{
		{
			Name: "key",
			Prompt: &survey.Input{
				Message: "Alias:",
				Default: h.Key,
			},
			Validate: survey.Required,
		},
		{
			Name: "host",
			Prompt: &survey.Input{
				Message: "Hostname:",
				Default: h.Hostname,
			},
			Validate:  survey.Required,
			Transform: survey.ToLower,
		},
		{
			Name: "port",
			Prompt: &survey.Input{
				Message: "Port:",
				Default: h.Port,
			},
			Transform: survey.ToLower,
		},
		{
			Name: "user",
			Prompt: &survey.Input{
				Message: "User:",
				Help:    "paste single space to leave this field empty (active user will be used when connecting)",
				Default: h.User,
			},
		},
		{
			Name: "password",
			Prompt: &survey.Password{
				Message: "Password (optional):",
			},
		},
	}

	answers := &answers{}

	// perform the questions
	err = survey.Ask(qs, answers)
	if err != nil {
		return err
	}

	askForIdentityFile(answers, h)
	askForJumphost(answers, h)

	h = &host.Host{
		Key:          answers.Key,
		Hostname:     answers.Host,
		Port:         answers.Port,
		User:         answers.User,
		IdentityFile: answers.KeyFile,
		JumpHost:     answers.JumpHost,
		Options:      make(map[string]string),
	}

	isOptions := false
	for {
		err = survey.AskOne(&survey.Confirm{
			Message: "Add additional SSH options?",
		}, &isOptions, nil)
		if err != nil {
			return err
		}
		if isOptions {
			option := struct {
				Key   string
				Value string
			}{}
			optionQuestions := []*survey.Question{
				{
					Name: "key",
					Prompt: &survey.Input{
						Message: "Option:",
					},
					Validate: survey.Required,
				},
				{
					Name: "value",
					Prompt: &survey.Input{
						Message: "Value:",
					},
					Validate: survey.Required,
				},
			}
			err = survey.Ask(optionQuestions, &option)
			if err != nil {
				return err
			}
			h.Options[option.Key] = option.Value
		} else {
			break
		}
	}

	err = keychain.Put(h.Key, &keychain.Record{
		Password:   answers.Password,
		PrivateKey: answers.KeyFileContent,
	})
	if err != nil {
		return err
	}

	ssh.Config.Add(h)

	return ssh.Config.Save()
}

func askForIdentityFile(answers *answers, srv *host.Host) (err error) {
	const OPTION_SKIP = "Leave empty"
	const OPTION_SELECT = "Select known key"
	const OPTION_INPUT = "Input custom path"
	const OPTION_EDITOR = "Paste file contents"

	if srv.IdentityFile != "" {
		var change bool
		err = survey.AskOne(&survey.Confirm{
			Message: "Do you want to change key information?",
		}, &change, nil)
		if err != nil {
			return err
		}

		if !change {
			answers.KeyFile = srv.IdentityFile
			return nil
		}
	}

	var options = []string{
		OPTION_INPUT,
		OPTION_EDITOR,
		OPTION_SKIP,
	}

	if len(ssh.Config.Keys) > 0 {
		options = append(options, OPTION_SELECT)
	}

	var choice string

	err = survey.AskOne(&survey.Select{
		Options: options,
		Message: "How do you want to provide key file?",
	}, &choice, survey.Required)
	if err != nil {
		return err
	}

	switch choice {
	case OPTION_SKIP:
		answers.KeyFile = srv.IdentityFile
		return
	case OPTION_SELECT:
		err = survey.AskOne(&survey.Select{
			Options: ssh.Config.Keys,
			Message: "Choose private key:",
			Default: srv.IdentityFile,
		}, &answers.KeyFile, nil)
	case OPTION_INPUT:
		err = survey.AskOne(&survey.Input{
			Message: "Private key path:",
			Default: srv.IdentityFile,
		}, &answers.KeyFile, nil)
	case OPTION_EDITOR:
		err = survey.AskOne(&survey.Editor{
			Message: "Private key content:",
		}, &answers.KeyFileContent, nil)
	}

	return err
}

func askForJumphost(answers *answers, srv *host.Host) (err error) {
	const OPTION_NONE = "None"
	const OPTION_SELECT = "Select existing host"

	var options = []string{OPTION_NONE}

	if len(ssh.Config.Hosts) > 0 {
		options = append(options, OPTION_SELECT)
	}

	var choice string

	err = survey.AskOne(&survey.Select{
		Options: options,
		Message: "Do you want to provide a jumphost?",
	}, &choice, survey.Required)
	if err != nil {
		return err
	}

	switch choice {
	case OPTION_NONE:
		answers.JumpHost = ""
	case OPTION_SELECT:
		var jumphostKey string
		err = survey.AskOne(&survey.Select{
			Options: ssh.Config.Hosts,
			Message: "Choose jumphost:",
		}, &jumphostKey, survey.Required)
		if err != nil {
			return err
		}
		answers.JumpHost = jumphostKey
	}

	return err
}
