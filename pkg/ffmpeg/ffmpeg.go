package ffmpeg

import (
	"bytes"

	"github.com/disintegration/imaging"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func GetCoverImage(videoPath, coverPath string) error {
	buf := bytes.NewBuffer(nil)
	err := ffmpeg.Input(videoPath).Filter("select", ffmpeg.Args{"gte(n,0)"}).
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf, nil).
		Run()
	if err != nil {
		return err
	}
	img, err := imaging.Decode(buf)
	if err != nil {
		return err
	}
	return imaging.Save(img, coverPath)
}
