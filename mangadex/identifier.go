package mangadex

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Subchapters beyond nine are not properly represented by floating
// point numbers and will be sorted incorrectly. However, subchapter
// numbers will rarely ever go beyond single digits so this should be
// fine for the time being.

type Identifier struct {
	numeric  float64
	fallback string
}

func NewIdentifier(num string, fallback string) Identifier {
	f, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return Identifier{
			numeric:  math.Inf(1),
			fallback: strings.TrimSpace(fallback),
		}
	} else {
		return Identifier{
			numeric: f,
		}
	}
}

func (n Identifier) Equal(o Identifier) bool {
	switch {
	case !n.IsSpecial() && !o.IsSpecial():
		return n.numeric == o.numeric
	case !n.IsUnknown() && !o.IsUnknown():
		return n.fallback == o.fallback
	default:
		return false
	}
}

func (n Identifier) Less(o Identifier) bool {
	switch {
	case n.IsUnknown() && o.IsUnknown():
		return false
	case n.IsSpecial() && !o.IsSpecial():
		return false
	case !n.IsSpecial() && o.IsSpecial():
		return true
	case n.IsSpecial() && o.IsSpecial():
		return n.fallback < o.fallback
	default:
		return n.numeric < o.numeric
	}
}

func (n Identifier) LessOrEqual(o Identifier) bool {
	return n.Equal(o) || n.Less(o)
}

func (n Identifier) IsSpecial() bool {
	return math.IsInf(n.numeric, 1)
}

func (n Identifier) IsUnknown() bool {
	return n.IsSpecial() && len(n.fallback) == 0
}

func (n Identifier) String() string {
	switch {
	case n.IsUnknown():
		return "Unknown"
	case n.IsSpecial():
		return n.fallback
	default:
		return fmt.Sprint(n.numeric)
	}
}

func (n Identifier) MarshalJSON() ([]byte, error) {
	switch {
	case n.IsUnknown():
		return []byte("nil"), nil
	case n.IsSpecial():
		return json.Marshal(n.fallback)
	default:
		return json.Marshal(n.numeric)
	}
}

func (n Identifier) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}
