package cmd

import (
	"context"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/B87/file-bridge/pkg/filesystem"
	"github.com/B87/file-bridge/pkg/manager"
	"github.com/B87/file-bridge/pkg/telemetry"
)

func init() {
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose, default false")
}

var cmdManager = initManager()

func initManager() *manager.Manager {
	// config, err := readFilebConfig()
	// fatalIfError(err)
	config := manager.DefaultConfig()
	m, err := manager.NewManager(context.Background(), config)
	fatalIfError(err)
	return m
}

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "fileb",
	Short: "File Bridge CLI",
	Long: `
File Bridge CLI interacts with files across file systems.
`,
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func readFilebConfig() (*manager.Config, error) {
	viper.SetConfigName(".fileb")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, using defaults
			return manager.DefaultConfig(), nil
		}
		return nil, err
	}

	var config FilebConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	fsConfigs := []filesystem.Config{}
	for _, fsConfig := range config.Filesystems {
		switch fsConfig.Type {
		case "local":
			var localConfig LocalFilesystemConfig
			if err := remapConfig(fsConfig, &localConfig); err != nil {
				fatalIfError(fmt.Errorf("invalid local filesystem config: %w", err))
			}
			fsConfigs = append(fsConfigs, filesystem.Config{
				Name:    fsConfig.Name,
				Type:    fsConfig.Type,
				Default: fsConfig.Default,
				Config: map[string]interface{}{
					"path": localConfig.Path,
				},
			})
		case "s3":
			var s3Config S3FilesystemConfig
			if err := remapConfig(fsConfig, &s3Config); err != nil {
				fatalIfError(fmt.Errorf("invalid S3 filesystem config: %w", err))
			}
			fsConfigs = append(fsConfigs, filesystem.Config{
				Name:    fsConfig.Name,
				Type:    fsConfig.Type,
				Default: fsConfig.Default,
				Config: map[string]interface{}{
					"endpoint":   s3Config.Endpoint,
					"region":     s3Config.Region,
					"bucket":     s3Config.Bucket,
					"access-key": s3Config.AccessKey,
					"secret-key": s3Config.SecretKey,
				},
			})
		case "gs":
			var gsConfig GSFilesystemConfig
			if err := remapConfig(fsConfig, &gsConfig); err != nil {
				fatalIfError(fmt.Errorf("invalid GS filesystem config: %w", err))
			}
			fsConfigs = append(fsConfigs, filesystem.Config{
				Name:    fsConfig.Name,
				Type:    fsConfig.Type,
				Default: fsConfig.Default,
				Config: map[string]interface{}{
					"bucket":           gsConfig.Bucket,
					"credentials-file": gsConfig.CredentialsFile,
				},
			})
		default:
			fatalIfError(fmt.Errorf("unknown filesystem type: %s", fsConfig.Type))
		}
	}
	return &manager.Config{FsConfigs: fsConfigs}, nil
}

// Helper function to remap config from one struct to another
func remapConfig(in interface{}, out interface{}) error {
	// Convert the input struct to a map
	inMap := make(map[string]interface{})

	// If the input is a FilesystemConfig, merge both the base fields and Config map
	if fsConfig, ok := in.(FilesystemConfig); ok {
		// Add base fields
		inMap["name"] = fsConfig.Name
		inMap["type"] = fsConfig.Type
		inMap["default"] = fsConfig.Default

		// Merge the remaining config fields
		for k, v := range fsConfig.Config {
			inMap[k] = v
		}
	} else {
		return fmt.Errorf("expected FilesystemConfig, got %T", in)
	}

	// Create a new Viper instance
	config := viper.New()

	// Set all the values in the map
	for k, v := range inMap {
		config.Set(k, v)
	}

	// Unmarshal directly to the output struct
	if err := config.Unmarshal(out); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// Add this function to initialize the console tracer
func InitConsoleTracer() (*trace.TracerProvider, error) {
	tp := trace.NewTracerProvider(
		trace.WithBatcher(&telemetry.SimpleSpanExporter{}),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}
