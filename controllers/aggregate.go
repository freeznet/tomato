package controllers

import (
	"encoding/json"

	"github.com/freeznet/tomato/errs"
	"github.com/freeznet/tomato/rest"
	"github.com/freeznet/tomato/types"
	"github.com/freeznet/tomato/utils"
)

type AggregateController struct {
	ClassesController
}

func (s *AggregateController) Prepare() {
	s.ClassesController.Prepare()
	if s.Ctx.ResponseWriter.Started == false {
		s.EnforceMasterKeyAccess()
	}
}

// HandleFind 处理查找对象请求
// @router /:className [get]
func (c *AggregateController) HandleFind() {
	if c.ClassName == "" {
		c.ClassName = c.Ctx.Input.Param(":className")
	}

	allowKeys := map[string]bool{
		"where":                   true,
		"distinct":                   true,
		"project":                   true,
		"match":                   true,
		"limit":                   true,
		"skip":                    true,
		"redact":                   true,
		"unwind":                   true,
		"group":                   true,
		"sort":                   true,
		"geoNear":                   true,
		"lookup":                   true,
		"out":                   true,
		"indexStats":                   true,
		"facet":                   true,
		"bucket":                   true,
		"bucketAuto":                   true,
		"sortByCount":                   true,
		"addFields":                   true,
		"replaceRoot":                   true,
		"count":                   true,
		"graphLookup":                   true,
	}
	for k := range c.Query {
		if allowKeys[k] == false {
			c.HandleError(errs.E(errs.InvalidQuery, "Invalid parameter for query: "+k), 0)
			return
		}
	}

	pipeline := types.M{}

	for k := range allowKeys {
		if c.Query[k] != "" {
			pipeline[k] = c.Query[k]
		} else if c.JSONBody != nil && c.JSONBody[k] != nil {
			pipeline[k] = utils.M(c.JSONBody[k])
		}
	}

	// 获取查询参数，并组装
	options := types.M{}

	if c.Query["distinct"] != "" {
		options["distinct"] = c.Query["distinct"]
	} else if c.JSONBody != nil && c.JSONBody["distinct"] != nil {
		options["distinct"] = c.JSONBody["distinct"]
	}

	options["pipeline"] = pipeline

	where := types.M{}
	if c.Query["where"] != "" {
		err := json.Unmarshal([]byte(c.Query["where"]), &where)
		if err != nil {
			c.HandleError(errs.E(errs.InvalidJSON, "where should be valid json"), 0)
			return
		}
	} else if c.JSONBody != nil && c.JSONBody["where"] != nil {
		where = utils.M(c.JSONBody["where"])
	}

	response, err := rest.Find(c.Auth, c.ClassName, where, options, c.Info.ClientSDK)
	if err != nil {
		c.HandleError(err, 0)
		return
	}
	if utils.HasResults(response) {
		results := utils.A(response["results"])
		for _, v := range results {
			result := utils.M(v)
			if result["sessionToken"] != nil && c.Info.SessionToken != "" {
				result["sessionToken"] = c.Info.SessionToken
			}
		}
	}

	c.Data["json"] = response
	c.ServeJSON()
}