package log

import (
    "testing"
    "bytes"
    "strings"
    "io"
    "io/ioutil"
    "regexp"
    "os"
)

const (
    defaultTestString = "This is some test output."
)

type outputChecker func(bool, io.Reader, *testing.T)
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

func checkForDate(expected bool, log io.Reader, t *testing.T) {
    contents, err := ioutil.ReadAll(log)
    if err != nil {
        t.Error("Unable to read log output:", err)
    }
    datePattern := `\s\d\d\d\d/\d\d/\d\d\s`
    actual, err := regexp.Match(datePattern, contents)
    if expected != actual || err != nil {
        t.Errorf("%s does not match expected date-containing behavior of (%v).", contents, expected)
    }
}

func checkForTime(expected bool, log io.Reader, t *testing.T) {
    contents, err := ioutil.ReadAll(log)
    if err != nil {
        t.Error("Unable to read log contents:", err)
    }
    timePattern := `\s\d\d:\d\d:\d\d\s`
    actual, err := regexp.Match(timePattern, contents)
    if expected != actual || err != nil {
        t.Errorf("%s does not match expected time-containing behavior of (%v).", contents, expected)
    }
}

func checkForMicroseconds(expected bool, log io.Reader, t *testing.T) {
    contents, err := ioutil.ReadAll(log)
    if err != nil {
        t.Error("Unable to read log contents:", err)
    }
    t.Logf("This is some contents with microseconds: %s", contents)
    t.Fail()   // to show contents
}

func checkForUTC(expected bool, log io.Reader, t *testing.T) {
    contents, err := ioutil.ReadAll(log)
    if err != nil {
        t.Error("Unable to read log contents:", err)
        t.Log("This is some contents with UTC.")
        t.Log(contents)
        t.Fail()
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
            },
        },
        {
            "microseconds",
            []assertion{
                assertion{checkForMicroseconds, true},
            },
        },
    }

    for _, tc := range testCases {
        func() {
            defer testInfoOut.Reset()

            SetInfoFlags(tc.flags)
            Info(defaultTestString)
            checkOutput(defaultTestString, testInfoOut.String(), t)
            for _, assertion := range tc.assertions {
                assertion.function(assertion.expected, &testInfoOut, t)
            }
        }()
    }
}
