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
	regMutex.Lock()
	defer regMutex.Unlock()

	var dataType int

	switch defaultValue.(type) {
	case string:
		dataType = StringType
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
		StringType:   "string",
		IntType:      "integer",
		Float64Type:  "float",
		BoolType:     "boolean",
		DurationType: "duration",
	}
)

// Display details about registered default variables.  May be called via a `--help` command-line
// parameter, or if some setting is invalid.  Produces colorized output to stdout.
func Help() {
	var keys []string
	var width, descWidth int
	for key, d := range registered {
		keys = append(keys, key)

		if len(key) > width {
			width = len(key)
		}

		if len(d.Description) > descWidth {
			descWidth = len(d.Description)
		}
	}

	if descWidth > 40 {
		descWidth = 40
	}

	sort.Strings(keys)

	termWidth, _, err := terminal.GetSize(0)
	if err != nil {
		termWidth = 80
	}

	for _, key := range keys {
		d := registered[key]

		_, _ = keyColor.Print(pad(key, width))
		fmt.Print("  ")
		_, _ = typeColor.Print(pad(typeNames[d.DataType], 12))
		fmt.Print("  ")
		_, _ = descColor.Print(pad(d.Description, descWidth))
		fmt.Print("    ")
		_, _ = defaultColor.Println(pad(fmt.Sprintf("%v", d.DefaultValue), termWidth - width - 62))
	}
}

func pad(val string, width int) string {
	if len(val) > width {
		return val[:width-3] + "..."
	}

	return val + strings.Repeat(" ", width-len(val))
}
