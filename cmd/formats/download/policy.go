package download

import "fmt"

type DataSaverPolicy int

const (
	DataSaverPolicyNo DataSaverPolicy = iota
	DataSaverPolicyPrefer
	DataSaverPolicyFallback
)

func (p *DataSaverPolicy) String() string {
	switch *p {
	case DataSaverPolicyNo:
		return "no"
	case DataSaverPolicyPrefer:
		return "prefer"
	case DataSaverPolicyFallback:
		return "fallback"
	default:
		panic("unreachable")
	}
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (p *DataSaverPolicy) Set(v string) error {
	switch v {
	case "no":
		*p = DataSaverPolicyNo
	case "prefer":
		*p = DataSaverPolicyPrefer
	case "fallback":
		*p = DataSaverPolicyFallback
	default:
		return fmt.Errorf(`must be one of: "no", "prefer", or "fallback"`)
	}

	return nil
}

// Type is only used in help text
func (p *DataSaverPolicy) Type() string {
	return "data-saver policy"
}
