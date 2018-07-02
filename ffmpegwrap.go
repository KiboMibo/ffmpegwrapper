package goFFmpegWrapper

import (
	"os/exec"
	"errors"
	"fmt"
	"os"
	"encoding/json"
	"path/filepath"
	"strings"
	"bufio"
)

type MediaFile struct {
	Filename string
	Info     *Metadata
}


// AnalyzeMetadata calls ffprobe on the file and parses its output to MedisaFile/Info structure.
func (m *MediaFile) AnalyzeMetadata() (err error) {
	cmdName, err := exec.LookPath("ffprobe")
	if err != nil {
		return errors.New("ffprobe is not installed")
	}
	cmdArgs := fmt.Sprintf("-show_format -show_streams -pretty -print_format json -hide_banner -i %s", m.Filename)
	out, err := exec.Command(cmdName, strings.Split(cmdArgs, " ")...).Output()
	if err != nil {
		fmt.Println(err)
		return errors.New("Failed read metadata")
	}
	if err := json.Unmarshal(out, m.Info); err != nil {
		return errors.New("Failed unmarshal from JSON")
	}
	return
}
// Convert checks that outFileName is abs path or put the output file in the same dir that original
func (m *MediaFile) Convert(outFileName string, args string) (chan string, error){
	cmdName, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, errors.New("ffmpeg is not installed")
	}
	// Check that out path is exist
	outPath := filepath.Dir(outFileName)
	_, err = os.Stat(outPath)
	if os.IsNotExist(err) || !filepath.IsAbs(outPath){
		outPath = filepath.Join(filepath.Dir(m.Filename), filepath.Base(outFileName))
	} else {
		outPath = outFileName
	}
	// Set ffmpeg defaults + user args
	// default arg is answer yes to rewrite file
	// show only errors and converting status
	var cmdArgs = []string{"-y", "-v", "error", "-stats", "-i", m.Filename}
	if len(args) > 0 {
		cmdArgs = append(cmdArgs, strings.Split(args, " ")...)
	}
	cmdArgs = append(cmdArgs, outPath)

	cmd := exec.Command(cmdName, cmdArgs...)

	cmdReader, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		return nil, err
	}
	// make scanner to read changing output by words
	scanner := bufio.NewScanner(cmdReader)
	scanner.Split(bufio.ScanWords)
	outChan := make(chan string)

	go func() {
		defer func(){
			close(outChan)
		}()

		var fullVal string
		var fullString []string

		for scanner.Scan() {
			w := scanner.Text()
			// Coz output readed by changing words, so we need join words to get normal status string
			if strings.Contains(w, "="){
				tmpValString := strings.Split(w, "=")
				if len(tmpValString) > 1 && tmpValString[1] == ""{
					fullVal = w
					continue
				}
				fullVal = w
			} else {
				fullVal = fullVal + w
			}
			// Finally concatenate status string
			fullString = append(fullString, fullVal)
			if len(fullString) >= 7 {
				outChan <- strings.Join(fullString, " ")
				fullString = nil
			}
		}
	}()

	go func(){
		err = cmd.Start()
		if err != nil {
			return
		}
		outChan <- "Converting started"
		err = cmd.Wait()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
			return
		}
	}()

	return outChan, nil
}


// getExistingPath ensures a path actually exists, and returns an existing absolute path or an error.
func getExistingPath(path string) (existingPath string, err error) {
	// check root exists or pwd+root exists
	if filepath.IsAbs(path) {
		existingPath = path
	} else {
		pwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		existingPath = filepath.Join(pwd, path)
	}
	// check root exists
	_, err = os.Stat(existingPath)
	return
}

// NewMediaFile initializes a MediaFile and parses its metadata with ffprobe.
func NewMediaFile(filename string) (mf *MediaFile, err error) {
	filename, err = getExistingPath(filename)
	if os.IsNotExist(err) {
		return nil, err
	}

	var meta Metadata
	mf = &MediaFile{Filename: filename, Info: &meta}
	err = mf.AnalyzeMetadata()
	return
}