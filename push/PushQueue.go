package push

import (
	"encoding/json"
	"errors"

<<<<<<< HEAD
	"github.com/freeznet/tomato/livequery/pubsub"
	"github.com/freeznet/tomato/rest"
	"github.com/freeznet/tomato/types"
=======
	"github.com/lfq7413/tomato/livequery/pubsub"
	"github.com/lfq7413/tomato/rest"
	"github.com/lfq7413/tomato/types"
	"github.com/lfq7413/tomato/utils"
>>>>>>> 13ed8fdfa0dda59315337ef00487581dc332bce1
)

const (
	pushChannel      = "parse-server-push"
	defaultBatchSize = 100
)

// pushQueue ...
type pushQueue struct {
	parsePublisher pubsub.Publisher
	channel        string
	batchSize      int
}

func newPushQueue(channel string, batchSize int) *pushQueue {
	if channel == "" {
		channel = pushChannel
	}
	if batchSize == 0 {
		batchSize = defaultBatchSize
	}
	return &pushQueue{
		parsePublisher: CreatePublisher(),
		channel:        channel,
		batchSize:      batchSize,
	}
}

func (q *pushQueue) enqueue(body, where types.M, auth *rest.Auth, status *pushStatus) error {
	limit := q.batchSize
	order := ""
	if isPushIncrementing(body) {
		order = "badge,createdAt"
	} else {
		order = "createdAt"
	}

<<<<<<< HEAD
=======
	where = utils.CopyMapM(where)
	if _, ok := where["deviceToken"]; !ok {
		where["deviceToken"] = types.M{"$exists": true}
	}

>>>>>>> 13ed8fdfa0dda59315337ef00487581dc332bce1
	options := types.M{
		"limit": 0,
		"count": true,
	}
	result, err := rest.Find(auth, "_Installation", where, options, nil)
	if err != nil {
		return err
	}

	count := 0
	if c, ok := result["count"].(int); ok {
		count = c
	}

	if count == 0 {
		return errors.New("PushController: no results in query")
	}
	status.setRunning(count)

	for skip := 0; skip < count; skip += limit {
		query := types.M{
			"where": where,
			"limit": limit,
			"skip":  skip,
			"order": order,
		}
		pushWorkItem := types.M{
			"body":       body,
			"query":      query,
			"pushStatus": types.M{"objectId": status.objectID},
		}
		b, err := json.Marshal(pushWorkItem)
		if err != nil {
			return err
		}
		q.parsePublisher.Publish(q.channel, string(b))
	}

	return nil
}
