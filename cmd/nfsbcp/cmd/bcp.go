package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/radondb/radondb-mysql-kubernetes/cmd/nfsbcp/api"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: `Create a nfs backup for RadonDB MySQL cluster`,
	Run: func(cmd *cobra.Command, args []string) {
		hostPath := cmd.Flag("hostpath").Value.String()
		cluster := cmd.Flag("cluster").Value.String()
		hostName := cmd.Flag("hostname").Value.String()
		capacity := cmd.Flag("capacity").Value.String()
		backupImage := cmd.Flag("backupImage").Value.String()
		nfsServerImage := cmd.Flag("nfsServerImage").Value.String()

		if err := api.Create(hostPath, cluster, hostName, capacity, backupImage, nfsServerImage); err != nil {
			log.Error(err, " run command failed")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().StringP("hostpath", "p", "", "The local storage path of nfs.")
	backupCmd.Flags().StringP("cluster", "c", "", "The cluster name to backup.")
	backupCmd.Flags().StringP("hostname", "n", "", "The host for which to take backup.")
	backupCmd.Flags().StringP("capacity", "s", "30", "The capacity of nfs server.")
	backupCmd.Flags().StringP("backupImage", "b", "radondb/mysql57-sidecar:v2.2.0", "The image of backup.")
	backupCmd.Flags().StringP("nfsServerImage", "i", "k8s.gcr.io/volume-nfs:0.8", "The image of nfs server.")
	backupCmd.MarkFlagRequired("hostpath")
	backupCmd.MarkFlagRequired("cluster")
}
