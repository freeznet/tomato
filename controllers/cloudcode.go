package controllers

import (
	"github.com/freeznet/tomato/cloud"
	"github.com/freeznet/tomato/errs"
	"github.com/freeznet/tomato/rest"
	"github.com/freeznet/tomato/types"
	"github.com/freeznet/tomato/utils"
	"time"
)

// CloudCodeController ...
type CloudCodeController struct {
	ClassesController
}

func formatJobSchedule(jobSchedule types.M) types.M  {
	if _, ok := jobSchedule["startAfter"]; !ok {
		jobSchedule["startAfter"] = utils.TimetoString(time.Now())
	}
	return jobSchedule
}

func (c *CloudCodeController) validateJobSchedule()  {
	jobs := cloud.GetJobs()
	if v, ok := c.JSONBody["jobName"]; ok {
		if _, ok := jobs[v.(string)]; !ok {
			c.HandleError(errs.E(errs.InternalServerError, "Cannot Schedule a job that is not deployed"), 0)
		}
	}
}

//Prepare ...
func (c *CloudCodeController) Prepare()  {
	c.ClassesController.Prepare()
	if c.Ctx.ResponseWriter.Started == false {
		c.EnforceMasterKeyAccess()
	}
}

// GetJobs ...
// @router /jobs [get]
func (c *CloudCodeController) GetJobs()  {
	response, err := rest.Find(c.Auth, "_JobSchedule", types.M{}, types.M{}, c.Info.ClientSDK)

	if err != nil {
		c.HandleError(err, 0)
		return
	}

	results := utils.A(response["results"])

	c.Data["json"] = results[0]
	c.ServeJSON()
}

// GetJobsData ...
// @router /jobs/data [get]
func (c *CloudCodeController) GetJobsData() {
	jobs := cloud.GetJobs()
	jobNames := []string{}
	for n := range jobs {
		jobNames = append(jobNames, n)
	}

	response, err := rest.Find(c.Auth, "_JobSchedule", types.M{}, types.M{}, nil)

	if err != nil {
		c.HandleError(err, 0)
		return
	}

	results := utils.A(response["results"])
	resultsJobNames := []string{}
	for _, v := range results {
		result := utils.M(v)
		resultsJobNames = append(resultsJobNames, result["jobName"].(string))
	}


	c.Data["json"] = types.M{"in_use": resultsJobNames, "jobs": jobNames}
	c.ServeJSON()
}

// CreateJob ...
// @router /jobs [create]
func (c *CloudCodeController) CreateJob()  {
	c.validateJobSchedule()

	c.ClassName = "_JobSchedule"
	c.JSONBody = formatJobSchedule(c.JSONBody)
	c.ClassesController.HandleCreate()
}

// EditJob ...
// @router /jobs/:objectId [put]
func (c *CloudCodeController) EditJob()  {
	c.validateJobSchedule()

	c.ClassName = "_JobSchedule"
	c.ObjectID = c.Ctx.Input.Param(":objectId")
	c.JSONBody = formatJobSchedule(c.JSONBody)
	c.ClassesController.HandleUpdate()
}

// DeleteJob ...
// @router / [delete]
func (c *CloudCodeController) DeleteJob()  {
	c.ClassName = "_JobSchedule"
	c.ClassesController.HandleDelete()
}
