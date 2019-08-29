package push

import (
	"fmt"
	"github.com/freeznet/tomato/push/umeng"
	"github.com/freeznet/tomato/types"
	"testing"
)

func TestUMengPushToAndroid(t *testing.T) {
	upush := newUMengPush()
	umeng.AndroidAppKey = "5d64990b3fc1956ef2000ff8"
	umeng.AndroidAppMasterSecret = "onsfyjygd5gyhguhyvcqnzjhfxq0yfrn"
	body := types.M{}
	data := types.M{}
	data["alert"] = "设备：xxxxx 发生 开 动作"
	data["title"] = "开关设备发生新的动作"
	body["data"] = data
	ret, err := upush.sendToAndroidDevices([]string{"AlUkx2vC0PhN5pEkJJ3F9c9qWzBYBXy6GnWlkbu8EpEd"}, body)
	fmt.Println(ret, err)
}
