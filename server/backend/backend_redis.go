package backend

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"log"
	"math"
	"math/rand"
)

const redisType = TypeBackend("redis")
const timestampBucketSize = 86400 * 1000 // 1 day in milliseconds

var redisNoConnForNamespaceErr = errors.New("no connection for namespace")

type RedisBackend struct {
	opts *RedisOpts

	connections []*redis.Client
}

func (instance *RedisBackend) Type() TypeBackend {
	return redisType
}

func (instance *RedisBackend) getKey(ctx Context, timestamp uint64) string {
	timestampBucket := timestamp - (timestamp % timestampBucketSize)
	return fmt.Sprintf("%d-%d-%d", ctx.Namespace, ctx.Series, timestampBucket)
}

func (instance *RedisBackend) Write(context ContextWrite, timestamps []uint64, values []float64) error {
	conn := instance.getConnection(Namespace(context.Namespace))
	if conn == nil {
		return redisNoConnForNamespaceErr
	}
	keyValues := make(map[string][]redis.Z)
	for idx, timestamp := range timestamps {
		// determine key
		key := instance.getKey(context.Context, timestamp)

		// init key
		if keyValues[key] == nil {
			keyValues[key] = make([]redis.Z, 0)
		}

		// value
		value := values[idx]

		// to make sure we don't ever have colliding timestamps
		tsPadded := float64(timestamp) + (rand.Float64() * maxPaddingSize)

		// member
		member := redis.Z{
			Score:  value,                       // Sorted sets are sorted by their score in an ascending way. The same element only exists a single time, no repeated elements are permitted.
			Member: fmt.Sprintf("%v", tsPadded), // must be string
		}

		// add
		keyValues[key] = append(keyValues[key], member)
	}

	// execute
	for key, members := range keyValues {
		// execute
		res := conn.ZAdd(key, members...)
		if res.Err() != nil {
			return res.Err()
		}
		log.Printf("%+v %v", res, members)
		if res.Val() != int64(len(members)) {
			return errors.New("failed write count")
		}
	}

	return nil
}

func (instance *RedisBackend) getKeysInRange(ctx ContextRead) []string {
	keys := make([]string, 0)
	for ts := ctx.From; ts < ctx.To; ts += timestampBucketSize {
		key := instance.getKey(ctx.Context, ts)
		keys = append(keys, key)
	}
	return keys
}

func (instance *RedisBackend) Read(context ContextRead) (res ReadResult) {
	keys := instance.getKeysInRange(context)
	conn := instance.getConnection(Namespace(context.Namespace))
	for _, key := range keys {
		read := conn.ZRangeByScoreWithScores(key, redis.ZRangeBy{
			Min: fmt.Sprintf("%v", context.From),
			Max: fmt.Sprintf("%v", context.To),
		})
		if read.Err() != nil {
			res.Error = read.Err()
			return
		}
		values := read.Val()
		for _, value := range values {
			log.Printf("%+v", value)
		}
		// @todo implement into ReadResult
	}
	res.Error = errors.New("not implemented")

	return
}

func (instance *RedisBackend) CreateOrUpdateSeries(create *CreateSeries) (result *CreateSeriesResult) {
	return nil
}

func (instance *RedisBackend) SearchSeries(search *SearchSeries) (result *SearchSeriesResult) {
	return nil
}

func (instance *RedisBackend) DeleteSeries(ops *DeleteSeries) (result *DeleteSeriesResult) {
	return nil
}

func (instance *RedisBackend) getConnection(namespace Namespace) *redis.Client {
	return instance.connections[namespace]
}

func (instance *RedisBackend) Init() error {
	var minNamespace = Namespace(math.MaxInt32)
	var maxNamespace = Namespace(math.MaxInt32 * -1)
	clients := make(map[Namespace]*redis.Client)
	for namespace, details := range instance.opts.ConnectionDetails {
		client := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", details.Addr, details.Port),
			Password: details.Password, // no password set
			DB:       details.Database, // use default DB
		})

		// ping pong
		_, err := client.Ping().Result()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to init namespace %d", namespace))
		}

		// assign to map
		clients[namespace] = client

		// track min max
		if namespace < minNamespace {
			minNamespace = namespace
		}
		if namespace > maxNamespace {
			maxNamespace = namespace
		}
	}

	// convert to array to not have concurrent map access issues
	instance.connections = make([]*redis.Client, maxNamespace-minNamespace+1)
	for k, v := range clients {
		instance.connections[k] = v
	}

	return nil
}

func NewRedisBackend(opts *RedisOpts) *RedisBackend {
	return &RedisBackend{
		opts: opts,
	}
}

type RedisOpts struct {
	// connection per namespace supported, use RedisDefaultConnectionNamespace for default
	ConnectionDetails map[Namespace]RedisConnectionDetails
}

type RedisConnectionDetails struct {
	Addr     string
	Port     int
	Password string
	Database int
}

const RedisDefaultConnectionNamespace = Namespace(0)
