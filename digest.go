package image

import "github.com/opencontainers/go-digest"

// Digest allows unique identification of an image, both thru its contents (like Id does)
// but also by its name. An image with a given binary contents (say Id = abcd) can thus be
// referenced as two distinct Digests (d1 = 3ebf when tagged foo/bar and d2 = 4a43 when
// tagged wiz/bot). Still, Digests are more "secure" than named tags, because the latter
// can be updated. Digests, on the other hand, will change if the contents of the image
// change.
type Digest struct {
	dig digest.Digest
}

func NewDigest(dig string) Digest {
	return Digest{digest.Digest(dig)}
}

var EmptyDigest Digest

func init() {
	EmptyDigest = Digest{""}
}

func (d Digest) String() string {
	return string(d.dig)
}
