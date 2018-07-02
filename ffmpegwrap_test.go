package goFFmpegWrapper

import (
	"testing"
	"fmt"
)

func Test(t *testing.T) {
	m, err := NewMediaFile("./test/test.MOV")
	if err != nil {
		t.Errorf("Failed get file info : %s", err)

		return
	}

	ffmpParams := "-crf 20 -bufsize 4096k -vf scale=1280:800:force_original_aspect_ratio=decrease"
	if out, err := m.Convert("pre_test.mov", ffmpParams); err != nil{
		t.Errorf("Failed convert file : %s", err)
	} else {
		for msg := range out {
		   fmt.Println(msg)
		}
	}
}
