package backend

import (
	"../../rpc/types"
	"fmt"
	"github.com/bsm/redis-lock"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"math"
	"math/rand"
	"strconv"
	"strings"
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
			Score:  tsPadded,                                            // Sorted sets are sorted by their score in an ascending way. The same element only exists a single time, no repeated elements are permitted.
			Member: FloatToString(value) + fmt.Sprintf(":%f", tsPadded), // must be string and unique
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
		if res.Val() != int64(len(members)) {
			return errors.New("failed write count")
		}
	}

	return nil
}

func FloatToString(val float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(val, 'f', 6, 64)
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
	resultMap := make(map[uint64]float64)
	for _, key := range keys {
		read := conn.ZRangeByScoreWithScores(key, redis.ZRangeBy{
			Min: FloatToString(float64(context.From)),
			Max: FloatToString(float64(context.To)),
		})
		if read.Err() != nil {
			res.Error = read.Err()
			return
		}
		values := read.Val()
		for _, value := range values {
			member := value.Member.(string)
			memberSplit := strings.Split(member, ":")
			floatValue, err := strconv.ParseFloat(memberSplit[0], 64)
			if err != nil {
				res.Error = err
				return
			}
			resultMap[uint64(value.Score)] = floatValue
		}
	}
	res.Results = resultMap

	return
}

func (instance *RedisBackend) createOrUpdateSeries(identifier types.SeriesCreateIdentifier, series types.SeriesMetadata) (result types.SeriesMetadataResponse, err error) {
	// get right client
	conn := instance.getConnection(Namespace(series.Namespace))

	// existing
	seriesKey := fmt.Sprintf("series_%d_%s", series.Namespace, series.Name) // always prefix with namespace
	res := conn.Get(seriesKey)
	if filterNilErr(res.Err()) != nil {
		err = res.Err()
		return
	}
	if res.Val() == "" {
		// not existing
		lockKey := "lock_" + seriesKey
		createLock, err := lock.Obtain(conn, lockKey, nil)
		if err != nil || createLock == nil {
			// fail to obtain lock
			return result, errors.New(fmt.Sprintf("failed to obtain metadata lock %v", err))
		}
		// locked

		// unlock
		defer func() {
			// unlock
			if lockErr := createLock.Unlock(); err != nil {
				err = lockErr
				return
			}
		}()

		// existing (check again in lock)
		res := conn.Get(seriesKey)
		if res.Err() == redis.Nil {
			// not existing, create

			// ID, increment in namespace
			idKey := fmt.Sprintf("id_%d", series.Namespace)
			idRes := conn.Incr(idKey)
			if filterNilErr(idRes.Err()) != nil {
				return result, errors.Wrap(idRes.Err(), "increment failed")
			}
			newId := uint64(idRes.Val())

			// write to redis
			writeRes := conn.Set(seriesKey, fmt.Sprintf("%d", newId), 0)
			if writeRes.Err() != nil {
				return result, writeRes.Err()
			}

			// result vars
			result.New = true
			result.Id = newId
		}

	} else {
		// existing
		id, err := strconv.ParseUint(res.Val(), 10, 64)
		if err != nil {
			return result, err
		}
		result.New = false
		result.Id = id
	}
	return
}

func (instance *RedisBackend) CreateOrUpdateSeries(create *CreateSeries) (result *CreateSeriesResult) {
	result = &CreateSeriesResult{}
	for identifier, series := range create.Series {
		subRes, err := instance.createOrUpdateSeries(identifier, series)
		if err != nil {
			result.Error = err
			return
		}
		if result.Results == nil {
			result.Results = make(map[types.SeriesCreateIdentifier]types.SeriesMetadataResponse)
		}
		result.Results[identifier] = subRes
	}
	return
}

func filterNilErr(err error) error {
	if err == redis.Nil {
		return nil
	}
	return err
}

func (instance *RedisBackend) SearchSeries(search *SearchSeries) (result *SearchSeriesResult) {
	return nil
}

func (instance *RedisBackend) DeleteSeries(ops *DeleteSeries) (result *DeleteSeriesResult) {
	return nil
}

func (instance *RedisBackend) getConnection(namespace Namespace) *redis.Client {
	val := instance.connections[namespace]
	if val != nil {
		return val
	}
	// fallback to default connection
	return instance.connections[RedisDefaultConnectionNamespace]
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
