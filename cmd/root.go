package cmd

import (
	"fmt"
	"os"
	"path"
	"runtime/pprof"
	"strconv"

	"github.com/leotaku/manki/cmd/util"
	"github.com/leotaku/manki/mangadex"
	"github.com/spf13/cobra"
)

var (
	languageArg         string
	rankArg             string
	kindleFolderModeArg bool
	dryRunArg           bool
	outArg              string
	cpuprofileArg       string
)

var rootCmd = &cobra.Command{
	Use:     "manki [flags..] <identifier>",
	Short:   "Generate Kindle-compatible EBooks from MangaDex",
	Version: "0.1",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 32)
		if err != nil {
			return fmt.Errorf(`parsing "%v": not a valid identifier`, args[0])
		}
		cmd.SilenceUsage = true

		if cpuprofileArg != "" {
			f, err := os.Create(cpuprofileArg)
			if err != nil {
				return err
			}
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}
		util.InitCleanup()
		defer util.RunCleanup()

		manga, err := downloadMetaFor(int(id), filterFromFlags)
		if err != nil {
			return err
		}

		// Write
		if dryRunArg {
			return nil
		} else if !kindleFolderModeArg {
			return runInNormalMode(*manga)
		} else {
			return runInKindleMode(*manga)
		}
	},
	DisableFlagsInUseLine: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func filterFromFlags(cl mangadex.ChapterList) mangadex.ChapterList {
	lang := util.MatchLang(languageArg)
	cl = filterLang(cl, lang)

	switch rankArg {
	case "newest":
		cl = rankNewest(cl)
	case "newest-total":
		cl = rankTotalNewest(cl)
	case "views":
		cl = rankViews(cl)
	case "views-total":
		cl = rankTotalViews(cl)
	case "most":
		cl = rankMost(cl)
	default:
		cl = make(mangadex.ChapterList, 0)
	}

	return doRank(cl)
}

func runInNormalMode(m mangadex.Manga) error {
	if outArg == "" {
		outArg = m.Info.Title
	}

	// Setup directories
	err := util.SetupDirectories(outArg)
	if err != nil {
		return err
	}

	return downloadAndWrite(m, outArg, nil)
}

func runInKindleMode(m mangadex.Manga) error {
	if outArg == "" {
		outArg = "kindle"
	}
	root := path.Join(outArg, "documents", m.Info.Title)
	thumbRoot := path.Join(outArg, "system", "thumbnails")

	// Setup directories
	err := util.SetupDirectories(root, thumbRoot)
	if err != nil {
		return err
	}

	return downloadAndWrite(m, root, &thumbRoot)
}

func init() {
	rootCmd.Flags().StringVarP(&languageArg, "language", "l", "en", "language for chapter downloads")
	rootCmd.Flags().StringVarP(&rankArg, "rank", "r", "newest-total", "chapter ranking method to use")
	rootCmd.Flags().BoolVarP(&kindleFolderModeArg, "kindle-folder-mode", "k", false, "generate folder structure for Kindle devices")
	rootCmd.Flags().BoolVarP(&dryRunArg, "dry-run", "d", false, "disable writing of any files")
	rootCmd.Flags().StringVarP(&outArg, "out", "o", "", "output directory")
	rootCmd.Flags().StringVarP(&cpuprofileArg, "cpuprofile", "", "", "write CPU profile to this file")
	rootCmd.Flags().SortFlags = false
	rootCmd.MarkFlagRequired("language")
	rootCmd.SetHelpFunc(help)
	rootCmd.SetUsageFunc(usage)
}
