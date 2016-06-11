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
    *f =  0//log.LstdFlags
    configuredFlags := strings.Split(flags, ",")
    for _, cf := range configuredFlags {
        cf = strings.Trim(cf, " ")
        if flagVal, ok := flagMap[cf]; ok {
            // flags that aren't in the flagmap are ignored
            *f = *f | flagVal
        }
    }
}

func SetLevel(l int8) {
    level = l
}

func SetDebugOut(f string) {
    err := setOutput(&debugOut, f)
    if err != nil {
        log.Println("Error assigning debug output:", err)
    }
    if debugLogger != nil {
        debugLogger.SetOutput(debugOut)
    }
}

func SetDebugOutDirect(w io.Writer) {
    debugOut = w
    if debugLogger != nil {
        debugLogger.SetOutput(w)
    }
}

func SetInfoOut(f string) {
    err := setOutput(&infoOut, f)
    if err != nil {
        log.Println("Error assigning info output:", err)
    }
    if infoLogger != nil {
        infoLogger.SetOutput(infoOut)
    }
}

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

// TODO Debug vs Debugln? seems unnecessary
func Debug(v ...interface{}) {
    if level < LevelDebug {
        return
    }

    createDebugLogger()
    debugLogger.Output(callDepth, fmt.Sprintln(v...))
}

func Debugf(format string, v ...interface{}) {
    if level < LevelDebug {
        return
    }
    createDebugLogger()
    debugLogger.Output(callDepth, fmt.Sprintf(format, v...))
}

func Info(v ...interface{}) {
    if level < LevelInfo {
        return
    }
    createInfoLogger()
    infoLogger.Output(callDepth, fmt.Sprintln(v...))
}
func Infof(format string, v ...interface{}) {
    if level < LevelInfo {
        return
    }
    createInfoLogger()
    infoLogger.Output(callDepth, fmt.Sprintf(format, v...))
}

// simply wrap standard fatal for now
func Fatal(v ...interface{}) {
    log.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
    log.Fatalf(format, v...)
}
