//go:build linux && amd64

package ffmpeg

import _ "embed"

//go:embed bin/ffmpeg-linux-amd64
var binary []byte
