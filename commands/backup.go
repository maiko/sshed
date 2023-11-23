// commands/backup.go

package commands

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli"
)

func (cmds *Commands) newBackupCommand() cli.Command {
	return cli.Command{
		Name:   "backup",
		Usage:  "Backs up SSH configuration and keychain into a .tgz file",
		Action: cmds.backupAction,
	}
}

func (cmds *Commands) backupAction(ctx *cli.Context) error {
	// Get the paths for SSH config and keychain from the context
	sshConfigPath := ctx.String("config")
	keychainPath := ctx.String("keychain")
	backupDir := ctx.String("backup-dir")

	// Validate that the paths are not empty
	if sshConfigPath == "" || keychainPath == "" || backupDir == "" {
		return errors.New("ssh-config-path, keychain-path, and backup-dir must be provided")
	}

	// Create a temporary directory for the backup
	tempDir, err := os.MkdirTemp("", "sshed_backup_")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir) // clean up

	// Define backup file paths in the temp directory
	sshConfigBackupPath := filepath.Join(tempDir, "ssh_config_backup")
	keychainBackupPath := filepath.Join(tempDir, "keychain_backup")

	// Backup the SSH config file
	if err := copyFile(sshConfigPath, sshConfigBackupPath); err != nil {
		return err
	}

	// Backup the keychain database
	if err := copyFile(keychainPath, keychainBackupPath); err != nil {
		return err
	}

	// Create a tgz file containing the backed up files
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	tgzFilename := fmt.Sprintf("sshed_backup_%s.tgz", timestamp)
	tgzPath := filepath.Join(backupDir, tgzFilename)
	if err := createTgz(tempDir, tgzPath); err != nil {
		return err
	}

	return nil
}

// createTgz creates a .tgz file at tgzPath containing the contents of backupDir
func createTgz(backupDir, tgzPath string) error {
	tgzFile, err := os.Create(tgzPath)
	if err != nil {
		return err
	}
	defer tgzFile.Close()

	gzipWriter := gzip.NewWriter(tgzFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Add files to the tar archive
	files := []string{"ssh_config_backup", "keychain_backup"}
	for _, file := range files {
		filePath := filepath.Join(backupDir, file)
		if err := addFileToTar(tarWriter, filePath); err != nil {
			return err
		}
	}

	return nil
}

// addFileToTar adds a file to the tar archive
func addFileToTar(tw *tar.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	// Create a header based on the file info
	header := &tar.Header{
		Name:    filepath.Base(filePath),
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	// Copy the file data to the tar writer
	if _, err := io.Copy(tw, file); err != nil {
		return err
	}

	return nil
}

// copyFile copies a file from src to dst and preserves file permissions
func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return errors.New(src + " is not a regular file")
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sourceFileStat.Mode())
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
