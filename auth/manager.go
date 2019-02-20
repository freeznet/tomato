package auth

import (
	"github.com/freeznet/tomato/config"
	"github.com/freeznet/tomato/errs"
	"github.com/freeznet/tomato/types"
)

var providers map[string]Provider
var options map[string]types.M

// TODO 在这里添加第三方登录支持
func init() {
	providers = map[string]Provider{
		"anonymous":      anonymous{},
		"facebook":       facebook{},
		"github":         github{},
		"google":         google{},
		"instagram":      instagram{},
		"janraincapture": janraincapture{},
		"janrainengage":  janrainengage{},
		"linkedin":       linkedin{},
		"meetup":         meetup{},
		"spotify":        spotify{},
		"vkontakte":      vkontakte{},
		"twitter":        twitter{},
		"digits":         twitter{},
		"weibo":          weibo{},
		"qq":             qq{},
		"weixin":         weixin{},
		"baidu":          baidu{},
		"douban":         douban{},
		"yixin":          yixin{},
		"youdao":         youdao{},
		"weapp":		  weapp{},
	}
	options = map[string]types.M{
		"facebook": types.M{
			"appIds": []string{},
		},
		"janraincapture": types.M{
			"janrain_capture_host": "https://my-app.janraincapture.com",
		},
		"janrainengage": types.M{
			"api_key": "example",
		},
		"spotify": types.M{
			"appIds": []string{},
		},
		"weapp": types.M{
			"appid": "wx871e7bb109ba3bf6",
			"secret": "e360f37b4a1a8ba9f99dc45b36f34536",
		},
	}
}

// ValidateAuthData 验证第三方登录数据
func ValidateAuthData(provider string, authData types.M) error {
	if provider == "anonymous" && config.TConfig.EnableAnonymousUsers == false {
		//不支持 anonymous
		return errs.E(errs.UnsupportedService, "This authentication method is unsupported.")
	}
	defaultProvider := providers[provider]
	if defaultProvider == nil {
		// 不支持该方式
		return errs.E(errs.UnsupportedService, "This authentication method is unsupported.")
	}

	return defaultProvider.ValidateAuthData(authData, options[provider])
}

type anonymous struct{}

func (a anonymous) ValidateAuthData(authData types.M, option types.M) error {
	return nil
}

// Provider ...
type Provider interface {
	ValidateAuthData(types.M, types.M) error
}
