package etc

import "github.com/cloudtask/cloudtask-agent/cache"
import "github.com/cloudtask/libtools/gounits/logger"
import "github.com/cloudtask/libtools/gounits/system"
import "github.com/cloudtask/libtools/gzkwrapper"
import "gopkg.in/yaml.v2"

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

var (
	SystemConfig *Configuration = nil
)

var (
	ErrConfigFileNotFound      = errors.New("config file not found.")
	ErrConfigGenerateFailure   = errors.New("config file generated failure.")
	ErrConfigFormatInvalid     = errors.New("config file format invalid.")
	ErrConfigServerDataInvalid = errors.New("config server data invalid.")
)

// Configuration is exported
type Configuration struct {
	Version      string `yaml:"version" json:"version"`
	PidFile      string `yaml:"pidfile" json:"pidfile"`
	RetryStartup bool   `yaml:"retrystartup" json:"retrystartup"`
	CenterAPI    string `yaml:"centerapi" json:"centerapi"`

	Cluster struct {
		Hosts     string `yaml:"hosts" json:"hosts"`
		Root      string `yaml:"root" json:"root"`
		Device    string `yaml:"device" json:"device"`
		Runtime   string `yaml:"runtime" json:"runtime"`
		OS        string `yaml:"os" json:"os"`
		Platform  string `yaml:"platform" json:"platform"`
		Pulse     string `yaml:"pulse" json:"pulse"`
		Threshold int    `yaml:"threshold" json:"threshold"`
	} `yaml:"cluster" json:"cluster"`

	API struct {
		Hosts      []string `yaml:"hosts" json:"hosts"`
		EnableCors bool     `yaml:"enablecors" json:"enablecors"`
	} `yaml:"api" json:"api"`

	Cache struct {
		MaxJobs       int    `yaml:"maxjobs" json:"maxjobs"`
		SaveDirectory string `yaml:"savedirectory" json:"savedirectory"`
		AutoClean     bool   `yaml:"autoclean" json:"autoclean"`
		CleanInterval string `yaml:"cleaninterval" json:"cleaninterval"`
		PullRecovery  string `yaml:"pullrecovery" json:"pullrecovery"`
		FileSrverAPI  string `yaml:"filesrverapi" json:"filesrverapi"`
	} `yaml:"cache" json:"cache"`

	Logger struct {
		LogFile  string `yaml:"logfile" json:"logfile"`
		LogLevel string `yaml:"loglevel" json:"loglevel"`
		LogSize  int64  `yaml:"logsize" json:"logsize"`
	} `yaml:"logger" json:"logger"`
}

// New is exported
func New(file string) error {

	if file != "" {
		if !system.FileExist(file) {
			cloudtaskENV, _ := os.LookupEnv("CLOUDTASK")
			if cloudtaskENV == "" {
				return ErrConfigFileNotFound
			}
			fileName := filepath.Base(file)
			if _, err := system.FileCopy("./etc/"+cloudtaskENV+"/"+fileName, file); err != nil {
				return ErrConfigGenerateFailure
			}
			log.Printf("[#etc#] ENV CLOUDTASK: %s\n", cloudtaskENV)
		}
	}

	buf, err := readConfigurationFile(file)
	if err != nil {
		return fmt.Errorf("config read %s", err.Error())
	}

	conf := &Configuration{RetryStartup: true}
	if err := yaml.Unmarshal(buf, conf); err != nil {
		return ErrConfigFormatInvalid
	}

	if err = conf.parseEnv(); err != nil {
		return fmt.Errorf("config parse env %s", err.Error())
	}

	centerAPI, err := validateURL(conf.CenterAPI)
	if err != nil {
		return fmt.Errorf("config centerapi invalid, %s", err.Error())
	}
	conf.CenterAPI = centerAPI

	fileServerAPI, err := validateURL(conf.Cache.FileSrverAPI)
	if err != nil {
		return fmt.Errorf("config fileserverapi invalid, %s", err.Error())
	}
	conf.Cache.FileSrverAPI = fileServerAPI

	parseDefaultParmeters(conf)
	SystemConfig = conf
	log.Printf("[#etc#] version: %s\n", SystemConfig.Version)
	log.Printf("[#etc#] pidfile: %s\n", SystemConfig.PidFile)
	log.Printf("[#etc#] retrystartup: %s\n", strconv.FormatBool(SystemConfig.RetryStartup))
	log.Printf("[#etc#] centerapi: %s\n", SystemConfig.CenterAPI)
	log.Printf("[#etc#] cluster: %+v\n", SystemConfig.Cluster)
	log.Printf("[#etc#] APIlisten: %+v\n", SystemConfig.API)
	log.Printf("[#etc#] cache: %+v\n", SystemConfig.Cache)
	log.Printf("[#etc#] logger: %+v\n", SystemConfig.Logger)
	return nil
}

//PidFile is exported
func PidFile() string {

	if SystemConfig != nil {
		return SystemConfig.PidFile
	}
	return ""
}

//RetryStartup is exported
func RetryStartup() bool {

	if SystemConfig != nil {
		return SystemConfig.RetryStartup
	}
	return false
}

//CenterAPI is exported
func CenterAPI() string {

	if SystemConfig != nil {
		return SystemConfig.CenterAPI
	}
	return ""
}

//ClusterConfigs is exported
func ClusterConfigs() *gzkwrapper.WorkerArgs {

	if SystemConfig != nil {
		return &gzkwrapper.WorkerArgs{
			Hosts:     SystemConfig.Cluster.Hosts,
			Root:      SystemConfig.Cluster.Root,
			Device:    SystemConfig.Cluster.Device,
			Location:  SystemConfig.Cluster.Runtime,
			OS:        SystemConfig.Cluster.OS,
			Platform:  SystemConfig.Cluster.Platform,
			APIAddr:   SystemConfig.API.Hosts[0],
			Pulse:     SystemConfig.Cluster.Pulse,
			Threshold: SystemConfig.Cluster.Threshold,
		}
	}
	return nil
}

//CacheConfigs is exported
func CacheConfigs() *cache.CacheConfigs {

	if SystemConfig != nil {
		return &cache.CacheConfigs{
			MaxJobs:       SystemConfig.Cache.MaxJobs,
			SaveDirectory: SystemConfig.Cache.SaveDirectory,
			AutoClean:     SystemConfig.Cache.AutoClean,
			CleanInterval: SystemConfig.Cache.CleanInterval,
			PullRecovery:  SystemConfig.Cache.PullRecovery,
			FileServerAPI: SystemConfig.Cache.FileSrverAPI,
		}
	}
	return nil
}

//LoggerConfigs is exported
func LoggerConfigs() *logger.Args {

	if SystemConfig != nil {
		return &logger.Args{
			FileName: SystemConfig.Logger.LogFile,
			Level:    SystemConfig.Logger.LogLevel,
			MaxSize:  SystemConfig.Logger.LogSize,
		}
	}
	return nil
}

func readConfigurationFile(file string) ([]byte, error) {

	fd, err := os.OpenFile(file, os.O_RDONLY, 0777)
	if err != nil {
		return nil, err
	}

	defer fd.Close()
	buf, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func validateURL(rawURL string) (string, error) {

	pURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	scheme := pURL.Scheme
	if scheme == "" {
		scheme = "http"
	}

	rawURL = scheme + "://" + pURL.Host
	if pURL.Path != "" {
		rawURL = rawURL + path.Clean(pURL.Path)
	}

	if pURL.RawQuery != "" {
		rawURL = rawURL + "?" + pURL.RawQuery
	}
	return rawURL, nil
}

func parseDefaultParmeters(conf *Configuration) {

	if conf.Cluster.Pulse == "" {
		conf.Cluster.Pulse = "8s"
	}

	if conf.Cluster.Threshold == 0 {
		conf.Cluster.Threshold = 1
	}

	if len(conf.API.Hosts) == 0 {
		conf.API.Hosts = []string{":8600"}
	}

	if conf.Cache.MaxJobs == 0 {
		conf.Cache.MaxJobs = 255
	}

	if conf.Cache.CleanInterval == "" {
		conf.Cache.CleanInterval = "30m"
	}

	if conf.Cache.PullRecovery == "" {
		conf.Cache.PullRecovery = "300s"
	}

	if conf.Logger.LogLevel == "" {
		conf.Logger.LogLevel = "info"
	}

	if conf.Logger.LogSize == 0 {
		conf.Logger.LogSize = 20971520
	}
}
