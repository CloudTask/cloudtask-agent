package cache

import "github.com/cloudtask/libtools/gounits/system"
import "github.com/cloudtask/libtools/gounits/utils"
import "github.com/cloudtask/libtools/gounits/logger"
import "github.com/cloudtask/common/models"

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

//DumpCleaner is exported
type DumpCleaner struct {
	Root     string
	Enabled  bool
	Duration time.Duration
	stopCh   chan struct{}
}

//NewDumpCleaner is exported
func NewDumpCleaner(configs *CacheConfigs) *DumpCleaner {

	duration, err := time.ParseDuration(configs.CleanInterval)
	if err != nil {
		duration, _ = time.ParseDuration("30m")
	}

	return &DumpCleaner{
		Root:     configs.SaveDirectory,
		Enabled:  configs.AutoClean,
		Duration: duration,
		stopCh:   nil,
	}
}

//Start is exported
//dumpCleaner GC start
func (dumpCleaner *DumpCleaner) Start() {

	if dumpCleaner.Enabled {
		if dumpCleaner.stopCh == nil {
			dumpCleaner.stopCh = make(chan struct{})
			go dumpCleaner.checkLoop()
		}
	}
}

//Stop is exported
//dumpCleaner GC stop
func (dumpCleaner *DumpCleaner) Stop() {

	if dumpCleaner.Enabled {
		if dumpCleaner.stopCh != nil {
			close(dumpCleaner.stopCh)
			dumpCleaner.stopCh = nil
		}
	}
}

func (dumpCleaner *DumpCleaner) checkLoop() {

	for {
		runTicker := time.NewTicker(dumpCleaner.Duration)
		select {
		case <-runTicker.C:
			{
				runTicker.Stop()
				jobs := readJobs(dumpCleaner.Root)
				jobfiles := []string{}
				for _, jobbase := range jobs {
					removeJobDirectories(dumpCleaner.Root, jobbase)
					if jobbase.FileName != "" {
						jobfiles = append(jobfiles, jobbase.FileName)
					}
				}
				removeJobFiles(dumpCleaner.Root, jobfiles)
			}
		case <-dumpCleaner.stopCh:
			{
				runTicker.Stop()
				return
			}
		}
	}
}

func readJobs(root string) []*models.JobBase {

	jobs := []*models.JobBase{}
	if fis, err := ioutil.ReadDir(root); err == nil {
		for _, fic := range fis {
			if fic.IsDir() {
				fpath := root + "/" + fic.Name() + "/job.json"
				if ret := system.FileExist(fpath); ret {
					fd, err := os.OpenFile(fpath, os.O_RDONLY, 0777)
					if err != nil {
						continue
					}
					jobbase := &models.JobBase{}
					if err := json.NewDecoder(fd).Decode(jobbase); err == nil {
						jobs = append(jobs, jobbase)
					}
					fd.Close()
				}
			}
		}
	}
	return jobs
}

func removeJobDirectories(root string, jobbase *models.JobBase) {

	jobroot, err := filepath.Abs(root + "/" + jobbase.JobId)
	if err != nil {
		return
	}

	fis, err := ioutil.ReadDir(jobroot)
	if err != nil {
		return
	}

	for _, fic := range fis {
		jobdirectory := fic.Name()
		if fic.IsDir() && jobbase.FileCode != jobdirectory {
			if err := os.RemoveAll(jobroot + "/" + jobdirectory); err != nil {
				logger.ERROR("[#cache#] dumpcleaner remove %s error:%s", jobdirectory, err.Error())
			} else {
				logger.INFO("[#cache#] dumpcleaner remove %s", jobdirectory)
			}
		}
	}
}

func removeJobFiles(root string, jobfiles []string) {

	filesroot, err := filepath.Abs(root + "/jobs")
	if err != nil {
		return
	}

	fis, err := ioutil.ReadDir(filesroot)
	if err != nil {
		return
	}

	for _, fic := range fis {
		jobfile := fic.Name()
		if !fic.IsDir() && !utils.Contains(jobfile, jobfiles) {
			if err := os.Remove(filesroot + "/" + jobfile); err != nil {
				logger.ERROR("[#cache#] dumpcleaner remove %s error:%s", jobfile, err.Error())
			} else {
				logger.INFO("[#cache#] dumpcleaner remove %s", jobfile)
			}
		}
	}
}
