package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/maiko/sshed/ssh"
	"github.com/urfave/cli"
)

func (cmds *Commands) newTransferCommand() cli.Command {
	return cli.Command{
		Name:      "transfer",
		Usage:     "Transfers files to/from a host",
		ArgsUsage: "<key> <source_path> <destination_path>",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "upload, u",
				Usage: "Set this flag to upload files to the host",
			},
			cli.BoolFlag{
				Name:  "download, d",
				Usage: "Set this flag to download files from the host",
			},
		},
		Action: cmds.transferAction,
	}
}

func (cmds *Commands) transferAction(c *cli.Context) error {
	if c.NArg() < 3 {
		cli.ShowCommandHelp(c, "transfer")
		return errors.New("missing arguments: <key> <source_path> <destination_path>")
	}

	key := c.Args().Get(0)
	sourcePath := c.Args().Get(1)
	destinationPath := c.Args().Get(2)
	upload := c.Bool("upload")
	download := c.Bool("download")

	if upload == download {
		return errors.New("specify either --upload or --download")
	}

	return cmds.transferFile(c, key, sourcePath, destinationPath, upload)
}

func (cmds *Commands) transferFile(c *cli.Context, key, sourcePath, destinationPath string, upload bool) error {
	srv := ssh.Config.Get(key)
	scp := cmds.scp

	if srv == nil {
		return errors.New("host not found")
	}

	var scpCommand string
	options := ""
	sshpass := ""

	if srv.IdentityFile != "" {
		options += " -i " + quotePath(srv.IdentityFile)
	}
	if srv.Port != "" {
		options += " -P " + srv.Port
	}

	// Handle JumpHost with the possibility of a password
	if srv.JumpHost != "" {
		jumpHostConfig := ssh.Config.Get(srv.JumpHost)
		if jumpHostConfig == nil {
			return errors.New("jumphost not found")
		}
		if jumpHostConfig.Password() != "" {
			// Use sshpass for the JumpHost password
			options += " -o ProxyCommand=\"sshpass -p " + jumpHostConfig.Password() + " ssh -W %h:%p -p " + jumpHostConfig.Port + " " + jumpHostConfig.User + "@" + jumpHostConfig.Hostname + "\""
		} else {
			// No password for JumpHost, use standard ProxyJump
			options += " -o ProxyJump=" + srv.JumpHost
		}
	}

	if upload && !fileExists(sourcePath) {
		return errors.New("source file does not exist")
	}

	// Construct the scp command with sshpass if a password is provided for the destination host
	if srv.Password() != "" {
		sshpass = fmt.Sprintf("sshpass -p %s ", srv.Password())
	}

	if upload {
		fmt.Printf("Uploading %s to %s in %s\n", sourcePath, srv.Hostname, destinationPath)
		scpCommand = fmt.Sprintf("%s%s%s %s %s@%s:%s",
			sshpass, scp, options, quotePath(sourcePath), srv.User, srv.Hostname, quotePath(destinationPath))
	} else {
		fmt.Printf("Downloading %s from %s to %s\n", sourcePath, srv.Hostname, destinationPath)
		scpCommand = fmt.Sprintf("%s%s%s %s@%s:%s %s",
			sshpass, scp, options, srv.User, srv.Hostname, quotePath(sourcePath), quotePath(destinationPath))
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", scpCommand)
	} else {
		cmd = exec.Command("sh", "-c", scpCommand)
	}

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		return err
	}

	fmt.Println("Done!")
	return nil
}

func quotePath(path string) string {
	// Prevent path traversal by cleaning the path and ensuring it does not start with a dash
	cleanPath := filepath.Clean(path)
	if strings.HasPrefix(cleanPath, "-") {
		cleanPath = "./" + cleanPath
	}
	// Escape single quotes in the path
	escapedPath := strings.Replace(cleanPath, "'", "'\"'\"'", -1)
	return fmt.Sprintf("'%s'", escapedPath)
}

func fileExists(path string) bool {
	// Use filepath.Clean to prevent path traversal attacks
	cleanPath := filepath.Clean(path)
	_, err := os.Stat(cleanPath)
	return !os.IsNotExist(err)
}
