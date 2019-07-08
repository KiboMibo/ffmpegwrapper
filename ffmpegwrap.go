package ffmpegwrapper

import (
	"bufio"
	"bytes"
	"encoding/json"
	"golang.org/x/xerrors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

type MediaFile struct {
	Filename string
	Info     *Metadata
}

// AnalyzeMetadata calls ffprobe on the file and parses its output to MedisaFile/Info structure.
func (m *MediaFile) AnalyzeMetadata() (err error) {
	cmdName, err := exec.LookPath("ffprobe")
	if err != nil {
		err = xerrors.Errorf("FFprobe: %w", err)
		return
	}
	if strings.ToLower(filepath.Ext(m.Filename)) == ".heic" {
		return xerrors.New("Found heic image no metadata with ffprobe will exported")
	}
	cmdArgs := []string{"-show_format", "-show_streams", "-pretty", "-print_format", "json", "-hide_banner"}
	cmdArgs = append(cmdArgs, []string{"-i", m.Filename}...)
	out, err := exec.Command(cmdName, cmdArgs...).Output()
	if err != nil {
		err = xerrors.Errorf("Analyze metadata: %w", err)
		return
	}

	if err = json.Unmarshal(out, m.Info); err != nil {
		err = xerrors.Errorf("Metadata unmarshal: %w", err)
		return
	}
	return
}

// Convert checks that outFileName is abs path or put the output file in the same dir that original
func (m *MediaFile) Convert(outFileName string, args []string) (chan string, error) {
	cmdName, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, xerrors.New("ffmpeg is not installed")
	}
	// Check that out path is exist
	outPath := filepath.Dir(outFileName)
	_, err = os.Stat(outPath)
	if os.IsNotExist(err) || !filepath.IsAbs(outPath) {
		outPath = filepath.Join(filepath.Dir(m.Filename), filepath.Base(outFileName))
	} else {
		outPath = outFileName
	}

	// Set ffmpeg defaults + user args
	// default arg is answer yes to rewrite file
	// show only errors and converting status
	var cmdArgs []string

	if len(args) > 0 && args[0] != " " {
		//put -ss flag before input
		if index := indexOf("-ss", args); index+1 > 0 {
			cmdArgs = []string{"-y", "-v", "error", "-stats", args[index], args[index+1], "-i", m.Filename}
			args = append(args[:index], args[index+2:]...)
		} else {
			cmdArgs = []string{"-y", "-v", "error", "-stats", "-i", m.Filename}
		}
		cmdArgs = append(cmdArgs, args...)
	}
	cmdArgs = append(cmdArgs, outPath)

	//exec command
	cmd := exec.Command(cmdName, cmdArgs...)
	cmdReader, err := cmd.StderrPipe()
	if err != nil {
		err = xerrors.Errorf("Error creating StdoutPipe for Cmd: %w", err)
		return nil, err
	}
	// make scanner to read changing output by words
	scanner := bufio.NewScanner(cmdReader)
	scanner.Split(scanLines)
	outChan := make(chan string)

	go func() {
		defer func() {
			close(outChan)
		}()

		for scanner.Scan() {
			outChan <- stripSpaces(scanner.Text())
		}
	}()

	go func() {
		err = cmd.Start()
		if err != nil {
			err = xerrors.Errorf("Error start cmd: %w", err)
			return
		}
		outChan <- "Converting started"
		err = cmd.Wait()
		if err != nil {
			err = xerrors.Errorf("Error waiting for Cmd %w", err)
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
			return "", xerrors.Errorf("Output file: %w", err)
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

// Modified ScanerLines from Bufio to check for \r
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	if j := bytes.IndexByte(data, '\r'); j >= 0 {
		return j + 1, dropCR(data[0:j]), nil
	}

	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

// From bufio too
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func stripSpaces(str string) string {
	var prevChar rune
	return strings.Map(func(r rune) rune {
		// if the character is a space, and prevChar is space then drop it
		if unicode.IsSpace(r) && unicode.IsSpace(prevChar) {
			return -1
		}
		// icharacter is a space and precChar is "=" then drop it
		if unicode.IsSpace(r) && prevChar == rune(61) {
			return -1
		}

		prevChar = r
		// else keep it in the string
		return r
	}, str)
}

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}
