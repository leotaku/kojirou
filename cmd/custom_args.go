package cmd

import (
	"fmt"

	"github.com/leotaku/kojirou/cmd/formats/download"
	"github.com/leotaku/kojirou/cmd/formats/kindle"
)

type DataSaverPolicyArg download.DataSaverPolicy

func (p *DataSaverPolicyArg) String() string {
	switch download.DataSaverPolicy(*p) {
	case download.DataSaverPolicyNo:
		return "no"
	case download.DataSaverPolicyPrefer:
		return "prefer"
	case download.DataSaverPolicyFallback:
		return "fallback"
	default:
		panic("unreachable")
	}
}

func (p *DataSaverPolicyArg) Set(v string) error {
	switch v {
	case "no":
		*p = DataSaverPolicyArg(download.DataSaverPolicyNo)
	case "prefer":
		*p = DataSaverPolicyArg(download.DataSaverPolicyPrefer)
	case "fallback":
		*p = DataSaverPolicyArg(download.DataSaverPolicyFallback)
	default:
		return fmt.Errorf(`must be one of: "no", "prefer", or "fallback"`)
	}

	return nil
}

func (p *DataSaverPolicyArg) Type() string {
	return "data-saver policy"
}

type AutosplitPolicyArg kindle.AutosplitPolicy

func (p *AutosplitPolicyArg) String() string {
	switch kindle.AutosplitPolicy(*p) {
	case kindle.AutosplitPolicyPreserve:
		return "preserve"
	case kindle.AutosplitPolicySplit:
		return "split"
	case kindle.AutosplitPolicyBoth:
		return "both"
	default:
		panic("unreachable")
	}
}

func (p *AutosplitPolicyArg) Set(v string) error {
	switch v {
	case "preserve":
		*p = AutosplitPolicyArg(kindle.AutosplitPolicyPreserve)
	case "split":
		*p = AutosplitPolicyArg(kindle.AutosplitPolicySplit)
	case "both":
		*p = AutosplitPolicyArg(kindle.AutosplitPolicyBoth)
	default:
		return fmt.Errorf(`must be one of: "preserve", "split", or "both"`)
	}

	return nil
}

func (p *AutosplitPolicyArg) Type() string {
	return "auto-split policy"
}
