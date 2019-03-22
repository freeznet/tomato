package controllers

import (
	"github.com/freeznet/tomato/cloud"
	"github.com/freeznet/tomato/errs"
	"github.com/freeznet/tomato/rest"
	"github.com/freeznet/tomato/types"
	"github.com/freeznet/tomato/utils"
)

// CloudCodeController ...
type CloudCodeController struct {
	ClassesController
}

// HandleFind ...
// @router / [get]
func (c *CloudCodeController) HandleFind()  {
	response, err := rest.Find(c.Auth, "_JobSchedule", types.M{}, types.M{}, c.Info.ClientSDK)

	if err != nil {
		c.HandleError(err, 0)
		return
	}

	results := utils.A(response["results"])

	c.Data["json"] = results[0]
	c.ServeJSON()
}

// HandleGet ...
// @router /jobs [get]
func (c *CloudCodeController) HandleGet() {
	jobs := cloud.GetJobs()
	jobNames := []string{}
	for n := range jobs {
		jobNames = append(jobNames, n)
	}

	response, err := rest.Find(c.Auth, "_JobSchedule", types.M{}, types.M{}, c.Info.ClientSDK)

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

// HandleCreate ...
// @router / [create]
func (c *CloudCodeController) HandleCreate()  {
	jobs := cloud.GetJobs()
	if v, ok := c.JSONBody["jobName"]; ok {
		if _, ok := jobs[v.(string)]; !ok {
			c.HandleError(errs.E(errs.InternalServerError, "Cannot Schedule a job that is not deployed"), 0)
		}
	}

	c.ClassName = "_JobSchedule"
	c.ClassesController.HandleUpdate()
}

// HandleUpdate ...
// @router / [put]
func (c *CloudCodeController) HandleUpdate()  {
	jobs := cloud.GetJobs()
	if v, ok := c.JSONBody["jobName"]; ok {
		if _, ok := jobs[v.(string)]; !ok {
			c.HandleError(errs.E(errs.InternalServerError, "Cannot Schedule a job that is not deployed"), 0)
		}
	}

	c.ClassName = "_JobSchedule"
	c.ObjectID = c.Ctx.Input.Param(":objectId")
	c.ClassesController.HandleUpdate()
}

// HandleDelete ...
// @router / [delete]
func (c *CloudCodeController) HandleDelete()  {
	c.ClassName = "_JobSchedule"
	c.ClassesController.HandleDelete()
}

// Post ...
// @router / [post]
func (c *CloudCodeController) Post() {
	c.ClassesController.Post()
}

