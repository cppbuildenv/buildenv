package io

import "testing"

func TestExtractTarGz(t *testing.T) {
	ExtractTarGz("testdata/temp/ti-processor-sdk-rtos-j721e-evm-07_03_00_07.tar.gz", "testdata/temp")
}
