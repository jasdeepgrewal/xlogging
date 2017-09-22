package debug

import (
	"bytes"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"

	xlog "github.com/jasdeepgrewal/xlogging"
)

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

		xlog.NoFmt(strBuffer.String())
	} else {
		xlog.Error("Logging: failed to get values from runtime.Caller()")
	}
}

//PrintProcessInfo prints information about the current running process
func PrintProcessInfo() {
	xlog.NoFmtf("Process pid[%v] ppid[%v] uid[%v]", os.Getpid(), os.Getppid(), os.Getuid())
}

//PrintOSInfo prints information about the OS and architecture
func PrintOSInfo() {
	xlog.NoFmtf("OS [%v] Arch [%v] MaxThreads[%v]", runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(-1))
	platform, family, version, _ := host.PlatformInformation()

	xlog.NoFmtf("Platform [%v] Family[%v] Version[%v]", platform, family, version)
}

//PrintCPUInfo prints detailed information about the cpu
func PrintCPUInfo() {
	v, _ := cpu.Info()

	for i := range v {
		xlog.NoFmtf("CPU[%v] Cores [%v] Mhz [%v] %v[%v]", v[i].CPU, v[i].Cores, v[i].Mhz, v[i].VendorID, v[i].ModelName)
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
		xlog.NoFmtf("Cpu[%v] %f%%", i, cpuPerf[i])
	}
}

func printCPUUsageWin32() {
	cpuPerf, err := cpu.PerfInfo()
	if err != nil {
		log.Println(err)
	}

	for i := range cpuPerf {
		xlog.NoFmtf("(Win) Cpu[%v] Name[%v] usage[%v]", i, cpuPerf[i].Name, cpuPerf[i].PercentProcessorTime)
	}
}

//PrintMemInfo prints information about system memory
func PrintMemInfo() {
	v, _ := mem.VirtualMemory()

	xlog.NoFmtf("MemSys(KB) Total: %v, Available:%v, Used:%f%%", v.Total/1024, v.Available/1024, v.UsedPercent)
}

//PrintMemUsage prints information about memory used by the application
func PrintMemUsage() {
	var mem = runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	xlog.NoFmtf("MemApp(KB) SysReserved %v, TotalAlloc %v, CurrentAlloc %v", mem.Sys/1024, mem.TotalAlloc/1024, mem.Alloc/1024)
}

//PrintDiskInfo prints information about the disk
func PrintDiskInfo() {
	partitions, _ := disk.Partitions(true)

	for i := range partitions {
		stat, _ := disk.Usage(partitions[i].Mountpoint)
		xlog.NoFmtf("Disk [%v] [%v] Total(MB)[%v] Free(MB)[%v] Used[%v%%]", partitions[i].Mountpoint, partitions[i].Fstype, stat.Total/1048576, stat.Free/1048576, stat.UsedPercent)
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
