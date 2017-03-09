package analytics

import "github.com/freeznet/tomato/types"

type analyticsAdapter struct {
}

func (a *analyticsAdapter) appOpened(body types.M) (types.M, error) {
	return types.M{}, nil
}

func (a *analyticsAdapter) trackEvent(eventName string, body types.M) (types.M, error) {
	return types.M{}, nil
}
