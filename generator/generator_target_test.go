package generator

import (
	"os"
	"testing"
)

func TestConfigVersion(t *testing.T) {
	config := configVersion{
		cmakeConfig: CMakeConfig{
			Libname: "yaml-cpp",
			Version: "0.8.0",
		},
	}

	if err := config.generate("temp/yaml-cpp"); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll("temp"); err != nil {
		t.Fatal(err)
	}
}

func TestConfig(t *testing.T) {
	config := config{
		cmakeConfig: CMakeConfig{
			Namespace: "yaml-cpp",
			Libname:   "yaml-cpp",
		},
	}

	if err := config.generate("temp/yaml-cpp"); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll("temp"); err != nil {
		t.Fatal(err)
	}
}

func TestTargets(t *testing.T) {
	config := targets{
		cmakeConfig: CMakeConfig{
			Libname:   "yaml-cpp",
			Namespace: "yaml-cpp",
			Libtype:   "SHARED",
		},
	}

	if err := config.generate("temp/yaml-cpp"); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll("temp"); err != nil {
		t.Fatal(err)
	}
}

func TestTargetsType(t *testing.T) {
	target := targetsBuildType{
		cmakeConfig: CMakeConfig{
			SystemName: "Linux",
			Namespace:  "yaml-cpp",
			Libname:    "yaml-cpp",
			Libtype:    "SHARED",
			BuildType:  "Release",
			Filename:   "libyaml-cpp.so.0.8.0",
			Soname:     "libyaml-cpp.0.8",
		},
	}

	if err := target.generate("temp/yaml-cpp"); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll("temp"); err != nil {
		t.Fatal(err)
	}
}

func TestModulesBuildType(t *testing.T) {
	generator := modulesBuildType{
		cmakeConfig: CMakeConfig{
			Libname:    "ffmpeg",
			Version:    "0.8.0",
			Libtype:    "SHARED",
			BuildType:  "Release",
			Namespace:  "ffmpeg",
			SystemName: "Linux",
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

func TestModules(t *testing.T) {
	generator := modules{
		cmakeConfig: CMakeConfig{
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
