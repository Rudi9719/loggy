package loggy

import "samhofi.us/x/keybase"
import "time"
import "fmt"
import "os"

// Level of importance for a log message
type LogLevel int

const (
	// Info level logging
	Info LogLevel = 5
	// Debugging output
	Debug LogLevel = 4
	// Will show if logger is set to warning
	Warnings LogLevel = 3
	// Errors will show by default
	Errors LogLevel = 2
	// Critical will show but can be silenced
	Critical LogLevel = 1
	// Special level for only showing via stdout
	StdoutOnly LogLevel = 0
)

// A basic Log struct with type and message
type Log struct {
	Level LogLevel
	Msg   string
}

// Logging options to be passed to NewLogger()
type LogOpts struct {
	toFile    bool
	toKeybase bool
	toStdout  bool
	OutFile   string
	KBTeam    string
	KBChann   string
	Level     LogLevel
	ProgName  string
	UseStdout bool
}

// A basic Logger with options for logging to file, keybase or stdout
type logger struct {
	opts LogOpts
	k    *keybase.Keybase
	team keybase.Channel
}

// Generate string from Log
func (msg Log) String() string {
	levels := [...]string{
		"StdoutOnly",
		"Critical",
		"Error",
		"Warning",
		"Debug",
		"Info"}
	return fmt.Sprintf("%s: %s", levels[msg.Level], msg.Msg)
}

// Generate a timestamp
func timeStamp() string {
	now := time.Now()
	return now.Format("02Jan06 15:04:05.9999")
}

// Write log to file
func (l logger) toFile(msg Log) {
	output := fmt.Sprintf("[%s] %s",
		timeStamp(), msg.String())

	f, err := os.OpenFile(l.opts.OutFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Unable to open logging file")
	}
	defer f.Close()
	if _, err := f.WriteString(fmt.Sprintf("%s\n", output)); err != nil {
		fmt.Println("Error writing output to logging file")
	}

}

// Send log to Keybase
func (l logger) toKeybase(msg Log) {
	output := fmt.Sprintf("[%s] %s",
		l.opts.ProgName, msg.String())
	chat := l.k.NewChat(l.team)
	chat.Send(output)

}

// Write log to Stdout
func (l logger) toStdout(msg Log) {
	output := fmt.Sprintf("[%s] %s",
		timeStamp(), msg.String())
	fmt.Println(output)
}

// Log Info
func (l logger) LogInfo(msg string) {
	var logMsg Log
	logMsg.Level = Info
	logMsg.Msg = msg
	handleLog(l, logMsg)
}

// Log Debug
func (l logger) LogDebug(msg string) {
	var logMsg Log
	logMsg.Level = Debug
	logMsg.Msg = msg

	handleLog(l, logMsg)
}

// Log Warning
func (l logger) LogWarn(msg string) {
	var logMsg Log
	logMsg.Level = Warnings
	logMsg.Msg = msg
	handleLog(l, logMsg)
}

// Log Error
func (l logger) LogError(msg string) {
	var logMsg Log
	logMsg.Level = Errors
	logMsg.Msg = msg
	handleLog(l, logMsg)
}

// Log Critical
func (l logger) LogCritical(msg string) {
	var logMsg Log
	logMsg.Level = Critical
	logMsg.Msg = msg

	handleLog(l, logMsg)
	os.Exit(3)
}

// Log error type
func (l logger) LogErrorType(e error) {
	var logMsg Log
	logMsg.Level = Critical
	logMsg.Msg = e.Error()
	handleLog(l, logMsg)
}

func handleLog(l logger, logMsg Log) {

	if logMsg.Level > l.opts.Level && logMsg.Level != 0 {
		return
	}
	if logMsg.Level == 0 {
		l.toStdout(logMsg)
		return
	}
	if l.opts.toKeybase {
		l.toKeybase(logMsg)
	}
	if l.opts.toFile {
		l.toFile(logMsg)
	}
	if l.opts.toStdout {
		l.toStdout(logMsg)
	}

}

func (l logger) Log(level LogLevel, msg string) {
	var logMsg Log
	logMsg.Level = level
	logMsg.Msg = msg
	handleLog(l, logMsg)
}

func (l logger) LogMsg(msg Log) {
	handleLog(l, msg)
}

// Create a new logger instance and pass it
func NewLogger(opts LogOpts) logger {
	if opts.Level == 0 {
		opts.Level = 2
	}
	var l logger
	if opts.KBTeam != "" {
		l.k = keybase.NewKeybase()
		var chann keybase.Channel
		if opts.KBChann != "" {
			chann.TopicName = opts.KBChann
			chann.MembersType = keybase.TEAM
		} else {
			chann.MembersType = keybase.USER
		}
		chann.Name = opts.KBTeam
		opts.toKeybase = true
		if !l.k.LoggedIn {
			fmt.Println("Not logged into keybase, but keybase option set.")
			os.Exit(-1)
		}
		l.team = chann
	}
	if opts.OutFile != "" {
		opts.toFile = true
	}
	opts.toStdout = opts.UseStdout
	l.opts = opts

	return l
}
