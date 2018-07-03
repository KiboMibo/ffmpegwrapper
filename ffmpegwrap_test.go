package ffmpegwrapper

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

	ffmpParams := "-ss 00:00:01 -vframes 1 -q:v 2 -vf scale=1280:-1"
	if out, err := m.Convert("pre_test.jpg", ffmpParams); err != nil{
		t.Errorf("Failed convert file : %s", err)
	} else {
		for msg := range out {
		   fmt.Println(msg)
		}
	}

}
