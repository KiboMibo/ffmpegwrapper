package ffmpegwrapper

import (
	"testing"
	"fmt"
)

func Test(t *testing.T) {
	m, err := NewMediaFile("./test/test2.mov")
	if err != nil {
		t.Errorf("Failed get file info : %s", err)

		return
	}

	ffmpParams := "-c:v libx264 -crf 20 -vf scale=1280:720"
	if out, err := m.Convert("pre_test.mp4", ffmpParams); err != nil{
		t.Errorf("Failed convert file : %s", err)
	} else {
		for msg := range out {
		   fmt.Println(msg)
		}
	}

}
