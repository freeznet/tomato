package controllers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/freeznet/tomato/errs"
	"github.com/freeznet/tomato/rest"
	"github.com/freeznet/tomato/types"
	"github.com/freeznet/tomato/utils"
)

var baseKeys = []string{"where", "distinct", "pipeline"}
var pipelineKeys = []string{"addFields",
	"bucket",
	"bucketAuto",
	"collStats",
	"count",
	"currentOp",
	"facet",
	"geoNear",
	"graphLookup",
	"group",
	"indexStats",
	"limit",
	"listLocalSessions",
	"listSessions",
	"lookup",
	"match",
	"out",
	"project",
	"redact",
	"replaceRoot",
	"sample",
	"skip",
	"sort",
	"sortByCount",
	"unwind"}

var allowedKeys map[string]int8

func init() {
	allowedKeys = make(map[string]int8, len(baseKeys)+len(pipelineKeys))
	for _, v := range baseKeys {
		allowedKeys[v] = 1
	}
	for _, v := range pipelineKeys {
		allowedKeys[v] = 1
	}
}

type AggregateController struct {
	ClassesController
}

func (s *AggregateController) Prepare() {
	s.ClassesController.Prepare()
	//if s.Ctx.ResponseWriter.Started == false {
	//	s.EnforceMasterKeyAccess()
	//}
}

// HandleFind 处理查找对象请求
// @router /:className [get]
func (c *AggregateController) HandleFind() {
	if c.ClassName == "" {
		c.ClassName = c.Ctx.Input.Param(":className")
	}

	// 获取查询参数，并组装
	options := types.M{}

	if c.Query["distinct"] != "" {
		options["distinct"] = c.Query["distinct"]
	} else if c.JSONBody != nil && c.JSONBody["distinct"] != nil {
		options["distinct"] = c.JSONBody["distinct"]
	}
	pipeline, err := getPipeline(c.JSONBody)
	if err != nil {
		c.HandleError(err, 0)
		return
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
	} else {
		for _, stage := range pipeline {
			if stage["$match"] != nil {
				where = utils.M(stage["$match"])
			}
		}
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

/* Builds a pipeline from the body. Originally the body could be passed as a single object,
 * and now we support many options
 *
 * Array
 *
 * body: [{
 *   group: { objectId: '$name' },
 * }]
 *
 * Object
 *
 * body: {
 *   group: { objectId: '$name' },
 * }
 *
 *
 * Pipeline Operator with an Array or an Object
 *
 * body: {
 *   pipeline: {
 *     group: { objectId: '$name' },
 *   }
 * }
 *
 */
func getPipeline(body types.M) ([]types.M, error) {
	var pipeline = []types.M{}

	if v, has := body["pipeline"]; has {
		if reflect.TypeOf(v).Kind() != reflect.Slice {
			tmp := utils.M(v)
			for k, vv := range tmp {
				pipeline = append(pipeline, types.M{k: vv})
			}
		} else {
			tmp := utils.A(v)
			for _, vv := range tmp {
				pipeline = append(pipeline, utils.M(vv))
			}
		}
	} else {
		pipeline = []types.M{body}
	}

	var result = []types.M{}

	for _, stage := range pipeline {
		var keys = []string{}
		for k := range stage {
			keys = append(keys, k)
		}
		if len(stage) != 1 {
			err := fmt.Sprintf("Pipeline stages should only have one key found %s", strings.Join(keys, ", "))
			return nil, errs.E(errs.InvalidJSON, err)
		}
		r, err := transformStage(keys[0], stage)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}

	return result, nil
}

func transformStage(stageName string, stage types.M) (types.M, error) {
	if _, has := allowedKeys[stageName]; !has {
		return nil, errs.E(errs.InvalidQuery, fmt.Sprintf("Invalid parameter for query: %s", stageName))
	}
	var stageValue = stage[stageName]
	var key = fmt.Sprintf("$%s", stageName)
	var result = types.M{}
	if stageName == "group" {
		if m := utils.M(stageValue); m != nil {
			if _, has := m["_id"]; has {
				return nil, errs.E(errs.InvalidQuery, "Invalid parameter for query: group. Please use objectId instead of _id")
			}
			if _, has := m["objectId"]; !has {
				return nil, errs.E(errs.InvalidQuery, "Invalid parameter for query: group. objectId is required")
			}
			m["_id"] = m["objectId"]
			delete(m, "objectId")
			result[key] = m
			return result, nil
		}
	} else {
		result[key] = stageValue
	}
	return result, nil
}
