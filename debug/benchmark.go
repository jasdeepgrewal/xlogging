package debug

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"time"

	xlog "github.com/jasdeepgrewal/xlogging"
)

const (
	stepLength = 100
	stepCount  = 100
	stepsTotal = stepCount * stepLength
	stepWait   = time.Millisecond

	testString = "benchmark"

	//ToNanoSeconds value 1
	ToNanoSeconds int64 = 1
	//ToMicroSeconds converts nano seconds to micro seconds
	ToMicroSeconds int64 = 1000
	//ToMilliSeconds converts nano seconds to milli seconds
	ToMilliSeconds int64 = ToMicroSeconds * 1000
	//ToSeconds converts nano seconds to seconds
	ToSeconds int64 = ToMilliSeconds * 1000
)

//TimeFactor sets time unit to show in benchmark.
var TimeFactor = ToNanoSeconds

//StopWatch used to time things with Begin() and GetTime()
type StopWatch struct {
	timeStamp time.Time
}

//Begin starts recording time
func (s *StopWatch) Begin() {
	s.timeStamp = time.Now()
}

//GetTimeNanoSec returns time elaplsed since Begin in Nanoseconds
func (s *StopWatch) GetTimeNanoSec() int64 {
	elapsedTime := time.Since(s.timeStamp)
	return elapsedTime.Nanoseconds()
}

func benchmarkln(fn func(v ...interface{})) int64 {
	var timeDump int64
	sw := StopWatch{}
	counter := 1

	for i := 0; i < stepCount; i++ {
		sw.Begin()
		for j := 0; j < stepLength; j++ {
			fn(testString, " ", strconv.Itoa(counter))
			counter++
		}
		timeDump += sw.GetTimeNanoSec()
		time.Sleep(stepWait)
	}

	timeDump /= stepsTotal
	return timeDump
}

func benchmarkf(fn func(format string, v ...interface{})) int64 {
	var timeDump int64
	sw := StopWatch{}
	counter := 1

	for i := 0; i < stepCount; i++ {
		sw.Begin()
		for j := 0; j < stepLength; j++ {
			fn("%v %v", testString, strconv.Itoa(counter))
			counter++
		}
		timeDump += sw.GetTimeNanoSec()
		time.Sleep(stepWait)
	}

	timeDump /= stepsTotal
	return timeDump
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func printBenchmark(i interface{}, timens int64) {
	timens /= TimeFactor
	fmt.Printf("%v time %v %v. Checked %v x %v (%v) times.\n", getFunctionName(i), timens, getTimeUnit(), stepCount, stepLength, stepsTotal)
}

func getTimeUnit() string {
	switch TimeFactor {
	case ToNanoSeconds:
		return "ns"
	case ToMicroSeconds:
		return "microSecs"
	case ToMilliSeconds:
		return "millisecs"
	case ToSeconds:
		return "s"
	default:
		return "<No_Unit_Format>"
	}
}

//BenchmarkLogInfo checks prints timing for Info()
func BenchmarkLogInfo() {
	t := benchmarkln(xlog.Info)
	printBenchmark(xlog.Info, t)
}

//BenchmarkLogInfof checks prints timing for Infof()
func BenchmarkLogInfof() {
	t := benchmarkf(xlog.Infof)
	printBenchmark(xlog.Infof, t)
}

//BenchmarkLogWarn checks prints timing for Warn()
func BenchmarkLogWarn() {
	t := benchmarkln(xlog.Warn)
	printBenchmark(xlog.Warn, t)
}

//BenchmarkLogWarnf checks prints timing for Warnf()
func BenchmarkLogWarnf() {
	t := benchmarkf(xlog.Warnf)
	printBenchmark(xlog.Warnf, t)
}

//BenchmarkLogError checks prints timing for Error()
func BenchmarkLogError() {
	t := benchmarkln(xlog.Error)
	printBenchmark(xlog.Error, t)
}

//BenchmarkLogErrorf checks prints timing for Errorf()
func BenchmarkLogErrorf() {
	t := benchmarkf(xlog.Errorf)
	printBenchmark(xlog.Errorf, t)
}

//BenchmarkLogNoFmt checks prints timing for NoFmt()
func BenchmarkLogNoFmt() {
	t := benchmarkln(xlog.NoFmt)
	printBenchmark(xlog.NoFmt, t)
}

//BenchmarkLogNoFmtf checks prints timing for NoFmtf()
func BenchmarkLogNoFmtf() {
	t := benchmarkf(xlog.NoFmtf)
	printBenchmark(xlog.NoFmtf, t)
}

//BenchmarkAllLogs ...does what it says...
func BenchmarkAllLogs() {
	fnln := make([]func(...interface{}), 0, 4)
	fnf := make([]func(string, ...interface{}), 0, 4)

	fnln = append(fnln, xlog.Info)
	fnln = append(fnln, xlog.Warn)
	fnln = append(fnln, xlog.Error)
	fnln = append(fnln, xlog.NoFmt)

	fnf = append(fnf, xlog.Infof)
	fnf = append(fnf, xlog.Warnf)
	fnf = append(fnf, xlog.Errorf)
	fnf = append(fnf, xlog.NoFmtf)

	tln := make([]int64, len(fnln))
	tf := make([]int64, len(fnf))

	for i, v := range fnln {
		tln[i] = benchmarkln(v)
	}

	for i, v := range fnf {
		tf[i] = benchmarkf(v)
	}

	for i, v := range fnln {
		printBenchmark(v, tln[i])
	}

	for i, v := range fnf {
		printBenchmark(v, tf[i])
	}
}
