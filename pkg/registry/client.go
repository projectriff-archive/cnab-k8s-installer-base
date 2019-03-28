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

package registry

import (
	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/registry"
	log "github.com/sirupsen/logrus"
)

type client struct {
	registryClient registry.Client
}

type Client interface {
	Relocate(fromRef, toRef string) (image.Name, error)
}

func NewClient() (Client, error) {
	return &client{
		registryClient: registry.NewRegistryClient(),
	}, nil
}

func (dc *client) Relocate(fromRef, toRef string) (image.Name, error) {
	log.Debugf("Relocating image from %s to %s\n", fromRef, toRef)
	from, err := image.NewName(fromRef)
	if err != nil {
		return image.EmptyName, err
	}

	to, err := image.NewName(toRef)
	if err != nil {
		return image.EmptyName, err
	}

	dig, err := dc.registryClient.Copy(from, to)
	if err != nil {
		return image.EmptyName, err
	}

	return to.WithDigest(dig)
}
