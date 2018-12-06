package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/livepeer/go-livepeer/monitor"
	"github.com/livepeer/lpms/ffmpeg"
	"github.com/livepeer/lpms/transcoder"
)

type Transcoder interface {
	Transcode(fname string, profiles []ffmpeg.VideoProfile) ([][]byte, error)
}

type LocalTranscoder struct {
	workDir string
}

func (lt *LocalTranscoder) Transcode(fname string, profiles []ffmpeg.VideoProfile) ([][]byte, error) {
	tr := transcoder.NewFFMpegSegmentTranscoder(profiles, lt.workDir)
	mid, profile, seqNo, err := parseURI(fname)
	if monitor.Enabled && err == nil {
		monitor.LogSegmentTranscodeStarting(seqNo, mid, profile)
	}
	data, err := tr.Transcode(fname)
	if monitor.Enabled && err == nil {
		monitor.LogSegmentTranscodeEnded(seqNo, mid, profile)
	}
	return data, err
}

func NewLocalTranscoder(workDir string) Transcoder {
	return &LocalTranscoder{workDir: workDir}
}

func parseURI(uri string) (string, string, uint64, error) {
	var mid, profile string
	var seqNo uint64
	parts := strings.Split(uri, "/")
	if len(parts) < 3 {
		return mid, profile, seqNo, fmt.Errorf("BadURI")
	}
	mid = parts[len(parts)-3]
	profile = parts[len(parts)-2]
	parts = strings.Split(parts[len(parts)-1], ".")
	seqNo, err := strconv.ParseUint(parts[0], 10, 64)
	return mid, profile, seqNo, err
}
