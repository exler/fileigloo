package server

import "path"

func SanitizeFilename(filename string) string {
	return path.Clean(path.Base(filename))
}
