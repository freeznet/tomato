package push

import (
	"fmt"
	"github.com/freeznet/tomato/config"
	"github.com/freeznet/tomato/push/umeng"
	"github.com/freeznet/tomato/types"
	"github.com/freeznet/tomato/utils"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type umengPushAdapter struct {
	validPushTypes []string
}

func newUMengPush() *umengPushAdapter {
	t := &umengPushAdapter{
		validPushTypes: []string{"ios", "android"},
	}
	umeng.AndroidAppKey = config.TConfig.UMengAndroidAppKey
	umeng.IOSAppKey = config.TConfig.UMengIOSAppKey
	umeng.AndroidAppMasterSecret = config.TConfig.UMengAndroidAppMasterSecret
	umeng.IOSAppMasterSecret = config.TConfig.UMengIOSAppMasterSecret
	return t
}

func (u *umengPushAdapter) sendToiOSDevices(tokens []string, body types.M) (umeng.Result, error) {
	if len(tokens) == 0 {
		return nil, errors.New("no token")
	}
	c := umeng.NewData(umeng.AppIOS)
	c.DeviceTokens = strings.Join(tokens, ",")
	c.Type = "listcast"
	if len(tokens) == 1 {
		c.Type = "unicast"
	}
	c.TimeStamp = time.Now().Unix()
	policy := umeng.Policy{}

	if t, ok := body["expiration_time"].(string); ok {
		timeInstance, _ := time.Parse("2006-01-02T15:04:05Z", t)
		policy.ExpireTime = timeInstance.Format("2006-01-02 15:04:05")
	}

	if t, ok := body["expiration_interval"].(int); ok {
		if pt, ok := body["push_time"].(string); ok {
			timeInstance, _ := time.Parse("2006-01-02T15:04:05Z", pt)
			policy.StartTime = timeInstance.Format("2006-01-02 15:04:05")
			timeInstance.Add(time.Duration(t) * time.Second)
			policy.ExpireTime = timeInstance.Format("2006-01-02 15:04:05")
		}
	}

	if pt, ok := body["push_time"].(string); ok {
		timeInstance, _ := time.Parse("2006-01-02T15:04:05Z", pt)
		policy.StartTime = timeInstance.Format("2006-01-02 15:04:05")
	}

	pushData := utils.M(body["data"])
	if pushData == nil {
		pushData = types.M{}
	}
	aps := umeng.IOSAps{}
	extras := make(map[string]string, 0)
	alert := types.M{}
	for key, v := range pushData {
		switch key {
		case "alert":
			alert["body"] = utils.S(v)
		case "badge":
			aps.Badge = fmt.Sprintf("%v", v)
		case "sound":
			aps.Sound = utils.S(v)
		case "content-available":
			if vv, ok := v.(int); ok {
				aps.ContentAvailable = string(vv)
			}
		case "category":
			aps.Category = utils.S(v)
		case "title":
			alert["title"] = utils.S(v)
		case "subtitle":
			alert["subtitle"] = utils.S(v)
		default:
			extras[key] = utils.S(v)
		}
	}
	aps.Alert = alert

	return c.Push(nil, aps, policy, extras)
}
func (u *umengPushAdapter) sendToAndroidDevices(tokens []string, body types.M) (umeng.Result, error) {
	if len(tokens) == 0 {
		return nil, errors.New("no token")
	}
	c := umeng.NewData(umeng.AppAndroid)
	c.DeviceTokens = strings.Join(tokens, ",")
	c.Type = "listcast"
	if len(tokens) == 1 {
		c.Type = "unicast"
	}
	c.TimeStamp = time.Now().Unix()
	policy := umeng.Policy{}

	if t, ok := body["expiration_time"].(string); ok {
		timeInstance, _ := time.Parse("2006-01-02T15:04:05Z", t)
		policy.ExpireTime = timeInstance.Format("2006-01-02 15:04:05")
	}

	if t, ok := body["expiration_interval"].(int); ok {
		if pt, ok := body["push_time"].(string); ok {
			timeInstance, _ := time.Parse("2006-01-02T15:04:05Z", pt)
			policy.StartTime = timeInstance.Format("2006-01-02 15:04:05")
			timeInstance.Add(time.Duration(t) * time.Second)
			policy.ExpireTime = timeInstance.Format("2006-01-02 15:04:05")
		}
	}

	if pt, ok := body["push_time"].(string); ok {
		timeInstance, _ := time.Parse("2006-01-02T15:04:05Z", pt)
		policy.StartTime = timeInstance.Format("2006-01-02 15:04:05")
	}

	pushData := utils.M(body["data"])
	if pushData == nil {
		pushData = types.M{}
	}
	payload := umeng.AndroidBody{}
	payload.DisplayType = "notification"
	extras := map[string]string{}
	for key, v := range pushData {
		switch key {
		case "alert":
			payload.Text = utils.S(v)
		case "uri":
			payload.AfterOpen = "go_app"
		case "title":
			payload.Ticker = utils.S(v)
			payload.Title = utils.S(v)
		default:
			extras[utils.S(key)] = utils.S(v)
		}
	}

	return c.Push(payload, nil, policy, extras)
}

func (u *umengPushAdapter) send(body types.M, installations types.S, pushStatus string) []types.M {
	deviceMap := classifyInstallations(installations, u.getValidPushTypes())
	results := []types.M{}

	loop := func(pushType string) {
		devices := deviceMap[pushType]
		if len(devices) == 0 {
			return
		}
		deviceTokens := []string{}

		for _, device := range devices {
			deviceTokens = append(deviceTokens, utils.S(device["deviceToken"]))
		}

		var status umeng.Result
		var err error

		switch pushType {
		case "ios":
			status, err = u.sendToiOSDevices(deviceTokens, body)
		case "android":
			status, err = u.sendToAndroidDevices(deviceTokens, body)
		}

		if err != nil {
			for _, device := range devices {
				result := types.M{
					"device":      device,
					"transmitted": false,
					"response":    map[string]string{"error": err.Error()},
				}
				results = append(results, result)
			}
			return
		}

		for index := range deviceTokens {
			var pushResult map[string]string
			if status != nil {
				pushResult = status
			} else {
				pushResult = nil
			}
			device := devices[index]

			resolution := types.M{
				"device":   device,
				"response": pushResult,
			}

			if pushResult == nil {
				resolution["transmitted"] = false
			} else {
				resolution["transmitted"] = true
			}

			results = append(results, resolution)
		}

	}

	for pushType := range deviceMap {
		loop(pushType)
	}

	return results

}

func (u *umengPushAdapter) getValidPushTypes() []string {
	return u.validPushTypes
}
