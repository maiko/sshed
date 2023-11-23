// commands/restore.go

package commands

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"github.com/urfave/cli"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func (cmds *Commands) newRestoreCommand() cli.Command {
	return cli.Command{
		Name:   "restore",
		Usage:  "Restores SSH configuration and keychain from a backup",
		Action: cmds.restoreAction,
	}
}

func (cmds *Commands) restoreAction(ctx *cli.Context) error {
	// Validate that the backup file exists
	backupFilePath := ctx.String("backup-file")
	if backupFilePath == "" {
		return errors.New("backup file path must be provided")
	}
	if _, err := os.Stat(backupFilePath); os.IsNotExist(err) {
		return errors.New("backup file does not exist")
	}

	// Create a temporary directory for the restore process
	tempRestoreDir, err := ioutil.TempDir("", "sshed_restore_")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempRestoreDir)

	// Extract the tgz file to the temporary directory
	if err := extractTgz(backupFilePath, tempRestoreDir); err != nil {
		return err
	}

	// Backup existing SSH config and keychain files before overwriting
	sshConfigPath := ctx.String("ssh-config-path")
	keychainPath := ctx.String("keychain-path")
	if err := backupExistingFiles([]string{sshConfigPath, keychainPath}); err != nil {
		return err
	}

	// Restore the SSH config file
	if err := moveFile(filepath.Join(tempRestoreDir, "ssh_config"), sshConfigPath); err != nil {
		return err
	}

	// Restore the keychain database
	if err := moveFile(filepath.Join(tempRestoreDir, "keychain.db"), keychainPath); err != nil {
		return err
	}

	return nil
}

func extractTgz(tgzPath, restoreDir string) error {
	tgzFile, err := os.Open(tgzPath)
	if err != nil {
		return err
	}
	defer tgzFile.Close()

	gzipReader, err := gzip.NewReader(tgzFile)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	// Extract files from the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(restoreDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

// backupExistingFiles creates backups of existing files before they are overwritten
func backupExistingFiles(paths []string) error {
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			backupPath := path + ".bak"
			if err := moveFile(path, backupPath); err != nil {
				return err
			}
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// moveFile moves a file from src to dst, effectively renaming it
func moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err != nil {
		return err
	}
	return nil
}
