//Package xlogging logs messages to console and file.
//Has file rotation with age and size.
//Uses INFO,WARN... types to log output.
//Log settings can be change from Json.
package xlogging

//TODO: Rule: New File: On new Instance
//TODO: Rule: New File: Size
//TODO: Rule: New File: Date

//TODO: Rule: Delete old file: By Num
//TODO: Rule: Delete old file: By Date

//TODO: Add values to Json

//TODO: Supress logger internal option
//TOOD: Remove fmt.xx logs that are not errors

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"
)

//Log Types
const (
	//logNone turns off all logs when assigned to loggingLevel
	logNone uint64 = 0
	//logInfo enables Log() output when assigned to loggingLevel
	logInfo uint64 = 1 << 0
	//logWarn enables Warn() output when assigned to loggingLevel
	logWarn uint64 = 1 << 1
	//logError enables Error() output when assigned to loggingLevel
	logError uint64 = 1 << 2
	//logAll enables all logs when assigned to loggingLevel
	logAll uint64 = logInfo | logWarn | logError
)

//loggingLevel bitFlag that defines which log types are printed
var loggingLevel = logAll

//enabledStreams bitFlag that defines which InfoS/InfoSf logs are printed (0-63)
var enabledStreams uint64

//Log Prefix (Similar to Log4Net so highlighters can use it)
const (
	prefixLog       = "LOG::"
	prefixWarn      = "WARN::"
	prefixError     = "ERROR! "
	prefixBadFormat = "<Bad_Format>::"
)

//Log Style types
const (
	//stNone Style Type None
	stNone uint64 = 0
	//stLongFileName Style Type Long File Name. Overrides StShortFileName if set
	stLongFileName uint64 = 1 << 0
	//stShortFileName Style Type Short File Name
	stShortFileName uint64 = 1 << 1
	//stPrintStack Style Type Print Stack
	stPrintStack uint64 = 1 << 2
	//stLogToTerminal sets wether logs need to be sent to terminal when a log file is attached
	stLogToTerminal uint64 = 1 << 3
)

//styleInfo style used for Info() and InfoS() outputs
var styleInfo = stNone | stLogToTerminal

//styleWarn  style used for Warn() outputs
var styleWarn = stLongFileName | stLogToTerminal

//styleError  style used for Error() outputs
var styleError = stShortFileName | stPrintStack | stLogToTerminal

//logNoFmtToTerminal sets weather NoFmt() logs should write to terminal if a logFile is present
var logNoFmtToTerminal = true

var useUTC = true
var showTime = true

var showLoggerInitLogs = false

func init() {
	setupLogFlags()

	err := setupFileIO()

	if err != nil {
		fmt.Println("[LoggerInit] Error: Failed to setup logFile. " + err.Error())
		debug.PrintStack()
		NoFmt("LOGGER SETUP: Log File Failed to attach!")
	} else if showLoggerInitLogs {
		NoFmt("LOGGER SETUP")
	}

	if useUTC {
		if showLoggerInitLogs {
			NoFmt("Logger is using UTC time")
			NoFmtf("LocalTime %v", time.Now())
		}
	} else {
		if showLoggerInitLogs {
			NoFmt("Logger is using Local time")
			NoFmtf("UTC Time %v", time.Now().UTC())
		}
	}
}

func setupLogFlags() {
	logFlags := 0
	if showTime {
		logFlags |= log.Ldate | log.Ltime
		if useUTC {
			logFlags |= log.LUTC
		}
	}
	log.SetFlags(logFlags)
}

func printLog(logType, style uint64, v ...interface{}) {
	if checkFlag(style, stPrintStack) {
		printSpace()
	}

	prefix := []interface{}{getLinePrefix(logType, 3)}
	finalOut := append(prefix, v...)
	log.Println(finalOut...)

	//fmt.Printf("logFile Attached %v\n", logFileAttached)
	//fmt.Printf("checkFlag %v\n", checkFlag(style, stLogToTerminal))
	logToTerminal := logFileAttached && checkFlag(style, stLogToTerminal)
	if logToTerminal {
		fmt.Println(finalOut...)
	}

	if checkFlag(style, stPrintStack) {
		printStack(logToTerminal)
		printSpace()
	}
}

func printStack(logToTerminal bool) {
	byteArray := debug.Stack()
	n := len(byteArray)
	s := string(byteArray[:n])

	log.Println(s)

	if logToTerminal {
		fmt.Println(s)
	}
}

func printLogf(logType, style uint64, format string, v ...interface{}) {
	if checkFlag(style, stPrintStack) {
		printSpace()
	}

	prefix := getLinePrefix(logType, 3)
	prefix += " " + format + "\n"
	log.Printf(prefix, v...)

	logToTerminal := logFileAttached && checkFlag(style, stLogToTerminal)
	if logToTerminal {
		fmt.Printf(prefix, v...)
	}

	if checkFlag(style, stPrintStack) {
		printStack(logToTerminal)
		printSpace()
	}
}

//Info prints using Println format to logInfo style log
func Info(v ...interface{}) {
	if canLog(logInfo) {
		printLog(logInfo, styleInfo, v...)
	}
}

//Infof prints using Printf format to logInfo style log
func Infof(format string, v ...interface{}) {
	if canLog(logInfo) {
		printLogf(logInfo, styleInfo, format, v...)
	}
}

//InfoS prints using Printf format to a seperate log stream of logInfo style. This can be enabled or disabled individually
func InfoS(stream byte, v ...interface{}) {
	if canLog(logInfo) && checkBit(enabledStreams, stream) {
		streamIndex := []interface{}{stream, "|"}
		finalOut := append(streamIndex, v...)

		printLog(logInfo, styleInfo, finalOut)
	}
}

//InfoSf prints using Println format to a seperate log stream of logInfo style. This can be enabled or disabled individually
func InfoSf(stream byte, format string, v ...interface{}) {
	if canLog(logInfo) && checkBit(enabledStreams, stream) {
		streamIndex := []interface{}{stream, "|"}
		finalOut := append(streamIndex, v...)

		printLogf(logInfo, styleInfo, format, finalOut)
	}
}

//Warn prints using Println format to logWarn style log
func Warn(v ...interface{}) {
	if canLog(logWarn) {
		printLog(logWarn, styleWarn, v...)
	}
}

//Warnf prints using Printf format to logWarn style log
func Warnf(format string, v ...interface{}) {
	if canLog(logWarn) {
		printLogf(logWarn, styleWarn, format, v...)
	}
}

//Error prints using Println format to logWarn style log
func Error(v ...interface{}) {
	if canLog(logError) {
		printLog(logError, styleError, v...)
	}
}

//Errorf prints using Printf format to logWarn style log
func Errorf(format string, v ...interface{}) {
	if canLog(logError) {
		printLogf(logError, styleError, format, v...)
	}
}

//NoFmt logs without any special formatting using Println
func NoFmt(v ...interface{}) {
	log.Println(v...)
	if logFileAttached && logNoFmtToTerminal {
		fmt.Println(v...)
	}
}

//NoFmtf logs without any special formatting using Printf
func NoFmtf(format string, v ...interface{}) {
	log.Printf(format, v...)
	if logFileAttached && logNoFmtToTerminal {
		fmt.Printf(format+"\n", v...)
	}
}

func canLog(logLv uint64) bool {
	return loggingLevel&logLv == logLv
}

func printSpace() {
	var orgFlags = log.Flags()
	log.SetFlags(0)
	log.Printf("\n")
	log.SetFlags(orgFlags)
}

func getLinePrefix(logType uint64, sourceDepth int) string {

	var strBuffer bytes.Buffer
	var style uint64
	switch logType {
	case logInfo:
		strBuffer.WriteString(prefixLog)
		style = styleInfo
	case logWarn:
		strBuffer.WriteString(prefixWarn)
		style = styleWarn
	case logError:
		strBuffer.WriteString(prefixError)
		style = styleError
	default:
		strBuffer.WriteString(prefixBadFormat)
	}

	fileNameType := 0
	if checkFlag(style, stLongFileName) {
		fileNameType = 2
	} else if checkFlag(style, stShortFileName) {
		fileNameType = 1
	}

	if fileNameType > 0 {
		_, file, line, ok := runtime.Caller(sourceDepth)

		if ok {
			strBuffer.WriteString(" ")
			if fileNameType == 2 {
				strBuffer.WriteString(file)
			} else {
				strBuffer.WriteString(filepath.Base(file))
			}
			strBuffer.WriteString("(")
			strBuffer.WriteString(strconv.Itoa(line))
			strBuffer.WriteString(")>>")
		} else {
			strBuffer.WriteString("File_Format_Failed>>")
		}
	}

	return strBuffer.String()
}

//EnableStream enables or disables a InfoS() log output. Range(0,63)
func EnableStream(enable bool, stream byte) {
	if stream > 63 {
		stream = 63
	}

	if enable {
		enabledStreams |= 1 << stream
	} else {
		enabledStreams &= ^(1 << stream)
	}
}

//EnableStreams enables or disables multiple InfoS() log outputs. Range(0,63)
func EnableStreams(enable bool, streams ...byte) {
	for i := range streams {
		EnableStream(enable, streams[i])
	}
}

//EnableAllStreams enables/disables all InfoS log outputs.
func EnableAllStreams(enable bool) {
	for i := byte(0); i < 64; i++ {
		EnableStream(enable, i)
	}
}

func checkBit(value uint64, bit byte) bool {
	if (value>>bit)&1 == 1 {
		return true
	}

	return false
}

func checkFlag(value, flag uint64) bool {
	return value&flag == flag
}
