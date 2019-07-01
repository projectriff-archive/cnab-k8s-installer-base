/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/projectriff/cnab-k8s-installer-base/pkg/apis/kab/v1alpha1"

	log "github.com/sirupsen/logrus"
)

const (
	CNAB_ACTION_ENV_VAR            = "CNAB_ACTION"
	CNAB_INSTALLATION_NAME_ENV_VAR = "CNAB_INSTALLATION_NAME"
	MANIFEST_FILE_ENV_VAR          = "MANIFEST_FILE"
	NODE_PORT_ENV_VAR              = "NODE_PORT"
	LOG_LEVEL_ENV_VAR              = "LOG_LEVEL"
)

func main() {

	setupLogging()

	path := getEnv(MANIFEST_FILE_ENV_VAR)
	if path == "" {
		// revert after duffle fixes the export parameter issue
		// https://github.com/deislabs/duffle/issues/753
		path = "/cnab/app/kab/manifest.yaml"
	}
	action := getEnv(CNAB_ACTION_ENV_VAR)
	action = strings.ToLower(action)
	log.Debugf("performing action: %s, manifest file: %s", action, path)
	switch action {
	case "install":
		install(path)
	// TODO restore uninstall
	// case "uninstall":
	// 	uninstall()
	// case "upgrade":
	default:
		log.Fatalf("unknown action '%s'. please set CNAB_ACTION environment variable", action)
	}
}

func install(path string) {
	manifest, err := v1alpha1.NewManifest(path)
	if err != nil {
		_, err = fmt.Fprintf(os.Stderr, "error while reading from %s: %v", path, err)
		os.Exit(1)
	}
	for _, resource := range manifest.Spec.Resources {
		cmd := exec.Command("kapp", "deploy", "-y", "-a", resource.Name, "-f", resource.Path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			_, err = fmt.Fprintf(os.Stderr, "error while applying %s: %v", resource.Name, err)
			os.Exit(1)
		}
	}
}

func setupLogging() {
	log.SetOutput(os.Stdout)
	log.SetLevel(getLogLevel())
}

func getLogLevel() log.Level {
	requestedLevel := getEnv(LOG_LEVEL_ENV_VAR)
	if requestedLevel == "" {
		return log.InfoLevel
	}
	level, err := log.ParseLevel(requestedLevel)
	if err != nil {
		log.Fatalf("Unknown log level %s", requestedLevel)
	}
	return level
}

// duffle sets the env value to "<nil>", so restore normal behavior
func getEnv(env_var string) string {
	val := os.Getenv(env_var)
	if strings.Contains(val, "nil") {
		return ""
	}
	return val
}
