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
