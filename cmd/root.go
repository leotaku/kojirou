package cmd

import (
	"os"
	"path"
	"strconv"

	"github.com/leotaku/manki/cmd/util"
	"github.com/spf13/cobra"
)

var (
	kindleFolderModeArg bool
	langArg             string
	outArg              string
)

var rootCmd = &cobra.Command{
	Use:     "manki [flags..] <identifier>",
	Short:   "Generate Kindle-compatible EBooks from MangaDex",
	Version: "0.1",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 32)
		if err != nil {
			return err
		}
		cmd.SilenceUsage = true

		util.InitCleanup()
		defer util.RunCleanup()

		down, err := downloadMetaFor(int(id), nil)
		if err != nil {
			return err
		}

		// Variables
		title := down.Info.Title

		// Write
		if !kindleFolderModeArg {
			if len(outArg) == 0 {
				outArg = title
			}

			// Setup directories
			err := util.SetupDirectories(outArg)
			if err != nil {
				return err
			}

			return downloadWriteVolumes(*down, outArg, nil)
		} else {
			if len(outArg) == 0 {
				outArg = "kindle"
			}
			root := path.Join(outArg, "documents", title)
			thumbRoot := path.Join(outArg, "system", "thumbnails")

			// Setup directories
			err := util.SetupDirectories(root, thumbRoot)
			if err != nil {
				return err
			}

			return downloadWriteVolumes(*down, root, &thumbRoot)
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
	rootCmd.Flags().StringVarP(&langArg, "language", "l", "en", "language for MangaDex download")
	rootCmd.Flags().BoolVarP(&kindleFolderModeArg, "kindle-folder-mode", "k", false, "generate folder structure for Kindle devices")
	rootCmd.Flags().StringVarP(&outArg, "out", "o", "", "output directory")
	rootCmd.Flags().SortFlags = false
	rootCmd.SetHelpFunc(help)
	rootCmd.SetUsageFunc(usage)
}
