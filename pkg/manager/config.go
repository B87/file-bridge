package manager

import (
	"context"

	"github.com/B87/file-bridge/pkg/filesystem"
	"github.com/B87/file-bridge/pkg/filesystem/gs"
	"github.com/B87/file-bridge/pkg/filesystem/local"
)

// FSConfig is a configuration for a file system.
type FSConfig struct {
	Name    string
	Type    string
	Default bool
	Config  map[string]interface{}
}

type Config struct {
	FsConfigs []FSConfig
	FsMap     map[string]filesystem.FS
}

func DefaultConfig() *Config {
	fsMap := make(map[string]filesystem.FS)

	fsMap["local"] = local.NewLocalFS()
	fsMap["local2"] = local.NewLocalFS()
	fsMap["gs"] = gs.NewGCPBucketFS(context.Background())

	return &Config{
		FsConfigs: []FSConfig{
			{
				Name: "local",
				Type: "local",
				Config: map[string]interface{}{
					"path": "./local/drive",
				},
				Default: true,
			},
			{
				Name: "local2",
				Type: "local",
				Config: map[string]interface{}{
					"path": "./local/drive2",
				},
				Default: false,
			},
			{
				Name: "gdrive",
				Type: "gdrive",
				Config: map[string]interface{}{
					"drive_id": "1XdpeHDEyyGERSpUyBcro02bOQPUziq9P",
				},
			},
			{
				Name: "gs",
				Type: "gs",
				Config: map[string]interface{}{
					"bucket_name": "gindumac-public",
				},
			},
		},
		FsMap: fsMap,
	}
}
