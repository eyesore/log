[![GoDoc](https://godoc.org/github.com/eyesore/log?status.svg)](https://godoc.org/github.com/eyesore/log)

# Eye Log

This is a simple logging package that allows the following:

- Configuration with environment variables
- 2 logging levels, with prefixed output and ability to turn off all logging
- Configurable output, at least what the standard logging library allows

## Usage

Out of the box you can have 2 levels of logging to stdout.

```
package main

import "github.com/eyesore/log"

func main() {
    log.Debug("Hello", "World")
    log.Debugf("Hello %s", "World")
    log.Info("Hello", "Again!")

    // runtime configuration
    log.SetInfoOut("/tmp/debug.log")
    log.Infof("Hello %s", "Again!") // writes to /tmp/debug.log
    log.SetDebugFlags("date,microseconds,shortfile")
    log.Debug("More information in output.")
    // [DEBUG]  2016/06/10 22:55:06.500921 main.go:19
}
```

You can also configure all of these options using environment variables.

### Environment Variables
The following environment variables can be set to change the logging behavior:

- EYELOG_DEBUG_OUT
- EYELOG_INFO_OUT

Both of these default to os.Stdout.  Setting the variables to a filename. Will cause the logs for that level to be written to that file.  If it does not exist, the file will be created, but if the directory does not exist, that output will be set to the default.

- EYELOG_LEVEL

Valid values are 0 (no logging), 1 (only INFO), and 2 (all levels)

#### Flags
Flags control the configuration of the log output.

- EYELOG_FLAGS_DEFAULT

Default flags that will be used if a configuration is not specified for a logger.

- EYELOG_FLAGS_DEBUG
- EYELOG_FLAGS_INFO

Configuration flags that will override the default flags for that log level.

The following flags are allowed:

- `date` - YYYY/MM/DD timestamp for each log entry
- `time` - HH:MM:SS timestamp for each log entry
- `microseconds` - HH:MM:SS.mmmmmm timestamp for each log entry.  Replaces "time"
- `shortfile` - Filename and line number from which the log entry was generated.  Always overrides longfile if both are present.
- `longfile` - Full path to file and line number from which the log entry was generated.
- `UTC` - Use UTC for timestamps instead of system time.

Multiple flags can be set as follows:

`EYELOG_FLAGS_DEFAULT=date,time,shortfile,UTC`
`EYELOG_FLAGS_DEBUG=microtime`

If no flags are configured, the default is `date,time`
There is currently no way to turn off all flags, but you can have just one if you like.

### Runtime Configuration
Environment variables are read at startup, but some methods are exported for changing logger configuration at runtime.

[See the documentation for details.](https://godoc.org/github.com/eyesore/log)





