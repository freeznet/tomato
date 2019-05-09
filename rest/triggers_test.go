package rest

import (
	"github.com/freeznet/tomato/orm"
	"reflect"
	"testing"

	"github.com/freeznet/tomato/cloud"
	"github.com/freeznet/tomato/errs"
	"github.com/freeznet/tomato/types"
	"github.com/freeznet/tomato/utils"
)

func Test_maybeRunTrigger(t *testing.T) {
	var result types.M
	var err error
	var expect types.M
	var expectErr error
	/****************************************************************************************/
	cloud.BeforeSave("user", func(request cloud.TriggerRequest, response cloud.Response) {
		object := request.Object
		if username := utils.S(object["username"]); username != "" {
			object["username"] = username + "_tomato"
			response.Success(nil)
		} else {
			response.Error(1, "need a username")
		}
	})
	_, err = maybeRunTrigger(cloud.TypeBeforeSave, Master(), types.M{"className": "user"}, nil)
	expectErr = errs.E(1, "need a username")
	if reflect.DeepEqual(expectErr, err) == false {
		t.Error("expect:", expectErr, "result:", err)
	}
	result, err = maybeRunTrigger(cloud.TypeBeforeSave, Master(), types.M{"className": "user", "username": "joe"}, nil)
	expect = types.M{
		"object": types.M{
			"className": "user",
			"username":  "joe_tomato",
		},
	}
	if reflect.DeepEqual(expect, result) == false {
		t.Error("expect:", expect, "result:", result)
	}
	cloud.UnregisterAll()
}

func Test_maybeRunTriggerHeader(t *testing.T)  {
	var err error
	var expectErr error
	/****************************************************************************************/
	headers := map[string]string{
		"Connection": "keep-alive",
		"Host": "www.cn",
	}
	cloud.BeforeSave("user", func(request cloud.TriggerRequest, response cloud.Response) {
		object := request.Object
		if username := utils.S(object["username"]); username != "" {
			object["username"] = username + "_tomato"
			response.Success(nil)
		} else {
			response.Error(1, "need a username")
		}

		if reflect.DeepEqual(headers, request.Headers) == false {
			t.Error("expect:", headers, "result:", request.Headers)
		}
	})

	info := &types.RequestInfo{
		Headers: headers,
	}
	_, err = maybeRunTrigger(cloud.TypeBeforeSave, Master(info), types.M{"className": "user"}, nil)
	expectErr = errs.E(1, "need a username")
	if reflect.DeepEqual(expectErr, err) == false {
		t.Error("expect:", expectErr, "result:", err)
	}
	cloud.UnregisterAll()
}

func Test_maybeRunTriggerHeaderGet(t *testing.T)  {
	var object, schema types.M
	var className string
	var err error

	initEnv()
	className = "user"
	schema = types.M{
		"fields": types.M{
			"objectId": types.M{"type": "String"},
			"key": types.M{"type": "String"},
		},
	}
	orm.Adapter.CreateClass(className, schema)
	object = types.M{
		"objectId": "01",
		"key": "hello",
	}
	headers := map[string]string{
		"Connection": "keep-alive",
		"Host": "www.cn",
	}
	info := &types.RequestInfo{
		Headers: headers,
	}
	cloud.BeforeFind(className, func(request cloud.TriggerRequest, response cloud.Response) {
		if request.IsGet != true {
			t.Error("expect: isGet is true", "result:", request.IsGet)
		}
		if reflect.DeepEqual(headers, request.Headers) == false {
			t.Error("expect:", headers, "result:", request.Headers)
		}
	})
	orm.Adapter.CreateObject(className, schema, object)
	_, err = Get(Master(info), className, "01", types.M{}, nil)
	if err != nil {
		t.Error("Error: ", err)
	}
	orm.TomatoDBController.DeleteEverything()
}