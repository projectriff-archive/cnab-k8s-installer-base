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

package docker

import (
	"cnab-k8s-installer-base/pkg/image"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"io"
	"strings"
)

type Client struct {
	cli *client.Client
	ctx context.Context
}

type DClient interface {
	Pull(ref string) (image.Name, image.Digest, error)
	Relocate(fromRef, toRef string) (image.Name, error)
	Tag(id image.Digest, name image.Name) (image.Name, error)
	Push(name image.Name) (image.Name, error)
}

type Event struct {
	Status         string `json:"status"`
	Error          string `json:"error"`
	Progress       string `json:"progress"`
	ProgressDetail struct {
		Current int `json:"current"`
		Total   int `json:"total"`
	} `json:"progressDetail"`
}

func NewDockerClient() (*Client, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	ctx := context.TODO()

	return &Client{
		cli: cli,
		ctx: ctx,
	}, nil
}

func (dc *Client) Pull(ref string) (image.Name, image.Digest, error) {
	log.Debugf("Pulling image %s...\n", ref)
	var err error
	n, err := image.NewName(ref)
	if err != nil {
		return image.EmptyName, image.EmptyDigest, err
	}
	events, err := dc.cli.ImagePull(dc.ctx, n.String(), types.ImagePullOptions{})
	if err != nil {
		return image.EmptyName, image.EmptyDigest, err
	}
	d := json.NewDecoder(events)
	var digest string
	var event *Event
	for {
		if err := d.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
		} else {
			if strings.Contains(strings.ToUpper(event.Status), "DIGEST") {
				digest = event.Status
			}
		}
	}
	digest, err = extractDigest(digest)
	if err != nil {
		return image.EmptyName, image.EmptyDigest, err
	}
	log.Debugln(event.Status, "DIGEST:", digest)
	id, err := dc.getImageId(digest)
	if err != nil {
		return image.EmptyName, image.EmptyDigest, err
	}
	name, err := image.NewName(ref)
	if err != nil {
		return image.EmptyName, image.EmptyDigest, err
	}
	return name, id, nil
}

func (dc *Client) Relocate(fromRef, toRef string) (image.Name, error) {
	_, id, err := dc.Pull(fromRef)
	if err != nil {
		return image.EmptyName, err
	}

	toName, err := image.NewName(toRef)
	if err != nil {
		return image.EmptyName, err
	}

	tag, err := dc.Tag(id, toName)
	if err != nil {
		return image.EmptyName, err
	}

	digestedRef, err := dc.Push(tag)
	if err != nil {
		return image.EmptyName, err
	}

	return digestedRef, nil
}

func (dc *Client) Tag(id image.Digest, name image.Name) (image.Name, error) {
	var err error
	tag, err := image.NewName(name.WithoutDigest())
	if name.Tag() != "" {
		tag, err = tag.WithTag(name.Tag())
		if err != nil {
			return image.EmptyName, err
		}
	}
	err = dc.cli.ImageTag(dc.ctx, id.String(), tag.String())
	if err != nil {
		return image.EmptyName, err
	}
	return tag, nil
}

func (dc *Client) Push(name image.Name) (image.Name, error) {
	log.Debugf("Pushing %s...", name.String())
	events, err := dc.cli.ImagePush(dc.ctx, name.String(), types.ImagePushOptions{RegistryAuth:"foo"})
	if err != nil {
		return image.EmptyName, err
	}
	d := json.NewDecoder(events)
	var digest string
	var event *Event
	for {
		if err := d.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
		} else {
			log.Traceln(event.Status)
			if strings.Contains(strings.ToUpper(event.Status), "DIGEST") {
				digest = event.Status
			}
		}
	}
	log.Debugln("done")
	digest, err = extractDigest(digest)
	newName, err := name.WithDigest(image.NewDigest(digest))
	log.Debugln("digest image reference:", newName)
	return newName, nil
}

func extractDigest(str string) (string, error) {
	arr := strings.Fields(str)
	for _, str := range arr {
		if strings.HasPrefix(str, "sha256") {
			return str, nil
		}
	}
	return "", errors.New(fmt.Sprintf("cannot extract digest from: %s", str))
}

func (dc *Client) getImageId(digest string) (image.Digest, error) {
	images, err := dc.cli.ImageList(dc.ctx, types.ImageListOptions{})
	if err != nil {
		return image.EmptyDigest, err
	}

	for _, img := range images {
		if arrayContainsSubstring(img.RepoDigests, digest) {
			id := image.NewDigest(img.ID)
			return id, nil
		}
	}
	return image.EmptyDigest, errors.New(fmt.Sprintf("No image found for digest %s\n", digest))
}

func arrayContainsSubstring(digests []string, digest string) bool {
	for _, str := range digests {
		if strings.Contains(str, digest) {
			return true
		}
	}
	return false
}
