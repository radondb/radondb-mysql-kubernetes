package sidecar

import (
	"os"
	"os/exec"
	"strings"
)

// RunTakeBackupCommand starts a backup command
func RunTakeBackupCommand(cfg *Config, name string) error {
	log.Info("backup mysql", "name", name)
	// cfg->XtrabackupArgs()
	xtrabackup := exec.Command(xtrabackupCommand, cfg.XtrabackupArgs()...)

	var err error

	xcloud := exec.Command(xcloudCommand, cfg.XCloudArgs()...)
	log.Info("xargs ", "xargs", strings.Join(cfg.XCloudArgs(), " "))
	if xcloud.Stdin, err = xtrabackup.StdoutPipe(); err != nil {
		log.Error(err, "failed to pipline")
		return err
	}
	xtrabackup.Stderr = os.Stderr
	xcloud.Stderr = os.Stderr

	if err := xtrabackup.Start(); err != nil {
		log.Error(err, "failed to start xtrabackup command")

		return err
	}
	if err := xcloud.Start(); err != nil {
		log.Error(err, "fail start xcloud ")
		return err
	}
	// copy stardout to xcloud
	if err := xtrabackup.Wait(); err != nil {
		log.Error(err, "failed waiting for xtrabackup to finish")

		return err
	}
	if err := xcloud.Wait(); err != nil {
		log.Error(err, "failed waiting for xcloud to finish")

		return err
	}

	return nil
}
