package api

import "github.com/gorilla/mux"

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"strings"
)

//ResolveJobBaseRequest is exported
func ResolveJobBaseRequest(c *Context) string {

	vars := mux.Vars(c.request)
	jobid := strings.TrimSpace(vars["jobid"])
	if len(jobid) == 0 {
		return ""
	}
	return jobid
}

//ResolveJobActionRequest is exported
func ResolveJobActionRequest(c *Context) *JobActionRequest {

	buf, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return nil
	}

	request := &JobActionRequest{}
	if err := json.NewDecoder(bytes.NewReader(buf)).Decode(request); err != nil {
		return nil
	}
	return request
}
