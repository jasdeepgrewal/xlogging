package logging

//TODO: Need to read values from Json
//TODO: Rotating File needs to be handled when running
//TODO: Rotating File needs max file count
//TODO: Rotating File Rules File age
//TOOD: Rotating File Rules Size

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

const (
	//LogNone turns off all logs when assigned to loggingLevel
	LogNone uint64 = 0
	//LogInfo enables Log() output when assigned to loggingLevel
	LogInfo uint64 = 1 << 0
	//LogWarn enables Warn() output when assigned to loggingLevel
	LogWarn uint64 = 1 << 1
	//LogError enables Error() output when assigned to loggingLevel
	LogError uint64 = 1 << 2
	//LogAll enables all logs when assigned to loggingLevel
	LogAll uint64 = LogInfo | LogWarn | LogError

	prefixLog       = "Log::"
	prefixWarn      = "Warn::"
	prefixError     = "ERROR! "
	prefixBadFormat = "Bad_Format::"

	//StNone Style Type None
	StNone uint64 = 0
	//StLongFileName Style Type Long File Name. Overrides StShortFileName if set
	StLongFileName uint64 = 1 << 0
	//StShortFileName Style Type Short File Name
	StShortFileName uint64 = 1 << 1
	//StPrintStack Style Type Print Stack
	StPrintStack uint64 = 1 << 2
	//StLogToTerminal sets wether logs need to be sent to terminal when a log file is attached
	StLogToTerminal uint64 = 1 << 3
)

const (
	logFileExtension = ".log"
	logFolderPath    = "logs"
	logBaseFileName  = "Log"
)

//LoggingLevel defines which debugs are printed
var LoggingLevel = LogAll

var enabledStreams uint64

//StyleInfo style used for Info() and InfoS() outputs
var StyleInfo = StNone | StLogToTerminal

//StyleWarn  style used for Warn() outputs
var StyleWarn = StLongFileName | StLogToTerminal

//StyleError  style used for Error() outputs
var StyleError = StShortFileName | StPrintStack | StLogToTerminal

//LogNoFmtToTerminal sets weather calls to NoFmt functions should write to terminal if a logFile is present
var LogNoFmtToTerminal = true

var logFileAttached = false

//Setup opens or creates a new .log file and links it to the logger.
func init() {
	log.SetFlags(log.LstdFlags | log.LUTC)
	err := setupFileIO()

	if err != nil {
		fmt.Println("[LoggerInit] Error: Failed to setup logFile. " + err.Error())
		NoFmt("LOGGER SETUP: Log File Failed to attach!")
	} else {
		NoFmt("LOGGER SETUP")
	}

	if isUsingUTC() {
		NoFmt("Logger is using UTC time")
		NoFmtf("LocalTime %v", time.Now())
	} else {
		NoFmt("Logger is using Local time")
		NoFmtf("UTC Time %v", time.Now().UTC())
	}
}

func isUsingUTC() bool {
	return log.Flags()&log.LUTC == log.LUTC
}

func setupFileIO() error {
	//Get folder path of log file
	folderPath, err := getLogFolderFullPath()
	if err != nil {
		return err
	}

	//Check if folder exists
	_, err = os.Stat(folderPath)
	if os.IsNotExist(err) {
		//Folder not found, create one
		err = os.Mkdir(folderPath, os.ModeDir)
		if err != nil {
			return err
		}
	} else if err != nil {
		//Folder exists, but there is some other error.
		return err
	}

	logFileName := getLogFileName()
	logFilePath, errFilePath := getLogFilePath(logFileName)

	if errFilePath != nil {
		return errFilePath
	}

	fmt.Println("[LoggerInit] LogFilePath: " + logFilePath)

	err = rotateCurrentLogFile()
	if err != nil {
		return err
	}

	_, err = os.Stat(logFilePath)

	if err != nil {
		f, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			fmt.Println("[LoggerInit] Logger Log file attached SUCCESSFULLY")
			logFileAttached = true
			log.SetOutput(f)
		} else {
			logFileAttached = false
			f.Close()
		}
		return err
	}

	rotFileError := errorString{"[LoggerInit] Error! Log file already exists. This should not happen.\n RotateXX() should have renamed the existing file."}
	return rotFileError
}

type errorString struct {
	s string
}

func (e errorString) Error() string {
	return e.s
}

func getLogFileName() string {
	return getFileNameNoExt() + logFileExtension
}

func getFileNameNoExt() string {
	t := time.Now()
	if isUsingUTC() {
		t = t.UTC()
	}

	year, month, day := t.Date()
	var strBuffer bytes.Buffer
	strBuffer.WriteString(logBaseFileName)
	strBuffer.WriteString("_")
	strBuffer.WriteString(strconv.Itoa(day))
	strBuffer.WriteString("_")
	strBuffer.WriteString(strconv.Itoa(int(month)))
	strBuffer.WriteString("_")
	strBuffer.WriteString(strconv.Itoa(year))

	return strBuffer.String()
}

func getLogFilePath(fileName string) (string, error) {
	var strBuffer bytes.Buffer
	folderPath, err := getLogFolderFullPath()
	if err == nil {
		strBuffer.WriteString(folderPath)
		strBuffer.WriteString(string(os.PathSeparator))
		strBuffer.WriteString(fileName)
	}

	return strBuffer.String(), err
}

func getLogFolderFullPath() (string, error) {
	folderPath, err := filepath.Abs(logFolderPath)

	return folderPath, err
}

func rotateCurrentLogFile() error {
	var errFilePath error
	var err error

	currFileName := getLogFileName()
	newPath, errFilePath := getLogFilePath(currFileName)
	currentLogFilePath := newPath

	if errFilePath != nil {
		return errFilePath
	}

	_, err = os.Stat(currentLogFilePath)
	if err != nil {
		//No Need to Rotate, file does not exist
		fmt.Println("[LoggerInit] FileRotation: Rotation not needed, Log file does not exist currently at = " + currentLogFilePath)
		err = nil
		return err
	}

	counter := 1

	for err == nil {
		_, err = os.Stat(newPath)
		if err != nil {
			//fmt.Println("[LoggerInit] FileRotation: FileNotFound " + newPath)
		} else {
			currFileName = getFileNameNoExt() + "_" + strconv.Itoa(counter) + logFileExtension
			newPath, errFilePath = getLogFilePath(currFileName)
			if errFilePath != nil {
				return errFilePath
			}
			//fmt.Println("[LoggerInit] FileRotation: UpdatedCheckPath: " + newPath)
			counter++
		}
	}

	err = os.Rename(currentLogFilePath, newPath)

	return err
}

func printLog(logType, style uint64, v ...interface{}) {
	if checkFlag(style, StPrintStack) {
		printSpace()
	}

	prefix := []interface{}{getPrefix(logType, 3)}
	finalOut := append(prefix, v...)
	log.Println(finalOut...)

	if logFileAttached && checkFlag(style, StLogToTerminal) {
		fmt.Println(finalOut...)
	}

	if checkFlag(style, StPrintStack) {
		debug.PrintStack()
		printSpace()
	}
}

func printLogf(logType, style uint64, format string, v ...interface{}) {
	if checkFlag(style, StPrintStack) {
		printSpace()
	}

	prefix := getPrefix(logType, 3)
	prefix += " " + format + "\n"
	log.Printf(prefix, v...)

	if logFileAttached && checkFlag(style, StLogToTerminal) {
		fmt.Printf(prefix, v...)
	}

	if checkFlag(style, StPrintStack) {
		debug.PrintStack()
		printSpace()
	}
}

//Info prints using Println format to logInfo style log
func Info(v ...interface{}) {
	if canLog(LogInfo) {
		printLog(LogInfo, StyleInfo, v...)
	}
}

//Infof prints using Printf format to logInfo style log
func Infof(format string, v ...interface{}) {
	if canLog(LogInfo) {
		printLogf(LogInfo, StyleInfo, format, v...)
	}
}

//InfoS prints using Printf format to a seperate log stream of logInfo style. This can be enabled or disabled individually
func InfoS(stream byte, v ...interface{}) {
	if canLog(LogInfo) && checkBit(enabledStreams, stream) {
		streamIndex := []interface{}{stream, "|"}
		finalOut := append(streamIndex, v...)

		printLog(LogInfo, StyleInfo, finalOut)
	}
}

//InfoSf prints using Println format to a seperate log stream of logInfo style. This can be enabled or disabled individually
func InfoSf(stream byte, format string, v ...interface{}) {
	if canLog(LogInfo) && checkBit(enabledStreams, stream) {
		streamIndex := []interface{}{stream, "|"}
		finalOut := append(streamIndex, v...)

		printLogf(LogInfo, StyleInfo, format, finalOut)
	}
}

//Warn prints using Println format to logWarn style log
func Warn(v ...interface{}) {
	if canLog(LogWarn) {
		printLog(LogWarn, StyleWarn, v...)
	}
}

//Warnf prints using Printf format to logWarn style log
func Warnf(format string, v ...interface{}) {
	if canLog(LogWarn) {
		printLogf(LogWarn, StyleWarn, format, v...)
	}
}

//Error prints using Println format to logWarn style log
func Error(v ...interface{}) {
	if canLog(LogError) {
		printLog(LogError, StyleError, v...)
	}
}

//Errorf prints using Printf format to logWarn style log
func Errorf(format string, v ...interface{}) {
	if canLog(LogError) {
		printLogf(LogError, StyleError, format, v...)
	}
}

//NoFmt logs without any special formatting using Println
func NoFmt(v ...interface{}) {
	log.Println(v...)
	if logFileAttached && LogNoFmtToTerminal {
		fmt.Println(v...)
	}
}

//NoFmtf logs without any special formatting using Printf
func NoFmtf(format string, v ...interface{}) {
	log.Printf(format, v...)
	if logFileAttached && LogNoFmtToTerminal {
		fmt.Printf(format+"\n", v...)
	}
}

func canLog(logLv uint64) bool {
	return LoggingLevel&logLv == logLv
}

func printSpace() {
	var orgFlags = log.Flags()
	log.SetFlags(0)
	log.Printf("\n\n")
	log.SetFlags(orgFlags)
}

//Note: Should be called from Log(),Warn()... functions only,
//otherwise Linenumber and file will not be offset correctly.
func getPrefix(logType uint64, sourceDepth int) string {

	var strBuffer bytes.Buffer
	var style uint64
	switch logType {
	case LogInfo:
		strBuffer.WriteString(prefixLog)
		style = StyleInfo
	case LogWarn:
		strBuffer.WriteString(prefixWarn)
		style = StyleWarn
	case LogError:
		strBuffer.WriteString(prefixError)
		style = StyleError
	default:
		strBuffer.WriteString(prefixBadFormat)
	}

	fileNameType := 0
	if checkFlag(style, StLongFileName) {
		fileNameType = 2
	} else if checkFlag(style, StShortFileName) {
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

//PrintCaller prints the values of runtime.Caller
func PrintCaller() {
	pc, file, line, ok := runtime.Caller(2)

	if ok {
		var strBuffer bytes.Buffer

		strBuffer.WriteString("Caller Details\n")

		strBuffer.WriteString("\t\t")
		strBuffer.WriteString(runtime.FuncForPC(pc).Name())
		strBuffer.WriteString("()\n")

		strBuffer.WriteString("\t\t")
		strBuffer.WriteString(file)
		strBuffer.WriteString("(")
		strBuffer.WriteString(strconv.Itoa(line))
		strBuffer.WriteString(")\n")

		NoFmt(strBuffer.String())
	} else {
		Error("Logging: failed to get values from runtime.Caller()")
	}
}

//PrintProcessInfo prints information about the current running process
func PrintProcessInfo() {
	NoFmtf("Process pid[%v] ppid[%v] uid[%v]", os.Getpid(), os.Getppid(), os.Getuid())
}

//PrintOSInfo prints information about the OS and architecture
func PrintOSInfo() {
	NoFmtf("OS [%v] Arch [%v] MaxThreads[%v]", runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(-1))
	platform, family, version, _ := host.PlatformInformation()

	NoFmtf("Platform [%v] Family[%v] Version[%v]", platform, family, version)
}

//PrintCPUInfo prints detailed information about the cpu
func PrintCPUInfo() {
	v, _ := cpu.Info()

	for i := range v {
		NoFmtf("CPU[%v] Cores [%v] Mhz [%v] %v[%v]", v[i].CPU, v[i].Cores, v[i].Mhz, v[i].VendorID, v[i].ModelName)
	}
}

//PrintCPUUsage prints cpu used by the system.
func PrintCPUUsage() {
	if runtime.GOOS == "windows" {
		printCPUUsageWin32()
	} else {
		printCPUUsageOther()
	}
}

func printCPUUsageOther() {
	cpuPerf, err := cpu.Percent(time.Millisecond*10, true)
	if err != nil {
		log.Println(err)
	}

	for i := range cpuPerf {
		NoFmtf("Cpu[%v] %f%%", i, cpuPerf[i])
	}
}

func printCPUUsageWin32() {
	cpuPerf, err := cpu.PerfInfo()
	if err != nil {
		log.Println(err)
	}

	for i := range cpuPerf {
		NoFmtf("(Win) Cpu[%v] Name[%v] usage[%v]", i, cpuPerf[i].Name, cpuPerf[i].PercentProcessorTime)
	}
}

//PrintMemInfo prints information about system memory
func PrintMemInfo() {
	v, _ := mem.VirtualMemory()

	NoFmtf("MemSys(KB) Total: %v, Available:%v, Used:%f%%", v.Total/1024, v.Available/1024, v.UsedPercent)
}

//PrintMemUsage prints information about memory used by the application
func PrintMemUsage() {
	var mem = runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	NoFmtf("MemApp(KB) SysReserved %v, TotalAlloc %v, CurrentAlloc %v", mem.Sys/1024, mem.TotalAlloc/1024, mem.Alloc/1024)
}

//PrintDiskInfo prints information about the disk
func PrintDiskInfo() {
	partitions, _ := disk.Partitions(true)

	for i := range partitions {
		stat, _ := disk.Usage(partitions[i].Mountpoint)
		NoFmtf("Disk [%v] [%v] Total(MB)[%v] Free(MB)[%v] Used[%v%%]", partitions[i].Mountpoint, partitions[i].Fstype, stat.Total/1048576, stat.Free/1048576, stat.UsedPercent)
	}
}

//PrintSystemInfo prints the system information. CPU, memory, disk...
func PrintSystemInfo() {
	PrintOSInfo()
	PrintProcessInfo()
	PrintCPUInfo()
	PrintMemInfo()
	PrintDiskInfo()
}
