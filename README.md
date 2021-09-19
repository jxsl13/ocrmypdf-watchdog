# ocrmypdf-watchdog

This is a wrapper around the python ocrmypdf toolkit that runts in a docker container.


## Requirements 

- `docker`

## Setup

Modify the `docker-compose.yml` file to have your needed volumes mounted to the correct host directory and `PUID`(UID of the output file) as well as the `PGID` (GID of the output file) set to whatever your user's uid/gid are.

## Startup

```shell
make
```

This creates a few intermediate images that should be cleaned up.
You can either do this manually or with `make clean`.
*BUT* be aware of the possible consequences that might unintentionally occur if you execute this.
You will be *WARNED* before continuing with the cleanup.

## What currently works

*Write* and *Create* events trigger the new file in the `in` directory to be checked for their file type.
Only proper PDF files are actually OCR'ed and then put into the target `out` directory.


## Environment variables

Additional environment variables that are not necessarily needed.

### OCRMYPDF_ARGS

This environment variable may contain all available flags that the `ocrmypdf` tool provides.

### PUID

User ID of the resulting OCR'd file

### PGID

Group ID of the resulting OCR'd file.

### CHMOD

File permissions of the resulting file.

### LOG_FLAGS

Integer value between 0 and 63

#### Possible values

```go
const (
    Ldate         = 1           // the date in the local time zone: 2009/01/23
    Ltime         = 2           // the time in the local time zone: 01:23:23
    Lmicroseconds = 4           // microsecond resolution: 01:23:23.123123.  assumes Ltime.
    Llongfile     = 8           // full file name and line number: /a/b/c/d.go:23
    Lshortfile    = 16          // final file name element and line number: d.go:23. overrides Llongfile
    LUTC          = 32          // if Ldate or Ltime is set, use UTC rather than the local time zone
    LstdFlags     = Ldate + Ltime// initial values for the standard logger
)
```

#### Example

```yamp
# Ldate + Ltime = 3
LOG_FLAGS: 3 # results in the log output '2020/07/15 14:55:56 Job finished with result successfully.'

```
