package backend

import (
	"encoding/json"
	"fmt"
	"github.com/RobinUS2/tsxdb/rpc/types"
	"github.com/alicebob/miniredis/v2"
	lock "github.com/bsm/redislock"
	"github.com/go-redis/redis/v7"
	"github.com/pkg/errors"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const RedisType = TypeBackend("redis")
const timestampBucketSize = 86400 * 1000 // 1 day in milliseconds
const defaultExpiryTime = time.Minute    // default time one can lock redis

var redisNoConnForNamespaceErr = errors.New("no connection for namespace")

type RedisBackend struct {
	opts *RedisOpts

	connections []redis.UniversalClient

	AbstractBackend
}

func (instance *RedisBackend) Type() TypeBackend {
	return RedisType
}

func (instance *RedisBackend) getDataKey(ctx Context, timestamp uint64) string {
	timestampBucket := timestamp - (timestamp % timestampBucketSize)
	return fmt.Sprintf("data_%d-%d-%d", ctx.Namespace, ctx.Series, timestampBucket)
}

func (instance *RedisBackend) Write(context ContextWrite, timestamps []uint64, values []float64) error {
	conn := instance.GetConnection(Namespace(context.Namespace))
	if conn == nil {
		return redisNoConnForNamespaceErr
	}
	keyValues := make(map[string][]*redis.Z)
	for idx, timestamp := range timestamps {
		// determine key
		key := instance.getDataKey(context.Context, timestamp)

		// init key
		if keyValues[key] == nil {
			keyValues[key] = make([]*redis.Z, 0)
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
		keyValues[key] = append(keyValues[key], &member)
	}

	// meta
	meta, err := instance.getMetadata(Namespace(context.Namespace), context.Series, false)
	if err != nil || meta.Id < 1 {
		return err
	}

	expireTime := time.Unix(int64(meta.TtlExpire), 0)
	// execute
	for key, members := range keyValues {
		zRangeRes := conn.ZRange(key, 0, -1)
		isNew := len(zRangeRes.Val()) == 0

		// execute
		// @todo use pipelined redis transaction for reduced network round-trip and CPU usage
		res := conn.ZAdd(key, members...)
		if res.Err() != nil {
			return res.Err()
		}

		if meta.TtlExpire > nowSeconds() {
			if isNew {
				conn.ExpireAt(key, expireTime)
			}
			ttlRes := conn.TTL(key)
			if ttlRes.Val() == 0 {
				// no ttl set, set again
				conn.ExpireAt(key, expireTime)
			}
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
		key := instance.getDataKey(ctx.Context, ts)
		keys = append(keys, key)
	}
	return keys
}

func (instance *RedisBackend) Read(context ContextRead) (res ReadResult) {
	conn := instance.GetConnection(Namespace(context.Namespace))

	// meta
	meta, err := instance.getMetadata(Namespace(context.Namespace), context.Series, false)
	if err != nil || meta.Id < 1 {
		res.Error = errors.Wrap(err, "missing metadata")
		return
	}

	// read
	keys := instance.getKeysInRange(context)
	var resultMap map[uint64]float64
	for _, key := range keys {
		read := conn.ZRangeByScoreWithScores(key, &redis.ZRangeBy{
			Min: FloatToString(float64(context.From)),
			Max: FloatToString(float64(context.To)),
		})
		if filterNilErr(read.Err()) != nil {
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
			if resultMap == nil {
				resultMap = make(map[uint64]float64)
			}
			resultMap[uint64(value.Score)] = floatValue
		}
	}
	// no data?
	if resultMap == nil {
		res.Error = ErrNoDataFound
		return
	}
	res.Results = resultMap

	return
}

func (instance *RedisBackend) getSeriesByNameKey(namespace Namespace, name string) string {
	return fmt.Sprintf("series_%d_%s", namespace, name) // always prefix with namespace
}

func (instance *RedisBackend) getSeriesMetaKey(namespace Namespace, id uint64) string {
	return fmt.Sprintf("series_%d_%d_meta", namespace, id) // always prefix with namespace
}

func (instance *RedisBackend) createOrUpdateSeries(identifier types.SeriesCreateIdentifier, series types.SeriesCreateMetadata) (result types.SeriesMetadataResponse, err error) {
	// get right client
	conn := instance.GetConnection(Namespace(series.Namespace))

	// existing
	seriesKey := instance.getSeriesByNameKey(Namespace(series.Namespace), series.Name)
	res := conn.Get(seriesKey)
	if filterNilErr(res.Err()) != nil {
		err = res.Err()
		return
	}
	if res.Val() == "" {
		// not existing
		lockKey := "lock_" + seriesKey
		createLock, err := lock.Obtain(conn, lockKey, defaultExpiryTime, nil)
		if err != nil || createLock == nil {
			// fail to obtain lock
			return result, errors.New(fmt.Sprintf("failed to obtain metadata lock %v", err))
		}
		// locked

		// unlock
		defer func() {
			// unlock
			if lockErr := createLock.Release(); err != nil {
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
		id, err := idStrToIdUint64(res.Val())
		if err != nil {
			return result, err
		}
		result.New = false
		result.Id = id
	}

	// only store metadata during creation for now
	if result.New {
		// metadata persist (just overwrite for now)
		{
			metaKey := instance.getSeriesMetaKey(Namespace(series.Namespace), result.Id)

			// convert to server version
			var ttlExpire uint64
			if series.Ttl > 0 {
				ttlExpire = nowSeconds() + uint64(series.Ttl)
			}
			data := SeriesMetadata{
				Name:      series.Name,
				Namespace: Namespace(series.Namespace),
				Tags:      series.Tags,
				Id:        Series(result.Id),
				TtlExpire: ttlExpire,
			}
			j, err := json.Marshal(data)
			if err != nil {
				panic(err)
			}
			if res := conn.Set(metaKey, string(j), 0); res.Err() != nil {
				return result, res.Err()
			}
		}

		// persist tags
		if series.Tags != nil {
			for _, tag := range series.Tags {
				tagKey := instance.getTagKey(Namespace(series.Namespace), tag)
				if res := conn.SAdd(tagKey, result.Id); res.Err() != nil {
					return result, res.Err()
				}
			}
		}
	}

	return
}

func (instance *RedisBackend) getTagKey(namespace Namespace, tag string) string {
	return fmt.Sprintf("tag_%d_%s", namespace, tag) // always prefix with namespace
}

func idStrToIdUint64(in string) (uint64, error) {
	id, err := strconv.ParseUint(in, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
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
	result = &SearchSeriesResult{
		Series: nil, // lazy init
	}
	if search.And != nil {
		result.Error = errors.New("no AND support yet")
		return
	}
	if search.Or != nil {
		result.Error = errors.New("no OR support yet")
		return
	}
	if search.Comparator != SearchSeriesComparatorEquals {
		result.Error = errors.New("only EQUALS support")
		return
	}
	if search.Tag != "" {
		result.Error = errors.New("not tag support yet")
		return
	}

	// by name
	if search.Name != "" {
		conn := instance.GetConnection(Namespace(search.Namespace))
		seriesKey := instance.getSeriesByNameKey(Namespace(search.Namespace), search.Name)
		res := conn.Get(seriesKey)
		if filterNilErr(res.Err()) != nil {
			result.Error = res.Err()
			return
		}
		if res.Err() == redis.Nil {
			// not found
			return
		}
		var id uint64
		var err error
		if id, err = idStrToIdUint64(res.Val()); err != nil {
			result.Error = err
			return
		}
		result.Series = []types.SeriesIdentifier{
			{
				Namespace: search.Namespace,
				Id:        id,
			},
		}
		return
	}

	return nil
}

func (instance *RedisBackend) getMetadata(namespace Namespace, id uint64, ignoreExpiry bool) (result SeriesMetadata, err error) {
	conn := instance.GetConnection(namespace)
	metaKey := instance.getSeriesMetaKey(namespace, id)

	res := conn.Get(metaKey)
	if res.Err() != nil {
		return result, errors.Wrapf(res.Err(), "series %d", id)
	}

	var data SeriesMetadata
	if err := json.Unmarshal([]byte(res.Val()), &data); err != nil {
		return result, err
	}

	// ttl of series
	if !ignoreExpiry && data.TtlExpire > 0 {
		nowSeconds := nowSeconds()
		if data.TtlExpire < nowSeconds {
			// expired, remove it
			res := instance.ReverseApi().DeleteSeries(&DeleteSeries{
				Series: []types.SeriesIdentifier{
					{
						Namespace: namespace.Int(),
						Id:        id,
					},
				},
			})
			if res.Error != nil {
				// @todo deal with in other way
				panic(res.Error)
			}
			err = errors.Wrapf(ErrNoDataFound, "series %d expired", id)
			return
		}
	}
	result = data

	return
}

func (instance *RedisBackend) DeleteSeries(ops *DeleteSeries) (result *DeleteSeriesResult) {
	result = &DeleteSeriesResult{}
	for _, op := range ops.Series {
		conn := instance.GetConnection(Namespace(op.Namespace))

		// meta
		meta, err := instance.getMetadata(Namespace(op.Namespace), op.Id, true)
		if err != nil {
			result.Error = err
			return
		}

		// id
		idStr := fmt.Sprintf("%d", meta.Id)

		// key
		if res := conn.Del(instance.getSeriesByNameKey(Namespace(op.Namespace), meta.Name)); res.Err() != nil {
			result.Error = res.Err()
			return
		}

		// tag memberships (based on meta)
		for _, tag := range meta.Tags {
			tagKey := instance.getTagKey(Namespace(op.Namespace), tag)
			res := conn.SRem(tagKey, idStr)
			if res.Err() != nil {
				result.Error = res.Err()
				return
			}
		}

		// meta key
		if res := conn.Del(instance.getSeriesMetaKey(Namespace(op.Namespace), uint64(meta.Id))); res.Err() != nil {
			result.Error = res.Err()
			return
		}

		// data keys are deleted using ttl expire
	}
	return
}

func (instance *RedisBackend) GetConnection(namespace Namespace) redis.UniversalClient {
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
	clients := make(map[Namespace]redis.UniversalClient)
	for namespace, details := range instance.opts.ConnectionDetails {
		var client redis.UniversalClient
		switch details.Type {
		case RedisCluster:
			client = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs: []string{
					fmt.Sprintf("%s:%d", details.Addr, details.Port),
				},
				Password: details.Password,
			})
		case RedisServer:
			client = redis.NewClient(&redis.Options{
				Addr:     fmt.Sprintf("%s:%d", details.Addr, details.Port),
				Password: details.Password, // no password set
				DB:       details.Database, // use default DB
			})
		case RedisMemory:
			miniRedis, err := miniredis.Run()
			if err != nil {
				panic(err)
			}
			client = redis.NewClient(&redis.Options{
				Addr: miniRedis.Addr(),
				DB:   details.Database,
			})
			go func() {
				// mini redis does not forward time, and thus never expires key.
				// simple time forwarding
				duration := time.Millisecond * 100
				ticker := time.NewTicker(duration)
				for range ticker.C {
					miniRedis.FastForward(duration)
				}
			}()
		default:
			panic(fmt.Sprintf("type %s not supported", details.Type))
		}
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
	instance.connections = make([]redis.UniversalClient, maxNamespace-minNamespace+1)
	for k, v := range clients {
		instance.connections[k] = v
	}

	return nil
}

func (instance *RedisBackend) Clear() error {
	return errors.New("backend redis does not support clearing, this is supposed to be only used for testing")
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

type RedisServerType string

const RedisCluster = "cluster"
const RedisServer = "server"
const RedisMemory = "memory" // miniredis client

type RedisConnectionDetails struct {
	Type     RedisServerType //sentinel cluster server
	Addr     string
	Port     int
	Password string
	Database int
}

const RedisDefaultConnectionNamespace = Namespace(0)
