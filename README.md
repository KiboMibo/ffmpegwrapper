## Simple ffmpeg wrapper

Big thanks to:

[barsanuphe]("https://github.com/barsanuphe") for [goexiftool]("https://github.com/barsanuphe/goexiftool/blob/master/goexiftool.go") wrapper.

[xfrr]("https://github.com/xfrr") for [goffmpeg]("https://github.com/xfrr/goffmpeg") wrapper.

### Defaults

Wrapper uses ffprobe to get mediafile info, so there is default args:

```
   ffprobe -show_format -show_streams -pretty -print_format json -hide_banner -i %file_name%
```

This command produce output in json format, so we can unmarshall it to Info struct.

Defaults for ffmpeg command:

```
   ffmpeg -y -v error -stats -i %file_name% %out_file_name%
```

``` -y ```       for answer yes on file rewrite

``` -v error ``` verbose with only errors

```-stats``` show transcoding status

```-i``` input file


### Example
```go
func main(){
     // Initialize
     m, err := NewMediaFile("./test/test.MOV")
     if err != nil {
        panic(fmt.Errorf("Failed get file info : %s", err))
     }
     // Data from ffprobe
     fmt.Println(m.Info)

     // ffmpeg params
     ffmpParams := "-crf 20 -bufsize 4096k -vf scale=1280:800:force_original_aspect_ratio=decrease"
     out, err := m.Convert("pre_test.mov", ffmpParams);
     if err != nil{
     		panic(fmt.Errorf("Failed convert file : %s", err))
     }
     // Getting status of transcoding
     for msg := range out {
       fmt.Println(msg)
     }
}
```
