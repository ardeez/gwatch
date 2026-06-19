package config

import (
	"errors"
	"flag"
	"path/filepath"
	"strings"
)

type Config struct {
	Entry    string
	Dir      string
	Ext      string
	Exclude  []string
	Interval int
}

func ParseConfig() (*Config, error) {
	entry := flag.String("entry", "", "Entry point to build, e.g. ./cmd/server (Required)")
	dir := flag.String("dir", ".", "Root directory to watch")
	ext := flag.String("ext", ".go", "File extension to watch")
	excludeRaw := flag.String("exclude", "vendor,tmp", "Comma-separated directories to exclude")
	interval := flag.Int("interval", 500, "Poll interval in milliseconds")

	flag.Parse()

	if *entry == "" {
		return nil, errors.New("flag -entry is required (e.g., -entry ./cmd/server)")
	}

	var excludes []string
	if *excludeRaw != "" {
		parts := strings.Split(*excludeRaw, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				excludes = append(excludes, filepath.Clean(trimmed))
			}
		}
	}
	extension := *ext
	if extension != "" && !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	return &Config{
		Entry:    filepath.Clean(*entry),
		Dir:      filepath.Clean(*dir),
		Ext:      extension,
		Exclude:  excludes,
		Interval: *interval,
	}, nil
}

