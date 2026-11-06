package controller

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
)

// PackFolder walks a directory and writes all files into a tar stream
func PackFolder(rootPath string, w io.Writer) error {
	tw := tar.NewWriter(w)
	defer tw.Close()

	info, err := os.Stat(rootPath)
	if err != nil {
		return err
	}

	// If a single file was passed instead of a directory, pack it directly
	if !info.IsDir() {
		file, err := os.Open(rootPath)
		if err != nil {
			return err
		}
		defer file.Close()

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		header.Name = filepath.Base(rootPath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		_, err = io.Copy(tw, file)
		return err
	}

	return filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == rootPath {
			return nil
		}

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.Mode().IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tw, file)
		return err
	})
}

// UnpackArchive extracts a tar stream into the destination directory
func UnpackArchive(r io.Reader, destDir string) error {
	tr := tar.NewReader(r)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			parentDir := filepath.Dir(targetPath)
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				return err
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, header.FileInfo().Mode())
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}
