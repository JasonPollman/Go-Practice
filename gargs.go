package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Arg An argument type.
// This is an alias for interface{}
type Arg interface{}

// Coerces an argument value
func coerceArgument(value string) Arg {
	// Parse numeric values
	if f, e := strconv.ParseFloat(value, 64); e == nil {
		return f
	}

	// Parse hex values
	if strings.HasPrefix(value, "0x") {
		if i, e := strconv.ParseUint(strings.Replace(value, "0x", "", 1), 16, 32); e == nil {
			return i
		}
	}

	// Parse boolean values
	if b, e := strconv.ParseBool(value); e == nil {
		return b
	}

	return value
}

// Parse A CLI Argument Parser
// Parses command line arguments and flags into a <string, Arg> map where the Arg type is an
// alias for interface{}. All regular arguments will be stored under the underscore "_" map key,
// and flags and single character options will be keyed by name, respectively.
//
// Typical Usage: Parse(nil)
// If nil is passed, os.Args[1:] will be used as the array of arguments to parse. However,
// you can provide any string array to parse. For example: Parse(os.Args[2:4]), etc.
//
// Supported argument types:
// - Standard arguments (stored under _)
// - Flags (--flag [value] or --flag=[value])
// - Options (-o)
//
// All arguments will be coerced in the following order:
// - float64
// - hex string
// - boolean
// - string
//
// If -- is parsed, all flag and option parsing will halt and all remaining arguments will
// be put into the "_" []Arg key.
//
// -- Is typically an indicator to stop parsing arguments as they are intended for another program.
func Parse(args []string) (parsed map[string]Arg, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()

	// Use arguments without go executable filepath if nil is passed
	if args == nil {
		args = os.Args[1:]
	}

	// A flag to indicate that -- has been hit and no more flags/options should be parsed.
	var escaped = false
	var escapedArgs = []Arg{}

	// Stores a Map<string, Arg> hash of the parsed arguments.
	// This map will hold the parsed arguments
	var plain = []Arg{}
	parsed = map[string]Arg{"_": []Arg{}}

	// Trim, replace all "=" with tokens, and flatten out each argument set into a single array
	// This will convert, for example: [a, b, c, d=3, --f=4, -foo=bar] to [a, b, c, d=3, f, 4, foo, bar]
	var sanitized []string
	for _, v := range args {
		value := strings.Trim(v, " ")

		// Stop parsing arguments after the empty flag (--)
		if value == "--" {
			escaped = true
			continue
		}

		if escaped {
			escapedArgs = append(escapedArgs, coerceArgument(value))
		} else if strings.HasPrefix(value, "--") {
			// Flags (--)
			sanitized = append(sanitized, strings.Split(strings.Replace(value, "=", " ", 1), " ")...)
		} else if strings.HasPrefix(value, "-") {
			// Options (-)
			options := strings.Split(strings.Replace(value, "-", "", 1), "")
			for _, v := range options {
				parsed[v] = true
			}
		} else {
			// Regular ole command line arg
			sanitized = append(sanitized, v)
		}
	}

	// Parse each argument
	for i := 0; i < len(sanitized); i++ {
		current := sanitized[i]

		// If this isn't a flag, push the argument to the "plain"
		// argument array
		if !strings.HasPrefix(current, "--") {
			plain = append(plain, coerceArgument(current))
			continue
		}

		// This is a flag...
		key := strings.Replace(current, "--", "", 1)
		i++

		// Here we'll associate the current argument as the flag "name" (key) and the next
		// argument as the flag value.
		var next string
		if i == len(sanitized) {
			next = "true"
		} else {
			next = sanitized[i]

			// Next argument is a flag as well, so it cannot be the value for this flag
			// set current flag value to "true", which will later be parsed as a boolean.
			if strings.HasPrefix(next, "--") {
				next = "true"
				i--
			}
		}

		// Allow for the --no-[flag] syntax
		if strings.HasPrefix(key, "no-") && next == "true" {
			key = strings.Replace(key, "no-", "", 1)
			next = "false"
		}

		// If the flag already exists, create an array of flag values, otherwise
		// assign the coerced value to the mapping with the flag as the key.
		if val, ok := parsed[key]; !ok || val == true || val == false {
			parsed[key] = coerceArgument(next)
		} else {
			parsed[key] = []Arg{parsed[key], coerceArgument(next)}
		}
	}

	parsed["_"] = append(plain, escapedArgs...)
	return parsed, err
}

// ParseArgs A CLI Argument Parser
// An alias for Parse(nil). See gargs.Parse() for more information.
func ParseArgs() (parsed map[string]Arg, err error) {
	return Parse(nil)
}

func main() {
	v, _ := ParseArgs()
	fmt.Printf("%v\n", v)
}
