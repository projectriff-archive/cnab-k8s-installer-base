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

package kab

import (
	"cnab-k8s-installer-base/pkg/client/clientset/versioned"
	"cnab-k8s-installer-base/pkg/registry"
	"cnab-k8s-installer-base/pkg/kustomize"
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
	coreClient     kubernetes.Interface
	extClient      apiext.Interface
	kabClient      versioned.Interface
	registryClient registry.Client
	kustomizer     kustomize.Kustomizer
}

func NewKnbClient(core kubernetes.Interface, ext apiext.Interface, kab versioned.Interface, registryClient registry.Client,
	kustomizer kustomize.Kustomizer) *Client {
	return &Client{
		coreClient:     core,
		extClient:      ext,
		kabClient:      kab,
		registryClient: registryClient,
		kustomizer:     kustomizer,
	}
}
