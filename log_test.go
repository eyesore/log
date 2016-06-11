package log

import (
    "testing"
    "bytes"
    "strings"
    "io/ioutil"
    "regexp"
    "os"
)

const (
    defaultTestString = "This is some test output."
)

type outputChecker func(bool, []byte, *testing.T)
type assertion struct {
    function    outputChecker
    expected    bool
}

var (
    testDebugOut, testInfoOut    bytes.Buffer
)

func init() {
    SetDebugOutDirect(&testDebugOut)
    SetInfoOutDirect(&testInfoOut)
}

func checkOutput(expected, actual string, t *testing.T) {
    success := strings.Contains(actual, expected)
    if !success {
        t.Log("Debug output did not contain expected string.")
        t.Logf("Expected: %s", expected)
        t.Errorf("Got: %s", actual)
    }
}

func checkForDate(expected bool, log []byte, t *testing.T) {
    datePattern := `\s\d\d\d\d/\d\d/\d\d\s`
    actual, err := regexp.Match(datePattern, log)
    if expected != actual || err != nil {
        t.Errorf("%s does not match expected date-containing behavior of (%v).", log, expected)
    }
}

func checkForTime(expected bool, log []byte, t *testing.T) {
    timePattern := `\s\d\d:\d\d:\d\d\s`
    actual, err := regexp.Match(timePattern, log)
    if expected != actual || err != nil {
        t.Errorf("%s does not match expected time-containing behavior of (%v).", log, expected)
    }
}

func checkForMicroseconds(expected bool, log []byte, t *testing.T) {
    timePattern := `\s\d\d:\d\d:\d\d.\d\d\d\d\d\d\s`
    actual, err := regexp.Match(timePattern, log)
    if expected != actual || err != nil {
        t.Errorf("%s does not match expected time-containing behavior of (%v).", log, expected)
    }
}

func checkForUTC(expected bool, log []byte, t *testing.T) {
    // TJ - UTC just adjusts the time value to (gasp) UTC
    // if time flag is not set, it does nothing
    // TODO figure out how to test this
    t.Log("This is some log with UTC.")
    t.Logf("%s", log)
    t.Fail()
}

func checkForShortfile(expected bool, log []byte, t *testing.T) {
    shortfilePattern := `\slog_test.go:\d*:\s`
    actual, err := regexp.Match(shortfilePattern, log)
    if expected != actual || err != nil {
        t.Errorf("%s does not match expected file-containing behavior of (%v).", log, expected)
    }
}

func checkForLongfile(expected bool, log []byte, t *testing.T) {
    longfilePattern := `/log_test.go:\d*:\s`
    actual, err := regexp.Match(longfilePattern, log)
    if expected != actual || err != nil {
        t.Errorf("%s does not match expected file-containing behavior of (%v).", log, expected)
    }
}

func checkForFileOutput(expected, path string, t *testing.T) {
    actual := getFileContents(path, t)
    if !strings.Contains(actual, expected) {
        t.Log("File did not contain expected string.")
        t.Logf("Expected: %s", expected)
        t.Errorf("Got: %s", actual)
    }
}

func getFileContents(path string, t *testing.T) string {
    f, err := os.Open(path)
    if err != nil {
        t.Error("Error opening log file for reading:", err)
        return ""
    }
    defer f.Close()

    contents, err := ioutil.ReadAll(f)
    if err != nil {
        t.Error("Error reading file contents:", err)
    }
    return string(contents)
}

func TestDebug(t *testing.T) {
    defer testDebugOut.Reset()
    Debug("This", "is", "some", "test", "output.")
    checkOutput(defaultTestString, testDebugOut.String(), t)
}

func TestDebugf(t *testing.T) {
    defer testDebugOut.Reset()
    Debugf("%s %s %s test output.", "This", "is", "some")
    checkOutput(defaultTestString, testDebugOut.String(), t)
}

func TestInfo(t *testing.T) {
    defer testInfoOut.Reset()
    Info("This", "is", "some", "test", "output.")
    checkOutput(defaultTestString, testInfoOut.String(), t)
}

func TestInfof(t *testing.T) {
    defer testInfoOut.Reset()
    Infof("%s is %s %s output.", "This", "some", "test")
    checkOutput(defaultTestString, testInfoOut.String(), t)
}

func TestDebugToFile(t *testing.T) {
    path := "/tmp/debug.log"
    defer os.Remove(path)
    SetDebugOut(path)
    Debug("This", "is some test", "output.")
    checkForFileOutput(defaultTestString, path, t)
    Debugf("This should be %s.", "appended")
    checkForFileOutput(defaultTestString, path, t)
    checkForFileOutput("This should be appended.", path, t)
}

func TestDebugToExistingFile(t *testing.T) {
    path := "/tmp/debug.log"
    defer os.Remove(path)
    fileContents := "This should not be overwritten."
    err := ioutil.WriteFile(path, []byte(fileContents), 0666)
    if err != nil {
        t.Error("Unable to prepare file for test.")
    }
    SetDebugOut(path)
    Debug(defaultTestString)
    checkForFileOutput(fileContents, path, t)
    checkForFileOutput(defaultTestString, path, t)
}

func TestFlags(t *testing.T) {
    testCases := []struct{
        flags           string
        assertions      []assertion
    }{
        {
            "date",
            []assertion{
                assertion{checkForDate, true},
                assertion{checkForTime, false},
                assertion{checkForShortfile, false},
                assertion{checkForLongfile, false},
            },
        },
        {
            "time",
            []assertion{
                assertion{checkForDate, false},
                assertion{checkForTime, true},
                assertion{checkForShortfile, false},
                assertion{checkForLongfile, false},
            },
        },
        {
            "microseconds",
            []assertion{
                assertion{checkForMicroseconds, true},
            },
        },
        // {
        //     "UTC,time",
        //     []assertion{
        //         assertion{checkForUTC, true},
        //     },
        // },
        {
            "",
            []assertion{
                // empty string should use default flags log.LstdFlags
                assertion{checkForDate, true},
                assertion{checkForTime, true},
                assertion{checkForShortfile, false},
                assertion{checkForLongfile, false},
            },
        },
        {
            "shortfile",
            []assertion{
                assertion{checkForDate, false},
                assertion{checkForTime, false},
                assertion{checkForShortfile, true},
                assertion{checkForLongfile, false},
            },
        },
        {
            "longfile",
            []assertion{
                assertion{checkForDate, false},
                assertion{checkForTime, false},
                assertion{checkForShortfile, false},
                assertion{checkForLongfile, true},
            },
        },
        {
            "longfile,shortfile,date,microseconds,time",
            []assertion{
                assertion{checkForDate, true},
                assertion{checkForMicroseconds, true},
                assertion{checkForShortfile, true},
                // shortfile ALWAYS overrides longfile
                assertion{checkForLongfile, false},
            },
        },
    }

    for _, tc := range testCases {
        func() {
            defer testInfoOut.Reset()
            t.Logf("Testing flag set: %s", tc.flags)
            SetInfoFlags(tc.flags)
            Info(defaultTestString)
            contents, err := ioutil.ReadAll(&testInfoOut)
            if err != nil {
                t.Error("Unable to read log output:", err)
            }
            checkOutput(defaultTestString, string(contents), t)
            for _, assertion := range tc.assertions {
                assertion.function(assertion.expected, contents, t)
            }
        }()
    }
}
