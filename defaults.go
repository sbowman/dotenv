package dotenv

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	StringType = iota
	StringSliceType
	IntType
	Float64Type
	BoolType
	DurationType
)

type descriptor struct {
	Var          string
	DataType     int
	DefaultValue interface{}
	Description  string
}

// Cache default values for environment variables.
var registered = make(map[string]descriptor)
var regMutex sync.RWMutex

// Register registers a default value for an environment variable.  When getting the value for that
// environment variable, if a value isn't set, the default is returned.
func Register(key string, defaultValue interface{}, description string) {
	var dataType int

	switch defaultValue.(type) {
	case string:
		dataType = StringType
	case []string:
		dataType = StringSliceType
	case int:
		dataType = IntType
	case float64:
		dataType = Float64Type
	case bool:
		dataType = BoolType
	case time.Duration:
		dataType = DurationType
	default:
		panic("invalid type")
	}

	registered[key] = descriptor{
		Var:          key,
		DataType:     dataType,
		DefaultValue: defaultValue,
		Description:  description,
	}
}

// Default returns the default setting set by the Register call.  Thread-safe.
func Default(key string) (descriptor, bool) {
	regMutex.RLock()
	defer regMutex.RUnlock()

	val, present := registered[key]

	return val, present
}

// Colorized output
var (
	keyColor     = color.New(color.FgYellow)
	typeColor    = color.New(color.FgCyan)
	defaultColor = color.New(color.FgWhite, color.Faint)
	descColor    = color.New(color.FgWhite)

	typeNames = map[int]string{
		StringType:      "string",
		StringSliceType: "[]string",
		IntType:         "integer",
		Float64Type:     "float",
		BoolType:        "boolean",
		DurationType:    "duration",
	}
)

// Help displays details about registered default variables.  May be called via a `--help`
// command-line parameter, or if some setting is invalid.  Produces colorized output to stdout.
func Help() {
	var keys []string
	var width, descWidth, defvalWidth int
	for key, d := range registered {
		keys = append(keys, key)

		if len(key) > width {
			width = len(key)
		}

		if len(d.Description) > descWidth {
			descWidth = len(d.Description)
		}

		w := len(fmt.Sprintf("%v", d.DefaultValue))
		if w > defvalWidth {
			defvalWidth = w
		}
	}

	termWidth, _, err := terminal.GetSize(0)
	if err != nil {
		termWidth = 80
	}

	if width+descWidth+defvalWidth+18 > termWidth {
		if defvalWidth > 20 {
			defvalWidth = 20
		}

		descWidth = termWidth - width - defvalWidth - 18
	}

	sort.Strings(keys)
	for _, key := range keys {
		d := registered[key]

		_, _ = keyColor.Print(pad(key, width))
		fmt.Print("  ")
		_, _ = typeColor.Print(pad(typeNames[d.DataType], 12))
		fmt.Print("  ")
		_, _ = descColor.Print(pad(d.Description, descWidth))
		fmt.Print("  ")
		_, _ = defaultColor.Println(pad(fmt.Sprintf("%v", d.DefaultValue), defvalWidth))
	}
}

func pad(val string, width int) string {
	if len(val) > width {
		return val[:width-3] + "..."
	}

	return val + strings.Repeat(" ", width-len(val))
}
