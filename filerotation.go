package xlogging

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var splitRuleNewRun = false

//Ignored if set to 0
var splitRuleSize = 0

//Ignored if set to 0
var splitRuleAge = 3600 //Seconds

//From seconds conversion conversion
const (
	secToMinute = 60
	secToHour   = secToMinute * 60
	secToDay    = secToHour * 24
)

const (
	logFileExtension = ".log"
	logFolderPath    = "logs" //Needs to be read from a json
	logBaseFileName  = "Log"  //Needs to be read from a json
)

//logFileAttached true if log file was attached successfully
var logFileAttached = false

var logFilePath = ""

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

	//fmt.Println("[LoggerInit] LogFilePath: " + logFilePath)

	if splitRuleNewRun {
		err = rotateCurrentLogFile()
		if err != nil {
			return err
		}

		//Check if the file exists at path
		_, err = os.Stat(logFilePath)
		if err == nil {
			err = stdError{"Log file already exists. This should not happen.\n RotateXX() should have renamed the existing file."}
			return err
		}
	} else {
		files, err := ioutil.ReadDir(folderPath)
		if err != nil {
			return err
		}

		if lenFiles := len(files); lenFiles > 0 {
			latestFile := getLatestFile(files)
			if latestFile != nil {
				//Set path to existing file
				logFilePath = folderPath + string(os.PathSeparator) + latestFile.Name()
				//fmt.Println("[LoggerInit] using existing file " + logFilePath)
			}
		}
		//Else: log file will be created with previous(see above in func) set logFilePath.
	}

	//Create or open the log file at logFilePath
	f, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err == nil {
		//fmt.Println("[LoggerInit] Logger Log file attached SUCCESSFULLY")
		logFileAttached = true
		log.SetOutput(f)
	} else {
		logFileAttached = false
		//fmt.Println("[LoggerInit] Logger failed to find specified file at path " + logFilePath)
		f.Close()
	}
	return err

}

func getLatestFile(files []os.FileInfo) os.FileInfo {
	index := -1
	var bestTime int64
	var currentTime int64
	for i := range files {
		if files[i].IsDir() {
			continue
		}

		currentTime = files[i].ModTime().Unix()
		if currentTime > bestTime {
			index = i
			bestTime = currentTime
		}
	}

	if index > -1 {
		return files[index]
	}

	return nil
}

func getLogFileName() string {
	return getFileNameNoExt() + logFileExtension
}

func getFileNameNoExt() string {
	t := time.Now()
	if useUTC {
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
		//fmt.Println("[LoggerInit] FileRotation: Rotation not needed, Log file does not exist currently at = " + currentLogFilePath)
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
