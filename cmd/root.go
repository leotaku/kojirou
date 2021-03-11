package cmd

import (
	"fmt"
	"os"
	"path"
	"runtime/pprof"
	"strconv"

	"github.com/leotaku/kojirou/cmd/util"
	"github.com/leotaku/kojirou/mangadex"
	"github.com/spf13/cobra"
)

var (
	languageArg         string
	rankArg             string
	kindleFolderModeArg bool
	dryRunArg           bool
	outArg              string
	cpuprofileArg       string
	helpRankingFlag     bool
	helpFilterFlag      bool
)

var rootCmd = &cobra.Command{
	Use:     "kojirou [flags..] <identifier>",
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
			if err := pprof.StartCPUProfile(f); err != nil {
				return err
			}
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

var helpRankingCmd = &cobra.Command{
	Use:   "ranking",
	Short: "Help topic for chapter ranking",
	Long: `Help for chapter ranking

As you might already know manga on MangaDex are scanned,
translated and typeset by independent hobbyist groups
generally referred to as "scantlators".  Because of the
lack of any monetary incentive, it is rare for a project
to be scantlated from beginning to end by a single group.

To make the best out of this situation, Kojirou provides a
rudimentary ranking system in order to select the highest
quality scantlations.

By running the following command, you can try out different
ranking algorithms without downloading chapters or images.
If you are happy with the resulting list of chapters, just
remove the "--dry-run" switch to download the manga.

  $ kojirou --language LANG --rank ALGORITHM --dry-run

Here is a short explanation for each of the available rankings.

  newest-total (default):
Prefer chapters by groups with the newest upload.
  newest:
Prefer chapters that have been uploaded most recently.
  views-total:
Prefer chapters by groups with the most total views.
  views:
Prefer chapters with the most views.
  most:
Prefer chapters by groups with the most uploaded chapters.`,
}

var helpFilterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Help topic for chapter filtering",
	Long: `Help for chapter filtering

The filtering system is not yet implemented.`,
}

func Execute() {
	if helpRankingFlag {
		helpRankingCmd.Help() //nolint:errcheck
	} else if helpFilterFlag {
		helpFilterCmd.Help() //nolint:errcheck
	} else if err := rootCmd.Execute(); err != nil {
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
	rootCmd.Flags().BoolVarP(&helpRankingFlag, "help-ranking", "R", false, "Help for chapter ranking")
	rootCmd.Flags().BoolVarP(&helpFilterFlag, "help-filter", "F", false, "Help for chapter filtering")
	rootCmd.Flags().SortFlags = false
	rootCmd.MarkFlagRequired("language") //nolint:errcheck
	rootCmd.SetHelpFunc(help)
	rootCmd.SetUsageFunc(usage)
	rootCmd.ParseFlags(os.Args) //nolint:errcheck
}
