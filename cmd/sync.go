package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	syncCmd.Flags().BoolP("recursive", "r", false, "Remove directories and their contents recursively")
	RootCmd.AddCommand(rmCmd)
}

var syncCmd = &cobra.Command{
	Use:   "sync [uri] [uri]",
	Short: "Sync files between two directories",
	Long:  `Sync files between two directories`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Context()
		_, _ = cmd.Flags().GetBool("recursive")
		_, _ = cmd.Flags().GetBool("verbose")

		// 	// Create a new span for the sync command
		// 	ctx, span := otel.Tracer("").Start(ctx, "sync-command")
		// 	defer span.End()

		// 	// Log with severity levels
		// 	if verbose {
		// 		logger.Debug(ctx, "Starting sync operation",
		// 			logs.WithAttributes(
		// 				attribute.StringSlice("args", args),
		// 				attribute.Bool("recursive", recursive),
		// 			),
		// 		)
		// 	}

		// 	uri, err := uri.ParseURI(args[0])
		// 	if err != nil {
		// 		logger.Error(ctx, "Failed to parse URI",
		// 			logs.WithAttributes(attribute.String("uri", args[0])),
		// 			logs.WithError(err),
		// 		)
		// 		fatalIfError(err)
		// 	}

		// 	err = cmdManager.Delete(ctx, uri, recursive)
		// 	if err != nil {
		// 		logger.Error(ctx, "Failed to delete",
		// 			logs.WithAttributes(
		// 				attribute.String("uri", uri.String()),
		// 				attribute.Bool("recursive", recursive),
		// 			),
		// 			logs.WithError(err),
		// 		)
		// 		fatalIfError(err)
		// 	}

		// 	logger.Info(ctx, "Successfully completed sync operation",
		// 		logs.WithAttributes(
		// 			attribute.String("uri", uri.String()),
		// 			attribute.Bool("recursive", recursive),
		// 		),
		// 	)
	},
}
