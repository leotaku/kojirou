package cmd

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func writeHelp(cmd *cobra.Command, w io.Writer) {
	groups := make(map[string][]pflag.Flag)
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		switch {
		case f.Hidden:
		case strings.HasPrefix(f.Name, "help") || f.Name == "version":
			groups["3Flags"] = append(groups["3Flags"], *f)
		case strings.HasSuffix(f.Name, "s"):
			groups["2Filters"] = append(groups["2Filters"], *f)
		default:
			groups["1Options"] = append(groups["1Options"], *f)
		}
	})

	keys := func(gs map[string][]pflag.Flag) []string {
		result := make([]string, 0)
		for name := range gs {
			result = append(result, name)
		}
		sort.Slice(result, func(i, j int) bool {
			return result[i] < result[j]
		})

		return result
	}

	fmt.Fprintf(w, "Usage:\n  %v\n", cmd.Use)
	for _, name := range keys(groups) {
		fmt.Fprintf(w, "\n%v:\n", name[1:])
		for _, f := range groups[name] {
			shorthand := ""
			if len(f.Shorthand) > 0 {
				shorthand = "-" + f.Shorthand + ", "
			}
			fmt.Fprintf(w, "  %4v--%-20v%v\n", shorthand, f.Name, toSentenceCase(f.Usage))
		}
	}
}

func help(cmd *cobra.Command, args []string) {
	fmt.Fprintf(os.Stdout, "%v\n", cmd.Short)
	writeHelp(cmd, os.Stdout)
}

func usage(cmd *cobra.Command) error {
	writeHelp(cmd, os.Stderr)
	return nil
}

func toSentenceCase(sentence string) string {
	words := strings.Split(sentence, " ")
	words[0] = cases.Title(language.English).String(words[0])
	return strings.Join(words, " ")
}
