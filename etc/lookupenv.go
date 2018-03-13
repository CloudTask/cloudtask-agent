package etc

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

//ParseEnv is exported
func (conf *Configuration) parseEnv() error {

	pidFile := os.Getenv("CLOUDTASK_PIDFILE")
	if pidFile != "" {
		conf.PidFile = pidFile
	}

	retryStartup := os.Getenv("CLOUDTASK_RETRYSTARTUP")
	if retryStartup != "" {
		value, err := strconv.ParseBool(retryStartup)
		if err != nil {
			return fmt.Errorf("CLOUDTASK_RETRYSTARTUP invalid, %s", err.Error())
		}
		conf.RetryStartup = value
	}

	useServerConfig := os.Getenv("CLOUDTASK_USESERVERCONFIG")
	if useServerConfig != "" {
		value, err := strconv.ParseBool(useServerConfig)
		if err != nil {
			return fmt.Errorf("CLOUDTASK_USESERVERCONFIG invalid, %s", err.Error())
		}
		conf.UseServerConfig = value
	}

	centerAPI := os.Getenv("CLOUDTASK_CENTERAPI")
	if centerAPI != "" {
		conf.CenterAPI = centerAPI
	}

	var err error
	//parse cluster env
	if err = parseClusterEnv(conf); err != nil {
		return err
	}

	//parse API env
	if err = parseAPIEnv(conf); err != nil {
		return err
	}

	//parse cache env
	if err = parseCacheEnv(conf); err != nil {
		return err
	}

	//parse logger env
	if err = parseLoggerEnv(conf); err != nil {
		return err
	}
	return nil
}

func parseClusterEnv(conf *Configuration) error {

	if clusterHosts := os.Getenv("CLOUDTASK_CLUSTER_HOSTS"); clusterHosts != "" {
		conf.Cluster.Hosts = clusterHosts
	}

	if clusterName := os.Getenv("CLOUDTASK_CLUSTER_NAME"); clusterName != "" {
		if ret := filepath.HasPrefix(clusterName, "/"); !ret {
			clusterName = "/" + clusterName
		}
		conf.Cluster.Root = clusterName
	}

	if clusterDevice := os.Getenv("CLOUDTASK_CLUSTER_DEVICE"); clusterDevice != "" {
		conf.Cluster.Device = clusterDevice
	}

	if clusterRuntime := os.Getenv("CLOUDTASK_CLUSTER_RUNTIME"); clusterRuntime != "" {
		conf.Cluster.Runtime = clusterRuntime
	}

	if clusterPulse := os.Getenv("CLOUDTASK_CLUSTER_PULSE"); clusterPulse != "" {
		if _, err := time.ParseDuration(clusterPulse); err != nil {
			return fmt.Errorf("CLOUDTASK_CLUSTER_PULSE invalid, %s", err.Error())
		}
		conf.Cluster.Pulse = clusterPulse
	}

	if clusterThreshold := os.Getenv("CLOUDTASK_CLUSTER_THRESHOLD"); clusterThreshold != "" {
		value, err := strconv.Atoi(clusterThreshold)
		if err != nil {
			return fmt.Errorf("CLOUDTASK_CLUSTER_THRESHOLD invalid, %s", err.Error())
		}
		conf.Cluster.Threshold = value
	}
	return nil
}

func parseAPIEnv(conf *Configuration) error {

	if apiHost := os.Getenv("CLOUDTASK_API_HOST"); apiHost != "" {
		hostIP, hostPort, err := net.SplitHostPort(apiHost)
		if err != nil {
			return fmt.Errorf("CLOUDTASK_API_HOST invalid, %s", err.Error())
		}
		if hostIP != "" {
			if _, err := net.LookupHost(hostIP); err != nil {
				return fmt.Errorf("CLOUDTASK_API_HOST invalid, %s", err.Error())
			}
		}
		conf.API.Hosts = []string{net.JoinHostPort(hostIP, hostPort)}
	}

	if enableCors := os.Getenv("CLOUDTASK_API_ENABLECORS"); enableCors != "" {
		value, err := strconv.ParseBool(enableCors)
		if err != nil {
			return fmt.Errorf("CLOUDTASK_API_ENABLECORS invalid, %s", err.Error())
		}
		conf.API.EnableCors = value
	}
	return nil
}

func parseCacheEnv(conf *Configuration) error {

	if maxJobs := os.Getenv("CLOUDTASK_CACHE_MAXJOBS"); maxJobs != "" {
		value, err := strconv.Atoi(maxJobs)
		if err != nil {
			return fmt.Errorf("CLOUDTASK_CACHE_MAXJOBS invalid, %s", err.Error())
		}
		conf.Cache.MaxJobs = value
	}

	if saveDirectory := os.Getenv("CLOUDTASK_CACHE_DIRECTORY"); saveDirectory != "" {
		conf.Cache.SaveDirectory = saveDirectory
	}

	if autoClean := os.Getenv("CLOUDTASK_CACHE_AUTOCLEAN"); autoClean != "" {
		value, err := strconv.ParseBool(autoClean)
		if err != nil {
			return fmt.Errorf("CLOUDTASK_CACHE_AUTOCLEAN invalid, %s", err.Error())
		}
		conf.Cache.AutoClean = value
	}

	if cleanInterval := os.Getenv("CLOUDTASK_CACHE_CLEANINTERVAL"); cleanInterval != "" {
		if _, err := time.ParseDuration(cleanInterval); err != nil {
			return fmt.Errorf("CLOUDTASK_CACHE_CLEANINTERVAL invalid, %s", err.Error())
		}
		conf.Cache.CleanInterval = cleanInterval
	}

	if pullRecovery := os.Getenv("CLOUDTASK_CACHE_PULLRECOVERY"); pullRecovery != "" {
		if _, err := time.ParseDuration(pullRecovery); err != nil {
			return fmt.Errorf("CLOUDTASK_CACHE_PULLRECOVERY invalid, %s", err.Error())
		}
		conf.Cache.PullRecovery = pullRecovery
	}

	if fileServerAPI := os.Getenv("CLOUDTASK_CACHE_FILESERVERAPI"); fileServerAPI != "" {
		conf.Cache.FileSrverAPI = fileServerAPI
	}
	return nil
}

func parseLoggerEnv(conf *Configuration) error {

	if logFile := os.Getenv("CLOUDTASK_LOG_FILE"); logFile != "" {
		conf.Logger.LogFile = logFile
	}

	if logLevel := os.Getenv("CLOUDTASK_LOG_LEVEL"); logLevel != "" {
		conf.Logger.LogLevel = logLevel
	}

	if logSize := os.Getenv("CLOUDTASK_LOG_SIZE"); logSize != "" {
		value, err := strconv.ParseInt(logSize, 10, 64)
		if err != nil {
			return fmt.Errorf("CLOUDTASK_LOG_SIZE invalid, %s", err.Error())
		}
		conf.Logger.LogSize = value
	}
	return nil
}
