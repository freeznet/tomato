package files

import (
	"github.com/freeznet/tomato/cloud"
	"github.com/freeznet/tomato/types"
)

func getRequest(triggerType string, orifilename string, data []byte, contentType string, filename string, location string) cloud.TriggerRequest {
	request := cloud.TriggerRequest{
		TriggerName: triggerType,
		Object:      nil,
		Master:      false,
	}

	if filename != "" {
		request.Filename = filename
	}
	if orifilename != "" {
		request.OriFilename = orifilename
	}
	if data != nil {
		request.Data = data
	}
	if contentType != "" {
		request.ContentType = contentType
	}
	if location != "" {
		request.Location = location
	}

	return request
}

func getResponse(request cloud.TriggerRequest) *cloud.TriggerResponse {
	response := &cloud.TriggerResponse{
		Request: request,
	}
	return response
}


func maybeRunTrigger(triggerType string, className string, orifilename string, data []byte, contentType string) (types.M, error) {
	if data == nil {
		return types.M{}, nil
	}

	trigger := cloud.GetTrigger(triggerType, className)
	if trigger == nil {
		return types.M{}, nil
	}
	request := getRequest(triggerType, orifilename, data, contentType, "", "")
	response := getResponse(request)
	trigger(request, response)
	return response.Response, response.Err
}

func maybeRunAfterTrigger(triggerType string, className, orifilename, filename, location string) (types.M, error) {
	if filename == "" {
		return types.M{}, nil
	}

	trigger := cloud.GetTrigger(triggerType, className)
	if trigger == nil {
		return types.M{}, nil
	}
	request := getRequest(triggerType, orifilename, nil, "", filename, location)
	response := getResponse(request)
	trigger(request, response)
	return response.Response, response.Err
}
