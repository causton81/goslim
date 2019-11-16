# goslim
An implementation of the Fitnesse Slim protocol in Go. This project is a plugin to http://fitnesse.org/FrontPage

## Requirements when launching FitNesse
```bash
# Ensure that `go version` works
export PATH=/path/to/golang-sdk/bin:$PATH

# Ensure that binaries installed with `go install` are on your path
export PATH=$GOPATH/bin:$PATH
```

Install `goslim` helper (see next section)
```
go install ./cmd/goslim
```

## What's the deal with the `goslim` binary in this module?
FitNesse ultimately has to spawn a program that you write with this module. FitNesse then speaks the SLiM protocol to a "SLiM server"
inside your program. Most of those details are handled by this module. Anyway, it would be tedious if you had to manually recompile your
SLiM server every time you update your fixtures. A nicer approach would be to set COMMAND_PATTERN like
`go run ./path/to/my/slim-server` and let Go recompile on-demand. Unfortunately, there are potential pathing issues related to the
working directory and path to the command.

If you run FitNesse from the same directory as your SLiM server then you might be able to get away with setting COMMAND_PATTERN to 
`go run ./path/to/your/server`.

An alternative is to use the `goslim` binary. It's a little wrapper that changes directory to the location of your Go module, then
executes `go run` for you. You can see examples of using it below. It currently does NOT interpolate environment variables.

As of this writing there are problems with special characters (like `&`) in COMMAND_PATTERN. At least on Windows the ampersand is
escaped somewhere to `&amp;` and I could not figure out a way around it.

# Run against the TwoMinuteExample
- Run FitNesse and browse to http://localhost:8001/FitNesse.UserGuide.TwoMinuteExample
- Edit the page and define `COMMAND_PATTERN` below TEST_SYSTEM (of course, your path to goslim repo will be different)
```
!define TEST_SYSTEM {slim}
!define COMMAND_PATTERN {goslim C:\Users\caust\git\goslim ./cmd/example}
```

# Integration Tests with FitNesse
These test results are from a patch of FitNesse https://github.com/causton81/fitnesse/tree/feat/goslim which I will try to push upstream.

**NOTE**: there are two mistakes that are easy to make and hard to notice:
1. if `FITNESSE_SLIM_SETUP` does not make it into the environment of the test runner, then the default built-in Java SLiM will be used
instead of goslim.
1. if `FITNESSE_SLIM_SETUP` does not start with `!define COMMAND_PATTERN {`, then the default Java SLiM will be used instead of SLiM.

As a sanity check, you should replace the gradlew command below with `bash -c 'echo $FITNESSE_SLIM_SETUP'` and verify the output is what
you want. I also recommend corrupting your `COMMAND_PATTERN` with a bad letter and then verifying that all of the tests *fail*.

## zsh on OSX 10.15.1
```bash
causton@CAUSTON-OSX fitnesse % which go
/usr/local/bin/go
causton@CAUSTON-OSX fitnesse % which goslim
/Users/causton/go/bin/goslim
causton@CAUSTON-OSX fitnesse % FITNESSE_SLIM_SETUP="\!define COMMAND_PATTERN {goslim $HOME/git/goslim ./cmd/responder}" ./gradlew test --tests HtmlSlimResponderTest

> Configure project :
Building FitNesse v20191116...

BUILD SUCCESSFUL in 0s
7 actionable tasks: 7 up-to-date
```

## Git Bash shell on Windows
Man, I hate Windows command line. Do yourself a favor and use Linux or OSX instead.

```bash
$ PATH=~/sdk/go1.13/bin:~/go/bin:$PATH FITNESSE_SLIM_SETUP='!'"define COMMAND_PATTERN {goslim $USERPROFILE\git\goslim ./cmd/responder}" ./gradlew test --tests HtmlSlimResponderTest

> Configure project :
Building FitNesse v20191116...

BUILD SUCCESSFUL in 35s
7 actionable tasks: 1 executed, 6 up-to-date

```
