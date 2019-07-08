package ffmpegwrapper

import (
	"fmt"
	"strings"
	"testing"
)

func TestWrapper(t *testing.T) {
	m, err := NewMediaFile("./test/test2.mp4")
	if err != nil {
		t.Errorf("Failed get file info : %s", err)

		return
	}

	ffmpParams := "-ss 00:01:30 -c:v libx264 -c:a copy -crf 20 -vf scale=740:-1"
	if out, err := m.Convert("pre_test.mp4", strings.Split(ffmpParams, " ")); err != nil {
		t.Errorf("Failed convert file : %s", err)
	} else {
		for msg := range out {
			fmt.Println(msg)
		}
	}

}
