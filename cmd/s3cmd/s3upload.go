package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func UploadBinLog(host, interP string) {

	S3EndPoint := os.Getenv("S3_ENDPOINT")
	S3AccessKey := os.Getenv("S3_ACCESSKEY")
	S3SecretKey := os.Getenv("S3_SECRETKEY")
	S3Bucket := os.Getenv("S3_BUCKET")
	cluster := os.Getenv("CLUSTER_NAME")
	var err error
	var minioClient *minio.Client
	if minioClient, err = minio.New(strings.TrimPrefix(strings.TrimPrefix(S3EndPoint, "https://"), "http://"), &minio.Options{
		Creds:  credentials.NewStaticV4(S3AccessKey, S3SecretKey, ""),
		Secure: strings.HasPrefix(S3Bucket, "https"),
		// Region: "pek3b",
	}); err != nil {
		log.Fatalf("error in minio %v", err)
		return
	}
	// mysqlbinlog dump and put load
	runBinlogDump(host, "3306", "root", interP)
	files, _ := ioutil.ReadDir("/tmp/")
	ctx := context.TODO()
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "mysql-bin") {
			fh, err := os.Open("/tmp/" + file.Name())
			if err != nil {
				log.Fatalf("failed to open file %v", err)
				return
			}
			log.Printf("put %s \n", cluster+"/"+filepath.Base(file.Name()))
			upinfo, err := minioClient.PutObject(ctx, S3Bucket,
				cluster+"-binlog/"+filepath.Base(file.Name()), fh, -1, minio.PutObjectOptions{})
			if err != nil {
				log.Fatalf("failed to upload file to S3 %v", err)
				return
			}
			log.Println("S3 upload", "upinfo", upinfo)
		}
	}

}

func runBinlogDump(host, port, user, pass string) error {
	cmdstr := fmt.Sprintf(`cd /tmp/;mysql -uroot -p%s -h%s -N -e "SHOW BINARY LOGS" | awk '{print "mysqlbinlog --read-from-remote-server --raw --host=%s --port=3306 --user=root --password=%s --raw", $1}'|bash`,
		pass, host, host, pass)

	cmd := exec.Command("bash", "-c", cmdstr)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
