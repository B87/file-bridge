package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "filer",
	Short: "Filer CLI",
	Long: `
Filer CLI interacts with files across file systems.
`,
}

func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func fatalIfError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// TODO: Find a better logging / console print system
type Logger struct {
	verbose bool
}

func NewLogger(verbose bool) *Logger {
	return &Logger{verbose: verbose}
}

func (l *Logger) Debug(v ...interface{}) {
	if l.verbose {
		fmt.Println(v...)
	}
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.verbose {
		fmt.Printf(format+"\n", v...)
	}
}

func (l *Logger) Print(v ...interface{}) {
	fmt.Println(v...)
}

func (l *Logger) Printf(format string, v ...interface{}) {
	fmt.Printf(format+"\n", v...)
}

func (l *Logger) Fatal(v ...interface{}) {
	fmt.Println(v...)
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	fmt.Println(v...)
	os.Exit(1)
}

func (l *Logger) PrintDebugGlobal() {
	l.Debug("-----------------------------------------")
	l.Debug("Global config:")
	l.Debug("Verbose: ", l.verbose)
	l.Debug("-----------------------------------------")
}

func init() {
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose, default false")
}
