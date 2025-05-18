package tester

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func unzipArchive(r io.ReaderAt, size int64, destPath string) error {
	reader, err := zip.NewReader(r, size)
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		if !strings.HasPrefix(file.Name, "tests/") {
			continue
		}

		destFilePath := filepath.Join(destPath, file.Name)

		if err := os.MkdirAll(filepath.Dir(destFilePath), 0755); err != nil {
			return err
		}

		if file.FileInfo().IsDir() {
			continue
		}

		outFile, err := os.Create(destFilePath)
		if err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
