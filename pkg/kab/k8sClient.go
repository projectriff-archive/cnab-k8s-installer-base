/*
 * Copyright 2019 The original author or authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package kab

import (
	"cnab-k8s-installer-base/pkg/client/clientset/versioned"
	"cnab-k8s-installer-base/pkg/docker"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"time"
)

const (
	maxRetries             = 18 // the sum of all retries would add up to 1 minute
	minRetryInterval       = 100 * time.Millisecond
	exponentialBackoffBase = 1.3
)

type Client struct {
	coreClient   *kubernetes.Clientset
	extClient    *apiext.Clientset
	kabClient    *versioned.Clientset
	dockerClient *docker.Client
}

func NewKnbClient(core *kubernetes.Clientset, ext *apiext.Clientset, kab *versioned.Clientset, dockerClient *docker.Client) *Client {
	return &Client{
		coreClient:   core,
		extClient:    ext,
		kabClient:    kab,
		dockerClient: dockerClient,
	}
}
