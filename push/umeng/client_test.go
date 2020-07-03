package umeng_test

import (
	"testing"

	"github.com/freeznet/tomato/push/umeng"
	"github.com/stretchr/testify/assert"
)

var data *umeng.Data

func init() {
	data = umeng.NewData(umeng.AppAndroid)
}

func TestStatus(t *testing.T) {
	_, err := data.Status()
	assert.Nil(t, err)
}

func TestCancel(t *testing.T) {
	_, err := data.Cancel()
	assert.Nil(t, err)
}

func TestUpload(t *testing.T) {
	_, err := data.Upload()
	assert.Nil(t, err)
}

func TestPush(t *testing.T) {
	body := umeng.AndroidBody{}
	body.DisplayType = "message"
	body.Custom = "test"
	body.Title = "1212"

	extra := make(map[string]string, 0)
	extra["key1"] = "1212x"
	extra["key2"] = "12123323"
	_, err := data.Push(body, nil, nil, extra)
	assert.Nil(t, err)
}
