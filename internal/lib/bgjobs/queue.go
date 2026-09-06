package bgjobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type queue struct {
	name string
	rdb  *redis.Client
}

func newQueue(name string, rdb *redis.Client) queue {
	return queue{name, rdb}
}

func (q queue) enqueue(ctx context.Context, t []byte) error {
	if _, err := q.rdb.LPush(ctx, q.name, t).Result(); err != nil {
		return err
	}

	return nil
}

func (q queue) popAndMoveTo(ctx context.Context, dst queue) (string, error) {
	v, err := q.rdb.BLMove(ctx, q.name, dst.name, "RIGHT", "LEFT", time.Duration(0)).Result()
	if err != nil {
		return "", err
	}

	return v, nil
}

func (q queue) remove(ctx context.Context, v any) error {
	if _, err := q.rdb.LRem(ctx, q.name, 0, v).Result(); err != nil {
		return err
	}

	return nil
}

/* Replaced by popAndMoveTo

func (q queue) dequeue(ctx context.Context) (string, error) {
raw, err := q.rdb.BRPop(ctx, time.Duration(0), q.name).Result()
if err != nil {
return "", err
}

return raw[0], nil
}
*/

func newPrimaryQueue(rdb *redis.Client) queue {
	return newQueue(Primary, rdb)
}

func newTempQueue(rdb *redis.Client) queue {
	return newQueue(Temp, rdb)
}

func newDLQ(rdb *redis.Client) queue {
	return newQueue(DLQ, rdb)
}

type deadLetter struct {
	Task
	Err       []string  `json:"err"`
	CreatedAt time.Time `json:"createdAt"`
}

func sendToDLQ(ctx context.Context, dlq queue, tempQ queue, raw string, t Task, errors ...string) {
	dl := deadLetter{
		Task:      t,
		Err:       errors,
		CreatedAt: time.Now(),
	}

	// marshall struct into json
	jdl, err := json.Marshal(dl)
	if err != nil {
		fmt.Printf("[Dead Letter]: Failed to marshal dl with errors [%v] into json: %s", errors, err.Error())
	}

	// send to DLQ
	if err := dlq.enqueue(ctx, jdl); err != nil {
		fmt.Printf("[Dead Letter]: Failed to enqueue to DLQ with errors [%v]: %s", errors, err.Error())
	}

	if err := tempQ.remove(ctx, raw); err != nil {
		fmt.Printf("Failed to remove task [%s] from temp queue: %s\n", t.Id, err)
	}
}
