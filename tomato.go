package tomato

import (
	"strings"

	"github.com/freeznet/tomato/config"
	_ "github.com/freeznet/tomato/routers"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/plugins/cors"
	"github.com/freeznet/tomato/controllers"
	"github.com/freeznet/tomato/livequery"
	"github.com/freeznet/tomato/orm"
)

// Run ...
func Run() {

	config.Validate()

	// 创建必要的索引
	orm.TomatoDBController.PerformInitialization()

	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	beego.ErrorController(&controllers.ErrorController{})

	allowMethodOverride()
	allowCrossDomain()

	beego.Run()
}

// RunLiveQueryServer 运行 LiveQuery 服务
func RunLiveQueryServer(args map[string]string) {
	livequery.Run(args)
}

func allowCrossDomain() {
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Authorization", "Access-Control-Allow-Origin",
			"Access-Control-Allow-Headers", "X-Parse-Master-Key", "X-Parse-REST-API-Key",
			"X-Parse-Javascript-Key", "X-Parse-Application-Id", "X-Parse-Client-Version", "X-Parse-Session-Token",
			"X-Requested-With", "X-Parse-Revocable-Session", "Content-Type"},
		AllowCredentials: true,
	}))
	beego.InsertFilter("*", beego.BeforeRouter, func(ctx *context.Context) {
		if ctx.Input.Method() == "OPTIONS" {
			ctx.Output.SetStatus(200)
			ctx.ResponseWriter.Started = true
		}
	})
}

func allowMethodOverride() {
	beego.InsertFilter("*", beego.BeforeRouter, func(ctx *context.Context) {
		if ctx.Input.Method() != "POST" {
			return
		}
		contentType := ctx.Input.Header("Content-type")
		if strings.HasPrefix(contentType, "text/plain") == false &&
			strings.HasPrefix(contentType, "application/json") == false {
			return
		}
		if ctx.Input.RequestBody == nil || len(ctx.Input.RequestBody) == 0 {
			return
		}
		// 通过字符串搜索查找 "_method": "GET" 中的 GET
		body := string(ctx.Input.RequestBody)
		// 查找 "_method"
		pos := strings.Index(body, `"_method"`)
		if pos == -1 {
			return
		}
		pos += len(`"_method"`)
		body = body[pos:]
		// 查找 : ，中间仅允许空格存在
		pos = strings.Index(body, `:`)
		if pos == -1 {
			return
		}
		s1 := body[:pos]
		if len(strings.Replace(s1, " ", "", -1)) > 0 {
			return
		}
		pos += len(`:`)
		body = body[pos:]
		// 查找 " ，中间仅允许空格存在
		pos = strings.Index(body, `"`)
		if pos == -1 {
			return
		}
		s2 := body[:pos]
		if len(strings.Replace(s2, " ", "", -1)) > 0 {
			return
		}
		pos += len(`"`)
		body = body[pos:]
		// 查找最后的 "
		pos = strings.Index(body, `"`)
		if pos == -1 {
			return
		}
		method := body[:pos]
		ctx.Request.Method = method
	})
}
