//go:build !(darwin && arm64) && !(darwin && amd64) && !(windows && amd64) && !(linux && amd64)

package ffmpeg

// empty
var binary []byte
