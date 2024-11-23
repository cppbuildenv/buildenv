package generator

import (
	"os"
	"testing"
)

func TestGenModules(t *testing.T) {
	generator := modules{
		config: GeneratorConfig{
			Libname:   "ffmpeg",
			Version:   "0.8.0",
			Libtype:   "SHARED",
			BuildType: "Release",
			Namespace: "ffmpeg",
			Components: []Component{
				{
					Component: "avutil",
					Soname:    "libavutil.so.56",
					Filename:  "libavutil.so.56.70.100",
				},
				{
					Component: "avcodec",
					Soname:    "libavcodec.so.58",
					Filename:  "libavcodec.so.58.134.100",
					Dependencies: []string{
						"avutil",
					},
				},
				{
					Component: "avdevice",
					Soname:    "libavdevice.so.58",
					Filename:  "libavdevice.so.58.13.100",
					Dependencies: []string{
						"avformat", "avutil",
					},
				},
				{
					Component: "avfilter",
					Soname:    "libavfilter.so.7",
					Filename:  "libavfilter.so.7.110.100",
					Dependencies: []string{
						"avcodec", "avformat", "avutil",
					},
				},
				{
					Component: "avformat",
					Soname:    "libavformat.so.58",
					Filename:  "libavformat.so.58.76.100",
					Dependencies: []string{
						"avcodec", "avutil",
					},
				},
			},
		},
	}

	if err := generator.generate("temp/ffmpeg"); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll("temp"); err != nil {
		t.Fatal(err)
	}
}
