// Copyright 2013 Dmitry Chestnykh. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package conflag is simple a wrapper around flag package that fills flag
// values from configuration files before parsing them from command-line
// arguments.
//
// Configuration files contains list of flags in the same format they would
// appear as command-line arguments, but without the leading dash:
//
// 	http=localhost:8080
//	play=false
//
// The order of loading configurations is:
//
// 	/etc/progname
//	$HOME/.progname
//
// These files are parsed before command-line arguments, so real arguments
// override flags from configuration file.
//
// 	$ mycmd -play=true
//
// TODO: Support Windows-specific paths.
//
//
// Use this package like you would normally use flag:
//
//	import (
//		...
//		flag "github.com/dchest/conflag"
//		...
//	)
//
// But before calling Parse(), set program name:
//
//	func main() {
//		flag.SetProgName("mycmd")
//		flag.Parse()
//		...
//	}
//
// If your program overwrites Usage variable, call
// SetUsage() instead.
package conflag

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

// Lookup returns the Flag structure of the named command-line flag,
// returning nil if none exists.
func Lookup(name string) *flag.Flag {
	return defaultSet.Lookup(name)
}

// Set sets the value of the named command-line flag.
func Set(name, value string) error {
	return defaultSet.Set(name, value)
}

// PrintDefaults prints to standard error the default values of all defined command-line flags.
func PrintDefaults() {
	defaultSet.PrintDefaults()
}

// SetUsage sets underlying flag Usage variable.
func SetUsage(usage func()) {
	flag.Usage = usage
}

// Usage calls the underlying flag Usage function.
func Usage() {
	flag.Usage()
}

// NFlag returns the number of command-line flags that have been set.
func NFlag() int { return defaultSet.NFlag() }

// Arg returns the i'th command-line argument.  Arg(0) is the first remaining argument
// after flags have been processed.
func Arg(i int) string {
	return defaultSet.Arg(i)
}

// NArg is the number of arguments remaining after flags have been processed.
func NArg() int { return defaultSet.NArg() }

// Args returns the non-flag command-line arguments.
func Args() []string { return defaultSet.Args() }

// BoolVar defines a bool flag with specified name, default value, and usage string.
// The argument p points to a bool variable in which to store the value of the flag.
func BoolVar(p *bool, name string, value bool, usage string) {
	defaultSet.BoolVar(p, name, value, usage)
}

// Bool defines a bool flag with specified name, default value, and usage string.
// The return value is the address of a bool variable that stores the value of the flag.
func Bool(name string, value bool, usage string) *bool {
	return defaultSet.Bool(name, value, usage)
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func IntVar(p *int, name string, value int, usage string) {
	defaultSet.IntVar(p, name, value, usage)
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func Int(name string, value int, usage string) *int {
	return defaultSet.Int(name, value, usage)
}

// Int64Var defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func Int64Var(p *int64, name string, value int64, usage string) {
	defaultSet.Int64Var(p, name, value, usage)
}

// Int64 defines an int64 flag with specified name, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
func Int64(name string, value int64, usage string) *int64 {
	return defaultSet.Int64(name, value, usage)
}

// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint  variable in which to store the value of the flag.
func UintVar(p *uint, name string, value uint, usage string) {
	defaultSet.UintVar(p, name, value, usage)
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The return value is the address of a uint  variable that stores the value of the flag.
func Uint(name string, value uint, usage string) *uint {
	return defaultSet.Uint(name, value, usage)
}

// Uint64Var defines a uint64 flag with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func Uint64Var(p *uint64, name string, value uint64, usage string) {
	defaultSet.Uint64Var(p, name, value, usage)
}

// Uint64 defines a uint64 flag with specified name, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func Uint64(name string, value uint64, usage string) *uint64 {
	return defaultSet.Uint64(name, value, usage)
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func StringVar(p *string, name string, value string, usage string) {
	defaultSet.StringVar(p, name, value, usage)
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func String(name string, value string, usage string) *string {
	return defaultSet.String(name, value, usage)
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func Float64Var(p *float64, name string, value float64, usage string) {
	defaultSet.Float64Var(p, name, value, usage)
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func Float64(name string, value float64, usage string) *float64 {
	return defaultSet.Float64(name, value, usage)
}

// DurationVar defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
func DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	defaultSet.DurationVar(p, name, value, usage)
}

// Duration defines a time.Duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
func Duration(name string, value time.Duration, usage string) *time.Duration {
	return defaultSet.Duration(name, value, usage)
}

// Var defines a flag with the specified name and usage string. The type and
// value of the flag are represented by the first argument, of type Value, which
// typically holds a user-defined implementation of Value. For instance, the
// caller could create a flag that turns a comma-separated string into a slice
// of strings by giving the slice the methods of Value; in particular, Set would
// decompose the comma-separated string into the slice.
func Var(value flag.Value, name string, usage string) {
	defaultSet.Var(value, name, usage)
}

// UserConfigFilePath returns user configuration file path ($HOME/.progname).
// If program name is not set, returns an empty string.
func UserConfigFilePath() string {
	if progName == "" {
		return ""
	}
	//TODO Proper Windows support.
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return filepath.Join(u.HomeDir, "."+progName)
}

// GlobalConfigFilePath returns user configuration file path (/etc/progname).
// If program name is not set, returns an empty string.
func GlobalConfigFilePath() string {
	if progName == "" {
		return ""
	}
	//TODO Proper Windows support.
	return filepath.Join("/etc/", progName)
}

// readConfig reads configuration file and returns a slice
// of arguments-formatted strings.
func readConfig(filename string) (args []string) {
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// Config doesn't exist, not an error.
			return nil
		}
		fmt.Fprintf(os.Stderr, "error opening config file %q: %s", filename, err)
		os.Exit(2)
	}
	defer f.Close()

	// Read each line, prefix it with "-" and put into args.
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		args = append(args, "-"+scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing config file %q: %s", filename, err)
		os.Exit(2)
	}
	return
}

// parseConfig parses configuration files.
func parseConfigs() {
	if args := readConfig(GlobalConfigFilePath()); args != nil {
		defaultSet.Parse(args)
	}
	if args := readConfig(UserConfigFilePath()); args != nil {
		defaultSet.Parse(args)
	}
}

// Parse parses the command-line flags from os.Args[1:].  Must be called
// after all flags are defined and before flags are accessed by the program.
func Parse() {
	// Parse config first if we have progName set.
	if progName != "" {
		parseConfigs()
	}
	// Now parse the command line.
	// Ignore errors; defaultSet is set for ExitOnError.
	defaultSet.Parse(os.Args[1:])
}

// Parsed returns true if the command-line flags have been parsed.
func Parsed() bool {
	return defaultSet.Parsed()
}

var progName string

// SetProgName sets program name, which is used for locating configuration file,
// and outputting usage or error information.
//
// If program name is not set, the package won't use configuration file, only
// command line arguments.
func SetProgName(name string) {
	if strings.ContainsRune(progName, filepath.Separator) {
		fmt.Println(filepath.Clean(progName))
		panic("conflag: SetProgName called with bad program name")
	}
	progName = name
}

// The default set of command-line flags, parsed from os.Args.
var defaultSet = NewFlagSet(os.Args[0], flag.ExitOnError)

// NewFlagSet returns a new, empty flag set with the specified name and
// error handling property.
func NewFlagSet(name string, errorHandling flag.ErrorHandling) *flag.FlagSet {
	return flag.NewFlagSet(name, errorHandling)
}
