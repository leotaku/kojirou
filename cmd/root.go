package cmd

import (
	"os"
	"path"
	"strconv"

	"github.com/leotaku/manki/cmd/util"
	"github.com/spf13/cobra"
)

var kindleFolderModeArg bool
var langArg string
var outArg string

var rootCmd = &cobra.Command{
	Use:     "manki [flags..] <identifier>",
	Short:   "Generate Kindle-compatible EBooks from MangaDex",
	Version: "0.1",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		util.InitCleanup()
		defer util.RunCleanup()

		id, err := strconv.ParseInt(args[0], 10, 32)
		if err != nil {
			return err
		}

		down, err := preDownload(int(id))
		if err != nil {
			return err
		}

		// Variables
		title := down.incomplete.Info.Title

		// Write
		if !kindleFolderModeArg {
			if len(outArg) == 0 {
				outArg = title
			}

			// Setup directories
			err := setupDirectories(outArg)
			if err != nil {
				return err
			}

			return download(*down, outArg, nil)
		} else {
			if len(outArg) == 0 {
				outArg = "kindle"
			}
			root := path.Join(outArg, "documents", title)
			thumbRoot := path.Join(outArg, "system", "thumbnails")

			// Setup directories
			err := setupDirectories(root, thumbRoot)
			if err != nil {
				return err
			}

			return download(*down, root, &thumbRoot)
		}
	},
	DisableFlagsInUseLine: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// rootCmd.Flags().IntVarP(&limitArg, "max", "x", 8, "maximum number of concurrent connections")
	// rootCmd.Flags().IntVarP(&retryArg, "retry", "r", 4, "number of retries on failed downloads")
	rootCmd.Flags().BoolVarP(&kindleFolderModeArg, "kindle-folder-mode", "k", false, "generate folder structure for Kindle devices")
	rootCmd.Flags().StringVarP(&langArg, "language", "l", "en", "language for MangaDex download")
	rootCmd.Flags().StringVarP(&outArg, "out", "o", "", "output directory")
	rootCmd.Flags().SortFlags = false
}
