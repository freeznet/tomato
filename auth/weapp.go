package auth

import (
	"github.com/freeznet/tomato/errs"
	"github.com/freeznet/tomato/types"
	"github.com/freeznet/tomato/utils"
	"fmt"
)

type weapp struct{}

func (a weapp) ValidateAuthData(authData types.M, options types.M) error {
	// 具体接口参考： https://developers.weixin.qq.com/miniprogram/dev/api/api-login.html#wxloginobject
	host := "https://api.weixin.qq.com/sns/"
	path := "jscode2session?appid=" + utils.S(options["appid"]) + "&secret=" + utils.S(options["secret"])+ "&js_code=" + utils.S(authData["js_code"])+ "&grant_type=authorization_code"
	fmt.Println(path)
	data, err := request(host+path, nil)
	if err != nil {
		return errs.E(errs.ObjectNotFound, "Failed to validate this access token with Weixin WeApp.")
	}
	if code, ok := data["errcode"].(float64); ok && code == 0 {
		if data["errmsg"] != nil && utils.S(data["errmsg"]) == "ok" {
			return nil
		}
	}
	return errs.E(errs.ObjectNotFound, "Weixin Weapp auth is invalid for this user.")
}

