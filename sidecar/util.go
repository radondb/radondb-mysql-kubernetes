/*
Copyright 2021 RadonDB.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sidecar

import (
	"io"
	"os"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/radondb/radondb-mysql-kubernetes/utils"
)

var (
	log = logf.Log.WithName("sidecar")

	// MysqlServerIDOffset represents the offset with which all server ids are shifted from 0
	mysqlServerIDOffset = 100

	// configPath is the mysql configs path.
	configPath = utils.ConfVolumeMountPath

	// clientConfPath is the client.cnf path.
	clientConfPath = utils.ConfClientPath

	// extraConfPath is the mysql extra configs path.
	extraConfPath = utils.ConfVolumeMountPath + "/conf.d"

	// configMapPath is the mounted configmap.
	configMapPath = utils.ConfMapVolumeMountPath

	// dataPath is the mysql data path.
	dataPath = utils.DataVolumeMountPath

	// scriptsPath is the scripts path used for xenon.
	scriptsPath = utils.ScriptsVolumeMountPath

	// sysPath is the linux kernel path used for install tokudb.
	sysPath = utils.SysVolumeMountPath

	// xenonPath is the xenon configs path.
	xenonPath = utils.XenonVolumeMountPath

	// initFilePath is the init files path for mysql.
	initFilePath = utils.InitFileVolumeMountPath
)

// copyFile the src file to dst.
// nolint: gosec
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if err1 := in.Close(); err1 != nil {
			log.Error(err1, "failed to close source file", "src_file", src)
		}
	}()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err1 := out.Close(); err1 != nil {
			log.Error(err1, "failed to close destination file", "dest_file", dst)
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return nil
}

// getEnvValue get environment variable by the key.
func getEnvValue(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		log.Info("environment is not set", "key", key)
	}

	return value
}

// checkIfPathExists check if the path exists.
func checkIfPathExists(path string) (bool, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		log.Error(err, "failed to open file", "file", path)
		return false, err
	}

	err = f.Close()
	return true, err
}
