package docker

import (
	"cnab-k8s-installer-base/pkg/image"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"io"
	"strings"
)

type dclient struct {
	cli *client.Client
	ctx context.Context
}

type DClient interface {
	Pull(ref string) (image.Name, image.Id, error)
	Relocate(fromRef, toRef string) error
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

func NewDockerClient() (*dclient, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	ctx := context.TODO()

	return &dclient{
		cli: cli,
		ctx: ctx,
	}, nil
}

func (dc *dclient) Pull(ref string) (image.Name, image.Id, error) {
	fmt.Printf("Pulling image %s...\n", ref)
	var err error
	n, err := image.NewName(ref)
	if err != nil {
		return image.EmptyName, image.EmptyId, err
	}
	events, err := dc.cli.ImagePull(dc.ctx, n.String(), types.ImagePullOptions{})
	if err != nil {
		return image.EmptyName, image.EmptyId, err
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
			if strings.Contains(event.Status, "Digest") {
				digest = event.Status
			}
		}
	}
	digest = strings.TrimSpace(strings.SplitN(digest, ":", 2)[1])
	fmt.Println(event.Status)
	id, err := dc.getImageId(digest)
	if err != nil {
		return image.EmptyName, image.EmptyId, err
	}
	name, err := image.NewName(ref)
	if err != nil {
		return image.EmptyName, image.EmptyId, err
	}
	return name, id, nil
}

func (dc *dclient) Relocate(fromRef, toRef string) error {
	_, id, err := dc.Pull(fromRef)
	if err != nil {
		return err
	}

	toName, err := image.NewName(toRef)
	if err != nil {
		return err
	}

	tag, err := dc.Tag(id, toName)
	if err != nil {
		return err
	}

	err = dc.Push(tag)
	if err != nil {
		return err
	}

	return nil
}

func (dc *dclient) Tag(id image.Id, name image.Name) (image.Name, error) {
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

func (dc *dclient) Push(name image.Name) error {
	fmt.Println("Pushing", name.String())
	events, err := dc.cli.ImagePush(dc.ctx, name.String(), types.ImagePushOptions{RegistryAuth:"foo"})
	if err != nil {
		return err
	}
	d := json.NewDecoder(events)
	var event *Event
	for {
		if err := d.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
		}
	}
	fmt.Println(event.Status)

	return nil
}

func (dc *dclient) getImageId(digest string) (image.Id, error) {
	images, err := dc.cli.ImageList(dc.ctx, types.ImageListOptions{})
	if err != nil {
		return image.EmptyId, err
	}

	for _, img := range images {
		if arrayContainsSubstring(img.RepoDigests, digest) {
			id := image.NewId(img.ID)
			return id, nil
		}
	}
	return image.EmptyId, errors.New(fmt.Sprintf("No image found for digest %s\n", digest))
}

func arrayContainsSubstring(digests []string, digest string) bool {
	for _, str := range digests {
		if strings.Contains(str, digest) {
			return true
		}
	}
	return false
}
