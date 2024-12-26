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

type WidepagePolicyArg kindle.WidepagePolicy

func (p *WidepagePolicyArg) String() string {
	switch kindle.WidepagePolicy(*p) {
	case kindle.WidepagePolicyPreserve:
		return "preserve"
	case kindle.WidepagePolicySplit:
		return "split"
	case kindle.WidepagePolicyPreserveAndSplit:
		return "preserve-and-split"
	case kindle.WidepagePolicySplitAndPreserve:
		return "split-and-preserve"
	default:
		panic("unreachable")
	}
}

func (p *WidepagePolicyArg) Set(v string) error {
	switch v {
	case "preserve":
		*p = WidepagePolicyArg(kindle.WidepagePolicyPreserve)
	case "split":
		*p = WidepagePolicyArg(kindle.WidepagePolicySplit)
	case "preserve-and-split":
		*p = WidepagePolicyArg(kindle.WidepagePolicyPreserveAndSplit)
	case "split-and-preserve":
		*p = WidepagePolicyArg(kindle.WidepagePolicySplitAndPreserve)
	default:
		return fmt.Errorf(`must be one of: "preserve", "split", or "both"`)
	}

	return nil
}

func (p *WidepagePolicyArg) Type() string {
	return "wide-page policy"
}
