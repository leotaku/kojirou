package cmd

import (
	"fmt"

	"github.com/leotaku/kojirou/cmd/formats/download"
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
