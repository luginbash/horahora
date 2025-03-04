package dashutils

import (
	"fmt"
	"os/exec"
	"path/filepath"
	log "github.com/sirupsen/logrus"
)

type H264Transcoder struct {
}

func (h H264Transcoder) TranscodeAndGenerateManifest(path string, local bool) (*DASHVideo, error) {
	// var encodeArgs []string
	// switch local {
	// case true:
	// 	// make the encoding really fast so we don't have to wait 90 minutes for integration tests
	// 	// I don't really understand the difference between the speed and deadline args, but the documentation implies
	// 	// they're separate
	// 	encodeArgs = []string{path, "-speed 16 -deadline realtime -r 1 -crf 63 -t 10"}
	// case false:
	// 	// -r 24 -deadline realtime -cpu-used 1
	// 	encodeArgs = []string{path, "-r 24 -deadline good -cpu-used 2"}
	// }

	cmd := exec.Command("/horahora/videoservice/scripts/transcode.sh", []string{path}...)
	out, err := cmd.CombinedOutput()
	if err != nil {
	log.Errorf("%s", out)
	return nil, err
	}

	// At this point it's been transcoded, so generate the DASH manifest
	cmd = exec.Command("/horahora/videoservice/scripts/manifest.sh", []string{filepath.Base(path)}...)
	out, err = cmd.CombinedOutput()
	if err != nil {
	log.Errorf("%s", out)
	return nil, fmt.Errorf("failed to generate dash manifest. Err: %s", err)
	}

	var fileList []string

	generatedFiles, err := filepath.Glob(fmt.Sprintf("%s_*", path))
	if err != nil {
	return nil, err
	}

	log.Infof("Generated files: %s", generatedFiles)

	for _, fileName := range generatedFiles {
	fileList = append(fileList, fileName)
	}

	manifest := fmt.Sprintf("%s.mpd", path)

	return &DASHVideo{
	ManifestPath:     &manifest,
	QualityMap:       fileList,
	OriginalFilePath: path,
	ThumbnailPath:    path + ".jpg",
	}, nil
}
