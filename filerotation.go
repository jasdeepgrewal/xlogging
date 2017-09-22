package xlogging

import (
	"log"
	"os"
)

var splitRuleNewRun = true

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

	err = rotateCurrentLogFile()
	if err != nil {
		return err
	}

	_, err = os.Stat(logFilePath)

	if err != nil {
		f, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			//fmt.Println("[LoggerInit] Logger Log file attached SUCCESSFULLY")
			logFileAttached = true
			log.SetOutput(f)
		} else {
			logFileAttached = false
			f.Close()
		}
		return err
	}

	rotFileError := stdError{"[LoggerInit] Error! Log file already exists. This should not happen.\n RotateXX() should have renamed the existing file."}
	return rotFileError
}
