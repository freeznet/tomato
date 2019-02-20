package analytics

import "github.com/freeznet/tomato/types"

type nullAnalyticsAdapter struct {
}

func (a *nullAnalyticsAdapter) appOpened(body types.M) (types.M, error) {
	return types.M{}, nil
}

func (a *nullAnalyticsAdapter) trackEvent(eventName string, body types.M) (types.M, error) {
	return types.M{}, nil
}
