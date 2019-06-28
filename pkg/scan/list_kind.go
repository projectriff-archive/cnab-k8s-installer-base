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

package scan

import (
	"fmt"
	"runtime"
	"sort"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pivotal/go-ape/pkg/furl"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListKind(res string, baseDir string) ([]string, error) {
	fmt.Printf("Scanning %s\n", res)
	contents, err := furl.Read(res, baseDir)
	if err != nil {
		return nil, err
	}
	return ListKindFromContent(contents)
}

func ListKindFromContent(contents []byte) ([]string, error) {
	var err error
	types := map[string]bool{}

	docs := strings.Split(string(contents), "---\n")
	if runtime.GOOS == "windows" {
		// allow lines to end in LF or CRLF since either may occur
		d := strings.Split(string(contents), "---\r\n")
		if len(d) > len(docs) {
			docs = d
		}
	}
	for _, doc := range docs {
		if strings.TrimSpace(doc) != "" {
			tm := metav1.TypeMeta{}
			err = yaml.Unmarshal([]byte(doc), &tm)
			if err != nil {
				return nil, fmt.Errorf("error parsing content: %v", err)
			}
			if tm.Kind != "" {
				types[tm.Kind] = true
			}
		}
	}

	retVal := make([]string, len(types))
	i := 0
	for k := range types {
		retVal[i] = k
		i++
	}
	sort.Strings(retVal) // for deterministic order in tests
	return retVal, nil
}
