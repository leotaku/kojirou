package mangadex

import (
	"fmt"
	"math"
	"strconv"
)

type Identifier struct {
	Numeric float64
}

func GuessIdentifier(num string) Identifier {
	f, err := strconv.ParseFloat(num, 64)
	if err != nil {
		f = math.NaN()
	}

	return Identifier{
		Numeric: f,
	}
}

func (n Identifier) Less(o Identifier) bool {
	switch {
	case n.IsUnknown():
		return false
	case o.IsUnknown():
		return true
	default:
		return n.Numeric < o.Numeric
	}
}

func (n Identifier) IsUnknown() bool {
	return math.IsNaN(n.Numeric)
}

func (n Identifier) String() string {
	if n.IsUnknown() {
		return "?"
	} else {
		return fmt.Sprint(n.Numeric)
	}
}

func (n Identifier) MarshalJSON() ([]byte, error) {
	if n.IsUnknown() {
		return []byte("null"), nil
	} else {
		return []byte(fmt.Sprint(n)), nil
	}
}
