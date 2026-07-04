package safepath

import (
	"io"
	"os"
	"path/filepath"
)

// Open opens the named file inside baseDir for reading.
func Open(baseDir, name string) (io.ReadCloser, error) {
	p := filepath.Join(baseDir, name)
	return os.Open(p)
}
