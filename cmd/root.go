package cmd

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/leotaku/kojirou/cmd/filter"
	md "github.com/leotaku/kojirou/mangadex"
	"github.com/spf13/cobra"
)

var (
	languageArg         string
	rankArg             string
	autocropArg         bool
	kindleFolderModeArg bool
	dryRunArg           bool
	outArg              string
	forceArg            bool
	cpuprofileArg       string
	groupsFilter        string
	chaptersFilter      string
	volumesFilter       string
	helpRankingFlag     bool
	helpFilterFlag      bool
)

var rootCmd = &cobra.Command{
	Use:     "kojirou [flags..] <identifier>",
	Short:   "Generate Kindle-compatible EBooks from MangaDex",
	Version: "0.1",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		return runBusinessLogic(args[0])
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

  $ kojirou ID --language LANG --rank ALGORITHM --dry-run

Here is a short explanation for each of the available rankings.

  most (default):
Prefer chapters by groups with the most uploaded chapters.
  newest-total:
Prefer chapters by groups with the newest upload.
  newest:
Prefer chapters that have been uploaded most recently.
  views-total:
Prefer chapters by groups with the most total views.
  views:
Prefer chapters with the most views.`,
}

var helpFilterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Help topic for chapter filtering",
	Long: `Help for chapter filtering

Sometimes when downloading manga from MangaDex, you might
want to ignore chapters uploaded by scantlation groups with
less than ideal quality standards.  Other times, you may
only be interested a certain group of chapters or volumes.

To support these situations Kojirou provides a simple typed
filter system.  For now, it is possible to filter MangaDex
identifier attributes (chapters and volumes) against a list
of identifier values and other attributes (groups) against a
regular expression.  When using multiple filters, they are
combined using boolean AND, so all filters must match for a
chapter to be selected for download.

  $ kojirou ID --language LANG --chapters 1..10,Oneshot

The previous command will download chapters one through ten
as well as the special "Oneshot" chapter of the given manga.

  $ kojirou ID --language LANG --volumes 8,9,Specials

The previous command will download volumes eight, nine and
the special "Special" volume of the given manga.  You might
want to combine filtering for the last released volume of a
regularly updated manga with the "--force" flag to download
volumes that might have changed.

  $ kojirou ID --language LANG --groups !REGEX

The previous command will download all available chapters of
the given manga while ignoring uploads by groups that match
the given regular expression.  If you remove the "!" prefix
of the regular expression, Kojirou will instead only download
chapters by groups that match the regular expression.

  $ kojirou ID --language BCP_47_LANGUAGE_TAG

Technically, the "--language" option is also implemented
as a filter, however it is non-optional and must always be
given.  It accepts the format of BCP 47 language tags.`,
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

func filterFromFlags(cl md.ChapterList) (md.ChapterList, error) {
	if languageArg != "" {
		lang := filter.MatchLang(languageArg)
		cl = filter.FilterByLanguage(cl, lang)
	}
	if groupsFilter != "" {
		cl = filter.FilterByRegex(cl, "GroupNames", groupsFilter)
	}
	if volumesFilter != "" {
		ranges := filter.ParseRanges(volumesFilter)
		cl = filter.FilterByIdentifier(cl, "VolumeIdentifier", ranges)
	}
	if chaptersFilter != "" {
		ranges := filter.ParseRanges(chaptersFilter)
		cl = filter.FilterByIdentifier(cl, "Identifier", ranges)
	}

	switch rankArg {
	case "newest":
		cl = filter.SortByNewest(cl)
	case "newest-total":
		cl = filter.SortByNewestGroup(cl)
	case "views":
		cl = filter.SortByViews(cl)
	case "views-total":
		cl = filter.SortByGroupViews(cl)
	case "most":
		cl = filter.SortByMost(cl)
	default:
		return nil, fmt.Errorf(`not a valid rankinging algorithm: "%v"`, rankArg)
	}

	return filter.RemoveDuplicates(cl), nil
}

func init() {
	rootCmd.Flags().StringVarP(&languageArg, "language", "l", "en", "language for chapter downloads")
	rootCmd.Flags().StringVarP(&rankArg, "rank", "r", "most", "chapter ranking method to use")
	rootCmd.Flags().BoolVarP(&autocropArg, "autocrop", "a", false, "crop whitespace from pages automatically")
	rootCmd.Flags().BoolVarP(&kindleFolderModeArg, "kindle-folder-mode", "k", false, "generate folder structure for Kindle devices")
	rootCmd.Flags().BoolVarP(&dryRunArg, "dry-run", "d", false, "disable writing of any files")
	rootCmd.Flags().StringVarP(&outArg, "out", "o", "", "output directory")
	rootCmd.Flags().BoolVarP(&forceArg, "force", "f", false, "overwrite existing volumes")
	rootCmd.Flags().StringVarP(&cpuprofileArg, "cpuprofile", "", "", "write CPU profile to this file")
	rootCmd.Flags().StringVarP(&volumesFilter, "volumes", "V", "", "volume identifiers for chapter downloads")
	rootCmd.Flags().StringVarP(&chaptersFilter, "chapters", "C", "", "chapter identifiers for chapter downloads")
	rootCmd.Flags().StringVarP(&groupsFilter, "groups", "G", "", "scantlation groups for chapter downloads")
	rootCmd.Flags().BoolVarP(&helpRankingFlag, "help-ranking", "R", false, "Help for chapter ranking")
	rootCmd.Flags().BoolVarP(&helpFilterFlag, "help-filter", "F", false, "Help for chapter filtering")
	rootCmd.Flags().SortFlags = false
	rootCmd.Flags().MarkHidden("cpuprofile") //nolint:errcheck
	rootCmd.MarkFlagRequired("language")     //nolint:errcheck
	rootCmd.SetHelpFunc(help)
	rootCmd.SetUsageFunc(usage)
	rootCmd.ParseFlags(os.Args) //nolint:errcheck
}
