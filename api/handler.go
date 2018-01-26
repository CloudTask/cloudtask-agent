package api

import "github.com/cloudtask/cloudtask-agent/cache"
import "github.com/cloudtask/cloudtask-agent/driver"

import (
	"net/http"
)

func getJobs(c *Context) error {

	response := &ResponseImpl{}
	cache := c.Get("Cache").(*cache.Cache)
	jobBase := cache.GetJobs()
	respData := GetJobsBaseResponse{JobBase: jobBase}
	response.SetContent(ErrRequestSuccessed.Error())
	response.SetData(respData)
	return c.JSON(http.StatusOK, response)
}

func getJob(c *Context) error {

	response := &ResponseImpl{}
	jobid := ResolveJobBaseRequest(c)
	if jobid == "" {
		response.SetContent(ErrRequestResolveInvaild.Error())
		return c.JSON(http.StatusBadRequest, response)
	}

	cache := c.Get("Cache").(*cache.Cache)
	jobBase := cache.GetJob(jobid)
	if jobBase == nil {
		response.SetContent(ErrRequestNotFound.Error())
		return c.JSON(http.StatusNotFound, response)
	}

	respData := GetJobBaseResponse{JobBase: jobBase}
	response.SetContent(ErrRequestSuccessed.Error())
	response.SetData(respData)
	return c.JSON(http.StatusOK, response)
}

func postJobsAlloc(c *Context) error {

	return c.JSON(http.StatusAccepted, nil)
	/*buf, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		//500
		return nil
	}

	jobsAllocChanged := &models.JobsAllocChanged{}
	if err := json.NewDecoder(bytes.NewBuffer(buf)).Decode(jobsAllocChanged); err != nil {
		//400
		return nil
	}

	nodeKey := ""
	nodeData := c.Get("NodeData").(*gzkwrapper.NodeData)
	cache := c.Get("Cache").(*cache.Cache)
	cache.SetJobsAlloc(nodeKey)

	cache := c.Get("Cache").(*cache.Cache)
	cache.SetJobsAlloc()
	*/
}

func putJobAction(c *Context) error {

	response := &ResponseImpl{}
	request := ResolveJobActionRequest(c)
	if request == nil {
		response.SetContent(ErrRequestResolveInvaild.Error())
		return c.JSON(http.StatusBadRequest, response)
	}
	driver := c.Get("Driver").(*driver.Driver)
	driver.Action(request.JobId, request.Action)
	response.SetContent(ErrRequestAccepted.Error())
	return c.JSON(http.StatusAccepted, response)
}
