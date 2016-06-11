// This is a simple logging package that exposes configuration through environment
// variables.
package log

import (
    "github.com/kelseyhightower/envconfig"

    "log"
    "os"
    "io"
    "fmt"
    "strings"
)

const (
    OutStdout = "STDOUT"

    LevelDebug int8                      = 2
    LevelInfo int8                       = 1
    LevelNone int8                       = 0

    callDepth int                        = 2
)

type specification struct {
    DebugOut            string  `default:"STDOUT" envconfig:"multi_word_var"`
    InfoOut             string  `default:"STDOUT" envconfig:"multi_word_var"`
    Level               int8    `default:"2"`

    FlagsDefault        string  `envconfig:"multi_word_var"`
    FlagsInfo           string  `envconfig:"multi_word_var"`
    FlagsDebug          string  `envconfig:"multi_word_var"`
}

var (
    config                  specification
    debugLogger             *log.Logger
    infoLogger              *log.Logger

    debugOut                io.Writer
    infoOut                 io.Writer

    defaultFlags            int
    debugFlags              int
    infoFlags               int

    level                   int8

    flagMap = map[string]int {
        "date": log.Ldate,
        "time": log.Ltime,
        "microseconds": log.Lmicroseconds,
        "longfile": log.Llongfile,
        "shortfile": log.Lshortfile,
        "UTC": log.LUTC,
    }
)

func init() {
    readEnv()
}

func readEnv() {
    // parse the environment into the Spec - if there is an error
    // log it, but continue using defaults
    err := envconfig.Process("eyelog", &config)
    if err != nil {
        log.Println("Error processing logging configuration:")
        log.Println(err)
    }

    setOutput(&debugOut, config.DebugOut)
    setOutput(&infoOut, config.InfoOut)

    setFlags(&defaultFlags, config.FlagsDefault)
    setFlags(&debugFlags, config.FlagsDebug)
    setFlags(&infoFlags, config.FlagsInfo)

    // prevent default flags from being set to -1
    if defaultFlags == -1 {
        defaultFlags = log.LstdFlags
    }

    level = config.Level
}

func setOutput(w *io.Writer, out string) error {
    if out == OutStdout {
        *w = os.Stdout
        return nil
    }

    openFlags := os.O_WRONLY | os.O_APPEND | os.O_CREATE
    f, err := os.OpenFile(out, openFlags, 0666)
    if err != nil {
        return err
    }

    *w = f
    return nil
}

func setFlags(f *int, flags string) {
    // real flags won't be -1, we use this as a marker to use default flags
    if flags == "" {
        *f = -1
        return
    }
     // init
    *f = 0
    configuredFlags := strings.Split(flags, ",")
    for _, cf := range configuredFlags {
        cf = strings.Trim(cf, " ")
        if flagVal, ok := flagMap[cf]; ok {
            // flags that aren't in the flagmap are ignored
            *f = *f | flagVal
        }
    }
}

func createDebugLogger() {
    var flags int
    if debugFlags == -1 {
        flags = defaultFlags
    } else {
        flags = debugFlags
    }
    if debugLogger == nil {
        debugLogger = log.New(debugOut, "[DEBUG]\t", flags)
    }
}

func createInfoLogger() {
    var flags int
    if infoFlags == -1 {
        flags = defaultFlags
    } else {
        flags = infoFlags
    }
    if infoLogger == nil {
        infoLogger = log.New(infoOut, "[INFO]\t", flags)
    }
}

// Update the log level at runtime.
// Valid values are log.LevelDebug, log.LevelInfo, log.LevelNone
func SetLevel(l int8) {
    level = l
}

// SetDebugOut sets the debug output to a file at location
// You can pass log.OutStdout to log debug level to os.Stdout
func SetDebugOut(location string) {
    err := setOutput(&debugOut, location)
    if err != nil {
        log.Println("Error assigning debug output:", err)
    }
    if debugLogger != nil {
        debugLogger.SetOutput(debugOut)
    }
}

// SetDebugOutDirect sets the debug output directly to the given Writer w.
func SetDebugOutDirect(w io.Writer) {
    debugOut = w
    if debugLogger != nil {
        debugLogger.SetOutput(w)
    }
}

// SetDebugOut sets the info output to a file at location
// You can pass log.OutStdout to log info level to os.Stdout
func SetInfoOut(f string) {
    err := setOutput(&infoOut, f)
    if err != nil {
        log.Println("Error assigning info output:", err)
    }
    if infoLogger != nil {
        infoLogger.SetOutput(infoOut)
    }
}

// SetDebugOutDirect sets the info output directly to the given Writer w.
func SetInfoOutDirect(w io.Writer) {
    infoOut = w
    if infoLogger != nil {
        infoLogger.SetOutput(w)
    }
}

func SetDefaultFlags(flags string) {
    setFlags(&defaultFlags, flags)
    if defaultFlags == -1 {
        defaultFlags = log.LstdFlags
    }
}

// SetDebugFlags changes the content of the debug output with a comma-separated list of flags.
// Valid flags are date,time,microseconds,shortfile,longfile,UTC
func SetDebugFlags(flags string) {
    setFlags(&debugFlags, flags)
    if debugLogger != nil {
        if debugFlags == -1 {
            debugLogger.SetFlags(defaultFlags)
        } else {
            debugLogger.SetFlags(debugFlags)
        }
    }
}

// SetDebugFlags changes the content of the info output with a comma-separated list of flags.
// Valid flags are date,time,microseconds,shortfile,longfile,UTC
func SetInfoFlags(flags string) {
    setFlags(&infoFlags, flags)
    if infoLogger != nil {
        if infoFlags == -1 {
            infoLogger.SetFlags(defaultFlags)
        } else {
            infoLogger.SetFlags(infoFlags)
        }
    }
}

// Debug logs debug output.  Args are in the style of fmt.Println
func Debug(v ...interface{}) {
    // TODO Debug vs Debugln? seems unnecessary
    if level < LevelDebug {
        return
    }

    createDebugLogger()
    debugLogger.Output(callDepth, fmt.Sprintln(v...))
}

// Debugf logs debug output.  Args are in the style of fmt.Printf
func Debugf(format string, v ...interface{}) {
    if level < LevelDebug {
        return
    }
    createDebugLogger()
    debugLogger.Output(callDepth, fmt.Sprintf(format, v...))
}

// Info logs info output.  Args are in the style of fmt.Println
func Info(v ...interface{}) {
    if level < LevelInfo {
        return
    }
    createInfoLogger()
    infoLogger.Output(callDepth, fmt.Sprintln(v...))
}

// Infof logs info output.  Args are in the style of fmt.Printf
func Infof(format string, v ...interface{}) {
    if level < LevelInfo {
        return
    }
    createInfoLogger()
    infoLogger.Output(callDepth, fmt.Sprintf(format, v...))
}

// Fatal logs to stderr and stops program execution (os.Exit)
// Args are in the style of fmt.Print
func Fatal(v ...interface{}) {
    log.Fatal(v...)
}

// Fatalf logs to sterr and stops program execution (os.Exit)
// Args are in the style of fmt.Printf
func Fatalf(format string, v ...interface{}) {
    log.Fatalf(format, v...)
}
