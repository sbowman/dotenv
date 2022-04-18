// Package dotenv manages the configuration of the application through a .env file and environment
// variables.
//
// The developer creates a .env file locally (and doesn't commit it to source control) with
// information about database connections, logging verbosity, and other settings he or she would
// like to override during development, and these settings will be loaded into environment variables
// when the application starts.
//
// In production, the container or virtual machine OS environment variables may be configured per
// deployment to override any settings relevant to that specific environment.  This leaves any
// containers free of configuration files and easily configured externally, as most container
// solutions provide configuration through environment variables.
package dotenv

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrBadUserFile returned when the .env file in the user's home directory is invalid.
	ErrBadUserFile = errors.New("unable to parse $HOME/.env file")

	// ErrBadLocalFile returned when the local .env file (in the same directory as the app) is
	// invalid.
	ErrBadLocalFile = errors.New("unable to parse .env file")
)

// Load the environment settings from:
//
// * the .env file in the startup directory
// * the .env file in the user's home directory
//
// like they are environment variables.  Any existing environment variables are overwritten.
func Load() error {
	if home, err := os.UserHomeDir(); err == nil {
		userEnv := path.Join(path.Clean(home), ".env")
		if exists(userEnv) {
			if err := process(userEnv); err != nil {
				return ErrBadUserFile
			}
		}
	}

	localEnv := ".env"
	if exists(localEnv) {
		if err := process(localEnv); err != nil {
			return ErrBadLocalFile
		}
	}

	return nil
}

// GetString returns the environment variable as a string value.  If the environment variable
// doesn't exist, returns the default value if present, otherwise a blank string.
func GetString(key string) string {
	if val, set := os.LookupEnv(key); set {
		return val
	}

	if descriptor, ok := Default(key); ok {
		if defaultValue, ok := descriptor.DefaultValue.(string); ok {
			return defaultValue
		}
	}

	return ""
}


// GetStringSlice returns the environment variable as a string slice value.  If the environment
// variable doesn't exist, returns the default value if present, otherwise a nil value.  Expects a
// environment variable value to be a comma-separated list of values.
func GetStringSlice(key string) []string {
	if val, set := os.LookupEnv(key); set {
		sliced := strings.Split(val, ",")
		return sliced
	}

	if descriptor, ok := Default(key); ok {
		if defaultValue, ok := descriptor.DefaultValue.([]string); ok {
			return defaultValue
		}
	}

	return nil
}

// GetInt returns the environment variable as an integer value.  If the environment variable doesn't
// exist or is not an integer, returns the default value if present, otherwise returns 0.
func GetInt(key string) int {
	if val, set := os.LookupEnv(key); set {
		if ival, err := strconv.Atoi(val); err == nil {
			return ival
		}
	}

	if descriptor, ok := Default(key); ok {
		if defaultValue, ok := descriptor.DefaultValue.(int); ok {
			return defaultValue
		}
	}

	return 0
}

// GetInt64 returns the environment variable as an int64 value.  If the environment variable doesn't
// exist or is not an int64, returns the default value if present, otherwise returns 0.
func GetInt64(key string) int64 {
	if val, set := os.LookupEnv(key); set {
		if ival, err := strconv.ParseInt(val, 10, 64); err == nil {
			return ival
		}
	}

	if descriptor, ok := Default(key); ok {
		switch defaultValue := descriptor.DefaultValue.(type) {
		case int:
			return int64(defaultValue)
		case int64:
			return defaultValue
		default:
			return 0
		}
	}

	return 0
}

// GetFloat64 returns the environment variable as an float64 value.  If the environment variable
// doesn't exist, returns the default value if present, otherwise returns 0.
func GetFloat64(key string) float64 {
	if val, set := os.LookupEnv(key); set {
		if fval, err := strconv.ParseFloat(val, 64); err == nil {
			return fval
		}
	}

	if descriptor, ok := Default(key); ok {
		if defaultValue, ok := descriptor.DefaultValue.(float64); ok {
			return defaultValue
		}
	}

	return 0
}

// GetBool returns the environment variable as a boolean value.  If the environment variable doesn't
// exist, returns the default value if present, otherwise returns false.
func GetBool(key string) bool {
	if val, set := os.LookupEnv(key); set {
		if strings.EqualFold(val, "true") {
			return true
		}
	}

	if descriptor, ok := Default(key); ok {
		if defaultValue, ok := descriptor.DefaultValue.(bool); ok {
			return defaultValue
		}
	}

	return false
}

// GetDuration returns the environment variable as an time.Duration value.  If the environment
// variable doesn't exist, returns the default value if present, otherwise returns 0.
func GetDuration(key string) time.Duration {
	if val, set := os.LookupEnv(key); set {
		if dval, err := time.ParseDuration(val); err == nil {
			return dval
		}
	}

	if descriptor, ok := Default(key); ok {
		if defaultValue, ok := descriptor.DefaultValue.(time.Duration); ok {
			return defaultValue
		}
	}

	return 0
}

func exists(filename string) bool {
	if info, err := os.Stat(filename); err == nil {
		if info.IsDir() {
			return false
		}

		return true
	}

	return false
}

// Process a file into environment variables.
func process(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	s := bufio.NewScanner(file)
	lineNo := -1

	for s.Scan() {
		line := s.Text()
		lineNo++

		if commentIdx := strings.Index(line, "#"); commentIdx == 0 {
			continue
		} else if commentIdx != -1 {
			line = line[0:commentIdx]
		}

		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("unable to parse line %s:%d", filename, lineNo)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" || value == "" {
			return fmt.Errorf("invalid environment variable assignment %s:%d", filename, lineNo)
		}

		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to assign %s value %s (%s:%d)", key, value, filename, lineNo)
		}
	}

	return nil
}
