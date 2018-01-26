package api

//JobActionRequest is exported
type JobActionRequest struct {
	Runtime string `json:"runtime"`
	JobId   string `json:"jobid"`
	Action  string `json:"action"`
}
