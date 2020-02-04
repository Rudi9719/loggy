package loggy

import "samhofi.us/x/keybase"
import "time"
import "fmt"
import "os"

// Level of importance for a log message
type LogLevel int

const (
	// Info level logging
	Info LogLevel = 4
	// Debugging output
	Debug LogLevel = 5
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
	// Will set to true if OutFile is set
	toFile bool
	// Will set to true if KBTeam is set
	toKeybase bool
	// Will set to true if UseStdout is true
	toStdout bool
	// Output file for logging - Required for file output
	OutFile string
	// Keybase Team for logging - Required for Keybase output
	KBTeam string
	// Keybase Channel for logging - Optional for Keybase output
	KBChann string
	// Log level / verbosity (see LogLevel)
	Level LogLevel
	// Program name for Keybase logging - Required for Keybase output
	ProgName string
	// Use stdout  - Required to print to stdout
	UseStdout bool
}

// A basic Logger with options for logging to file, keybase or stdout.
// More functionality could be added within the internal handleLog() func.
type Logger struct {
	opts LogOpts
	k    *keybase.Keybase
	team keybase.Channel
}

// Generate string from type Log with severity prepended
func (msg Log) String() string {
	levels := [...]string{
		"StdoutOnly",
		"Critical",
		"Error",
		"Warning",
		"Info",
		"Debug"}
	return fmt.Sprintf("%s: %s", levels[msg.Level], msg.Msg)
}

// Generate a timestamp for non-Keybase logs
func timeStamp() string {
	now := time.Now()
	return now.Format("02Jan06 15:04:05.9999")
}

// Write log to file from LogOpts
func (l Logger) toFile(msg Log) {
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
func (l Logger) toKeybase(msg Log) {
	tag := ""
	if msg.Level <= 2 {
		tag = "@everyone "
	}
	output := fmt.Sprintf("[%s] %s%s",
		l.opts.ProgName, tag, msg.String())

	chat := l.k.NewChat(l.team)
	chat.Send(output)

}

// Write log to Stdout
func (l Logger) toStdout(msg Log) {
	output := fmt.Sprintf("[%s] %s",
		timeStamp(), msg.String())
	fmt.Println(output)
}

// Log Info shortcut from string
func (l Logger) LogInfo(msg string) {
	var logMsg Log
	logMsg.Level = Info
	logMsg.Msg = msg
	go handleLog(l, logMsg)
}

// Log Debug shortcut from string
func (l Logger) LogDebug(msg string) {
	var logMsg Log
	logMsg.Level = Debug
	logMsg.Msg = msg

	go handleLog(l, logMsg)
}

// Log Warning shortcut from string
func (l Logger) LogWarn(msg string) {
	var logMsg Log
	logMsg.Level = Warnings
	logMsg.Msg = msg
	go handleLog(l, logMsg)
}

// Log Error shortcut from string - Will notify Keybase users
func (l Logger) LogError(msg string) {
	var logMsg Log
	logMsg.Level = Errors
	logMsg.Msg = msg
	go handleLog(l, logMsg)
}

// Log Critical shortcut from string - Will notifiy Keybase users
func (l Logger) LogCritical(msg string) {
	var logMsg Log
	logMsg.Level = Critical
	logMsg.Msg = msg
	go handleLog(l, logMsg)
}

// Log Critical shortcut that terminates program
func (l Logger) LogPanic(msg string) {
	var logMsg Log
	logMsg.Level = Critical
	logMsg.Msg = msg
	handleLog(l, logMsg)
	os.Exit(-1)
}

// Log error type for compatibility - Will notify keybase users
func (l Logger) LogErrorType(e error) {
	var logMsg Log
	// Will set Level to Critical without terminating program
	logMsg.Level = Critical
	logMsg.Msg = e.Error()
	go handleLog(l, logMsg)
}

// Func to hack to add other logging functionality
func handleLog(l Logger, logMsg Log) {

	if logMsg.Level > l.opts.Level && logMsg.Level != 0 {
		return
	}
	if logMsg.Level == 0 {
		go l.toStdout(logMsg)
		return
	}
	if l.opts.toKeybase {
		go l.toKeybase(logMsg)
	}
	if l.opts.toFile {
		go l.toFile(logMsg)
	}
	if l.opts.toStdout {
		go l.toStdout(logMsg)
	}

}

// Log func, takes LogLevel and string and passes to internal handler.
func (l Logger) Log(level LogLevel, msg string) {
	var logMsg Log
	logMsg.Level = level
	logMsg.Msg = msg
	go handleLog(l, logMsg)
}

// LogMsg takes a type Log and passes it to internal handler.
func (l Logger) LogMsg(msg Log) {
	go handleLog(l, msg)
}

// Create a new logger instance and pass it
func NewLogger(opts LogOpts) Logger {
	if opts.Level == 0 {
		opts.Level = 2
	}
	var l Logger
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
