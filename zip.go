package zip

import (
	libzip "archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/src-d/go-billy.v4"
)

// find
func Find(basePath, targetDir string) ([]string, error) {

	paths := []string{}

	if targetDir != "" {
		paths = append(paths, targetDir)
	}
	err := filepath.Walk(basePath+targetDir,
		func(path string, info os.FileInfo, err error) error {
			rel, err := filepath.Rel(basePath+targetDir, path)
			if err != nil {
				return err
			}

			if info.IsDir() {
				paths = append(paths, fmt.Sprintf("%s%s/", targetDir, rel))
				return nil
			}

			paths = append(paths, fmt.Sprintf("%s%s", targetDir, rel))

			return nil
		})

	if err != nil {
		return nil, err
	}

	return paths, nil
}

func Compress(compressedFile io.Writer, targetDir string, files []string) error {
	w := libzip.NewWriter(compressedFile)

	for _, filename := range files {
		filepath := fmt.Sprintf("%s%s", targetDir, filename)
		info, err := os.Stat(filepath)
		if err != nil {
			return err
		}

		if info.IsDir() {
			continue
		}

		hdr, err := libzip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		// clear mod time
		hdr.SetModTime(time.Unix(0, 0))

		hdr.Name = filename
		hdr.Method = libzip.Deflate

		f, err := w.CreateHeader(hdr)
		if err != nil {
			return err
		}

		contents, _ := ioutil.ReadFile(filepath)
		_, err = f.Write(contents)
		if err != nil {
			return err
		}
	}

	if err := w.Close(); err != nil {
		return err
	}

	return nil
}

func FindFromFileSystem(fs billy.Filesystem, basePath, targetDir string) ([]string, error) {

	paths := []string{}

	if targetDir != "" {
		paths = append(paths, targetDir)
	}
	files, err := fs.ReadDir(basePath + targetDir)
	if err != nil {
		return nil, err
	}

	for _, info := range files {
		if info.IsDir() {
			dirPath := fmt.Sprintf("%s%s/", targetDir, info.Name())
			dFiles, err := FindFromFileSystem(fs, basePath, dirPath)
			if err != nil {
				return nil, err
			}
			if len(dFiles) > 0 {
				paths = append(paths, dFiles...)
			}
			continue
		}

		paths = append(paths, fmt.Sprintf("%s%s", targetDir, info.Name()))
	}

	return paths, nil
}

func CompressFromFileSystem(fs billy.Filesystem, compressedFile io.Writer, targetDir string, files []string) error {
	w := libzip.NewWriter(compressedFile)

	for _, filename := range files {
		filepath := fmt.Sprintf("%s%s", targetDir, filename)
		info, err := fs.Stat(filepath)
		if err != nil {
			return err
		}

		if info.IsDir() {
			continue
		}

		file, err := fs.Open(filepath)
		if err != nil {
			return err
		}
		defer file.Close()

		hdr, err := libzip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		// clear mod time
		hdr.SetModTime(time.Unix(0, 0))

		hdr.Name = filename
		hdr.Method = libzip.Deflate

		f, err := w.CreateHeader(hdr)
		if err != nil {
			return err
		}

		contents, _ := ioutil.ReadAll(file)
		_, err = f.Write(contents)
		if err != nil {
			return err
		}
	}

	if err := w.Close(); err != nil {
		return err
	}

	return nil
}
