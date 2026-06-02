// Copyright 2026 The Planwright Authors
// SPDX-License-Identifier: Apache-2.0

package artifact

import (
	"bytes"
	"path/filepath"
	"sort"
)

type File struct {
	Path string
	Data []byte
}

func ByPath(files []File, path string) (File, bool) {
	for _, file := range files {
		if filepath.ToSlash(file.Path) == filepath.ToSlash(path) {
			return file, true
		}
	}
	return File{}, false
}

func Paths(files []File) []string {
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, filepath.ToSlash(file.Path))
	}
	sort.Strings(paths)
	return paths
}

func JoinContents(files []File) []byte {
	sorted := append([]File(nil), files...)
	Sort(sorted)

	var joined bytes.Buffer
	for _, file := range sorted {
		joined.Write(file.Data)
		joined.WriteByte('\n')
	}
	return joined.Bytes()
}

func Sort(files []File) {
	sort.Slice(files, func(i, j int) bool {
		return filepath.ToSlash(files[i].Path) < filepath.ToSlash(files[j].Path)
	})
}

func Prefix(prefix string, files []File) []File {
	prefixed := make([]File, 0, len(files))
	for _, file := range files {
		prefixed = append(prefixed, File{
			Path: filepath.ToSlash(filepath.Join(prefix, file.Path)),
			Data: append([]byte(nil), file.Data...),
		})
	}
	return prefixed
}
