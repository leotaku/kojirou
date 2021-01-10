package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func usage(cmd *cobra.Command) error {
	groups := make(map[string][]pflag.Flag)
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" || f.Name == "version" {
			groups["Flags"] = append(groups["Flags"], *f)
		} else {
			groups["Options"] = append(groups["Options"], *f)
		}
	})

	keys := func(gs map[string][]pflag.Flag) []string {
		result := make([]string, 0)
		for name := range gs {
			result = append(result, name)
		}
		sort.Slice(result, func(i, j int) bool {
			return result[i] > result[j]
		})

		return result
	}

	fmt.Printf("Usage:\n  %v\n", cmd.Use)
	for _, name := range keys(groups) {
		fmt.Printf("\n%v:\n", name)
		for _, f := range groups[name] {
			shorthand := ""
			if len(f.Shorthand) > 0 {
				shorthand = "-" + f.Shorthand + ", "
			}
			fmt.Printf("  %4v--%-20v%v\n", shorthand, f.Name, toSentenceCase(f.Usage))
		}
	}

	return nil
}

func help(cmd *cobra.Command, args []string) {
	fmt.Printf("%v\n", cmd.Short)
	err := usage(cmd)
	if err != nil {
		panic("unreachable")
	}
}

func toSentenceCase(sentence string) string {
	words := strings.Split(sentence, " ")
	words[0] = strings.Title(words[0])
	return strings.Join(words, " ")
}
