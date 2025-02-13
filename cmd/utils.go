package cmd

import (
	"fmt"
	"os"
	"time"
)

func fatalIfError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	case diff < 48*time.Hour:
		return "yesterday"
	case diff < 168*time.Hour:
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	default:
		return t.Format("Jan 02, 2006")
	}
}

type FilebConfig struct {
	Filesystems []FilesystemConfig `mapstructure:"filesystems"`
	Storage     StorageConfig      `mapstructure:"storage"`
}

type StorageConfig struct {
	Type string `mapstructure:"type"`
	DB   string `mapstructure:"db"`
}

// BaseFilesystemConfig contains common fields for all filesystem types
type BaseFilesystemConfig struct {
	Name    string `mapstructure:"name"`
	Type    string `mapstructure:"type"`
	Default bool   `mapstructure:"default"`
}

// LocalFilesystemConfig for local filesystem
type LocalFilesystemConfig struct {
	BaseFilesystemConfig `mapstructure:",squash"`
	Path                 string `mapstructure:"path"`
}

// S3FilesystemConfig for S3 filesystem
type S3FilesystemConfig struct {
	BaseFilesystemConfig `mapstructure:",squash"`
	Endpoint             string `mapstructure:"endpoint"`
	Region               string `mapstructure:"region"`
	Bucket               string `mapstructure:"bucket"`
	AccessKey            string `mapstructure:"access-key"`
	SecretKey            string `mapstructure:"secret-key"`
}

// GSFilesystemConfig for Google Cloud Storage
type GSFilesystemConfig struct {
	BaseFilesystemConfig `mapstructure:",squash"`
	Bucket               string `mapstructure:"bucket"`
	CredentialsFile      string `mapstructure:"credentials-file"`
}

// FilesystemConfig is now a wrapper that can hold any type of filesystem config
type FilesystemConfig struct {
	BaseFilesystemConfig `mapstructure:",squash"`
	Config               map[string]interface{} `mapstructure:",remain"`
}
