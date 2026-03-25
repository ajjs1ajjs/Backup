//go:build !windows
// +build !windows

package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/sys/unix"
)

func (p *LocalProvider) GetSpace() (total, free, used int64, err error) {
	stat := &unix.Statfs_t{}
	err = unix.Statfs(p.Path, stat)
	if err != nil {
		return 0, 0, 0, err
	}
	total = int64(stat.Blocks) * int64(stat.Bsize)
	free = int64(stat.Bfree) * int64(stat.Bsize)
	used = total - free
	return total, free, used, nil
}
