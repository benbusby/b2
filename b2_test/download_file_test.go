package b2_test

import (
	"testing"
)

func TestDownloadFile(t *testing.T) {
	file := uploadTestFile("download.txt")

	contents, err := account.DownloadById(file.FileID)
	if err != nil {
		t.Fatal("Failed to download file contents")
	} else if len(contents) == 0 {
		t.Fatal("Downloaded file is empty")
	} else if string(contents) != testString {
		t.Fatal("Downloaded content does not match expected")
	}
}

func TestPartialDownload(t *testing.T) {
	file := uploadTestFile("partial-download.txt")

	contents, err := account.PartialDownloadById(file.FileID, 0, 4)
	if err != nil {
		t.Fatalf("Failed partial download: %v", err)
	} else if len(contents) != 5 {
		t.Fatalf("Incorrect partial download size: "+
			"expected=%d, received=%d",
			5,
			len(contents))
	} else if string(contents) != testString[0:5] {
		t.Fatalf("Invalid download contents: "+
			"expected=%s, received=%s",
			testString[0:5],
			string(contents))
	}
}

func TestLocalDownload(t *testing.T) {
	file := uploadLocalTestFile("local-download.txt")

	contents, err := dummyAccount.DownloadById(file.FileID)
	if err != nil {
		t.Fatalf("Failed to \"download\" local file contents: %v", err)
	} else if len(contents) == 0 {
		t.Fatal("Local file is empty")
	} else if string(contents) != testString {
		t.Fatal("Local file content does not match expected")
	}

	contents, err = dummyAccount.PartialDownloadById(file.FileID, 1, 5)
	if err != nil {
		t.Fatalf("Failed partial local download: %v", err)
	} else if len(contents) != 5 {
		t.Fatalf("Incorrect partial local download size: "+
			"expected=%d, received=%d",
			5,
			len(contents))
	} else if string(contents) != testString[1:6] {
		t.Fatalf("Invalid local download contents: "+
			"expected=%s, received=%s",
			testString[1:6],
			string(contents))
	}
}
