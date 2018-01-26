package api

import "github.com/cloudtask/cloudtask-agent/cache"
import "github.com/cloudtask/cloudtask-agent/etc"
import "github.com/cloudtask/common/models"
import "github.com/cloudtask/libtools/gzkwrapper"
import "github.com/gorilla/mux"

import (
	"net/http"
)

type handler func(c *Context) error

var routes = map[string]map[string]handler{
	"GET": {
		"/cloudtask/v2/_ping":        ping,
		"/cloudtask/v2/jobs":         getJobs,
		"/cloudtask/v2/jobs/{jobid}": getJob,
	},
	"POST": {
		"/cloudtask/v2/jobsalloc": postJobsAlloc,
	},
	"PUT": {
		"/cloudtask/v2/jobs/action": putJobAction,
	},
}

func NewRouter(enableCors bool, store Store) *mux.Router {

	router := mux.NewRouter()
	for method, mappings := range routes {
		for route, handler := range mappings {
			routemethod := method
			routepattern := route
			routehandler := handler
			wrap := func(w http.ResponseWriter, r *http.Request) {
				if enableCors {
					writeCorsHeaders(w, r)
				}
				c := NewContext(w, r, store)
				routehandler(c)
			}
			router.Path(routepattern).Methods(routemethod).HandlerFunc(wrap)
			if enableCors {
				optionsmethod := "OPTIONS"
				optionshandler := optionsHandler
				wrap := func(w http.ResponseWriter, r *http.Request) {
					if enableCors {
						writeCorsHeaders(w, r)
					}
					c := NewContext(w, r, store)
					optionshandler(c)
				}
				router.Path(routepattern).Methods(optionsmethod).HandlerFunc(wrap)
			}
		}
	}
	return router
}

func ping(c *Context) error {

	pangData := struct {
		AppCode      string               `json:"app"`
		NodeKey      string               `json:"key"`
		NodeData     *gzkwrapper.NodeData `json:"node"`
		SystemConfig *etc.Configuration   `json:"systemconfig"`
		ServerConfig *models.ServerConfig `json:"serverconfig"`
		Cache        struct {
			AllocVersion int `json:"allocversion"`
			JobsTotal    int `json:"jobstotal"`
		} `json:"cache"`
	}{
		AppCode:      c.Get("AppCode").(string),
		NodeKey:      c.Get("NodeKey").(string),
		NodeData:     c.Get("NodeData").(*gzkwrapper.NodeData),
		SystemConfig: c.Get("SystemConfig").(*etc.Configuration),
		ServerConfig: c.Get("ServerConfig").(*models.ServerConfig),
	}

	cache := c.Get("Cache").(*cache.Cache)
	pangData.Cache.AllocVersion = cache.GetAllocVersion()
	pangData.Cache.JobsTotal = cache.GetJobsCount()
	return c.JSON(http.StatusOK, pangData)
}

func optionsHandler(c *Context) error {

	c.WriteHeader(http.StatusOK)
	return nil
}
