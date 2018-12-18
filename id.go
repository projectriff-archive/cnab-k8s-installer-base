package image

import (
	"github.com/opencontainers/go-digest"
	"strings"
)

// Id is an image id, which happens to be represented as a digest string. An image id
// is based on the binary contents of an image, but not with its name (see Digest).
type Id struct {
	dig digest.Digest
}

var EmptyId Id

func init() {
	EmptyId = Id{""}
}

func NewId(id string) Id {
	return Id{digest.Digest(id)}
}

func (id Id) String() string {
	return string(id.dig)
}

// Filename returns a filesystem-friendly name for this image id.
// Filenames in Windows cannot contain ":".
func (id Id) Filename() string {
	return strings.Replace(id.String(), ":", "-", -1)
}
