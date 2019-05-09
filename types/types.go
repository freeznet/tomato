package types

// M ...
type M map[string]interface{}

// S ...
type S []interface{}

// RequestInfo http 请求的权限信息
type RequestInfo struct {
	AppID          string
	MasterKey      string
	ClientKey      string
	JavaScriptKey  string
	DotNetKey      string
	RestAPIKey     string
	SessionToken   string
	InstallationID string
	ClientVersion  string
	ClientSDK      map[string]string
	Headers        map[string]string
}
