package backend

import (
	"encoding/json"
	"fmt"
	"github.com/RobinUS2/tsxdb/rpc/types"
	"github.com/alicebob/miniredis/v2"
	lock "github.com/bsm/redislock"
	"github.com/go-redis/redis/v7"
	"github.com/jinzhu/now"
	"github.com/karlseguin/ccache/v2"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"log"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const expireWrittenCacheDuration = 60 * time.Minute

const RedisType = TypeBackend("redis")
const timestampBucketSize = 86400 * 1000 // 1 day in milliseconds
var timestampBucketSizeStrLength = len(fmt.Sprintf("%d", timestampBucketSize))

const defaultExpiryTime = time.Minute // default time one can lock redis
const MetadataLocalCacheDuration = 1000 * time.Millisecond

var redisNoConnForNamespaceErr = errors.New("no connection for namespace")

type RedisBackend struct {
	opts *RedisOpts

	connections []redis.UniversalClient

	AbstractBackend

	expireWrittenCache *cache.Cache // used to deduplicate expire writes (not always needed)

	pipelines    map[RequestId]redis.Pipeliner
	pipelinesMux sync.RWMutex

	metadataCache *ccache.Cache
}

func (instance *RedisBackend) Type() TypeBackend {
	return RedisType
}

func (instance *RedisBackend) getDataKey(ctx Context, timestamp uint64) (string, uint64) {
	timestampBucket := timestamp - (timestamp % timestampBucketSize)
	return fmt.Sprintf("data_%d-%d-%d", ctx.Namespace, ctx.Series, timestampBucket), timestampBucket
}

func (instance *RedisBackend) getKeyScoreAndMember(context ContextWrite, timestamp uint64, value float64) (key string, score float64, member string) {
	// this function basically creates a unique "member" for the redis set so that it's not overwritten, it uses the timestamp prefix with a random value
	// the timestamp prefix will be truncated to last digits (most significant, seconds etc versus the month bit).
	var tsBucket uint64
	key, tsBucket = instance.getDataKey(context.Context, timestamp)
	score = float64(timestamp) + (rand.Float64() * maxPaddingSize)
	tsBucketPrefix := fmt.Sprintf("%d", tsBucket)
	// how many chars do we need of the timestamp suffix / last digits?
	finalChars := len(tsBucketPrefix) - timestampBucketSizeStrLength
	if finalChars >= 1 {
		// very old timestamps (e.g. timestamp 12345) will have not enough left to remove the timestampBucketSizeStrLength (which is generally 8), so will only do if makes sense
		tsBucketPrefix = tsBucketPrefix[0:finalChars]
	}
	tsPaddedStr := strings.TrimPrefix(fmt.Sprintf("%f", score), tsBucketPrefix)
	member = FloatToString(value) + fmt.Sprintf(":%s", tsPaddedStr) // must be string and unique, that's why we take the timestamp with random value
	return key, score, member
}

func (instance *RedisBackend) FlushPendingWrites(requestId RequestId) error {
	if IsEmptyRequestId(requestId) {
		return fmt.Errorf("empty request id %s", requestId)
	}
	// flush pipeline
	instance.pipelinesMux.RLock()
	v, found := instance.pipelines[requestId]
	instance.pipelinesMux.RUnlock()
	if !found {
		return fmt.Errorf("missing pipeliner for request %s", requestId)
	}

	// remove once done
	defer func() {
		instance.pipelinesMux.Lock()
		delete(instance.pipelines, requestId)
		instance.pipelinesMux.Unlock()
	}()

	// execute buffer
	if _, err := v.Exec(); err != nil {
		return errors.Wrap(err, "redis flush pending writes failed")
	}

	return nil
}

func (instance *RedisBackend) getPipeline(context ContextWrite) (redis.Pipeliner, error) {
	// existing?
	instance.pipelinesMux.RLock()
	v, found := instance.pipelines[context.RequestId]
	instance.pipelinesMux.RUnlock()
	if found {
		return v, nil
	}

	// redis connection
	conn := instance.GetConnection(Namespace(context.Namespace))
	if conn == nil {
		return nil, redisNoConnForNamespaceErr
	}

	// pipeline, reduces (tcp) round-trip overhead, system calls, etc
	pipeline := conn.Pipeline()

	// store for use during this request
	instance.pipelinesMux.Lock()
	instance.pipelines[context.RequestId] = pipeline
	instance.pipelinesMux.Unlock()

	return pipeline, nil
}

func (instance *RedisBackend) Write(context ContextWrite, timestamps []uint64, values []float64) error {
	if IsEmptyRequestId(context.RequestId) {
		return fmt.Errorf("empty request id %s", context.RequestId)
	}

	keyValues := make(map[string][]*redis.Z)
	for idx, timestamp := range timestamps {

		// value
		value := values[idx]

		// determine key
		key, score, setMember := instance.getKeyScoreAndMember(context, timestamp, value)

		// init key
		if keyValues[key] == nil {
			keyValues[key] = make([]*redis.Z, 0)
		}

		// member
		member := redis.Z{
			Score:  score,     // Sorted sets are sorted by their score in an ascending way. The same element only exists a single time, no repeated elements are permitted. The score is basically the timestamp of the write request.
			Member: setMember, // must be string and unique
		}

		// add
		keyValues[key] = append(keyValues[key], &member)
	}

	// meta
	meta, err := instance.getMetadata(Namespace(context.Namespace), context.Series, false)
	if err != nil {
		if strings.Contains(err.Error(), types.RpcErrorSeriesExpired.String()) {
			// series expired, not a real problem
			return nil
		}
	}

	// get redis pipeline
	pipeline, err := instance.getPipeline(context)
	if err != nil {
		return err
	}

	// when to expire?
	var expireTime time.Time
	if meta.TtlExpire == 0 {
		expireTime = now.EndOfDay()
	} else {
		expireTime = time.Unix(int64(meta.TtlExpire), 0)
	}

	// add commands to redis pipeline
	for key, members := range keyValues {
		if expireTime.Unix() > 0 && time.Since(expireTime) > 0 {
			// key already expired, skip
			continue
		}

		// execute
		res := pipeline.ZAdd(key, members...)
		if res.Err() != nil {
			return res.Err()
		}

		if expireTime.Unix() > 0 {
			// deduplicate expire at, if we've recently done
			expireWrittenCacheKey := fmt.Sprintf("%d", context.Series)
			if _, found := instance.expireWrittenCache.Get(expireWrittenCacheKey); !found {
				pipeline.ExpireAt(key, expireTime)
				instance.expireWrittenCache.Set(expireWrittenCacheKey, true, expireWrittenCacheDuration)
			}
		}
	}

	// we do NOT execute the transaction here, we do that during final flush

	return nil
}

var replaceLeadingZeroDot = regexp.MustCompile(`^0\.`)

func FloatToString(val float64) string {
	// to convert a float number to a string, trim trailing zeros to save space
	return replaceLeadingZeroDot.ReplaceAllString(strings.TrimRight(strconv.FormatFloat(val, 'f', 6, 64), "0"), ".")
}

func (instance *RedisBackend) getKeysInRange(ctx ContextRead) ([]string, []uint64) {
	keys := make([]string, 0)
	tsBuckets := make([]uint64, 0)
	for ts := ctx.From; ts < ctx.To; ts += timestampBucketSize {
		key, tsBucket := instance.getDataKey(ctx.Context, ts)
		keys = append(keys, key)
		tsBuckets = append(tsBuckets, tsBucket)
	}
	return keys, tsBuckets
}

func (instance *RedisBackend) Read(context ContextRead) (res ReadResult) {
	// meta
	meta, err := instance.getMetadata(Namespace(context.Namespace), context.Series, false)
	if err != nil {
		res.Error = err
		return
	} else if meta.Id < 1 {
		res.Error = errors.New("missing id")
		return
	}

	conn := instance.GetConnection(Namespace(context.Namespace))

	// read
	keys, _ := instance.getKeysInRange(context)
	var resultMap map[uint64]float64
	for _, key := range keys {
		read := conn.ZRangeByScoreWithScores(key, &redis.ZRangeBy{
			Min: FloatToString(float64(context.From - 1)), // pad 1 ms to make sure the scores are available due to float rounding
			Max: FloatToString(float64(context.To + 1)),
		})
		if filterNilErr(read.Err()) != nil {
			res.Error = read.Err()
			return
		}
		values := read.Val()
		for _, value := range values {
			member := value.Member.(string)
			memberSplit := strings.Split(member, ":")
			if len(memberSplit) != 2 {
				panic("should always be 2 parts")
			}
			var floatValue float64
			var err error
			if memberSplit[0] == "." {
				floatValue = 0.0
			} else {
				floatValue, err = strconv.ParseFloat(memberSplit[0], 64)
				if err != nil {
					res.Error = fmt.Errorf("parse float err %s,%v: %s", key, value, err)
					return
				}
			}

			if resultMap == nil {
				resultMap = make(map[uint64]float64)
			}
			resultMap[uint64(value.Score)] = floatValue
		}
	}
	// no data?
	if resultMap == nil {
		res.Error = types.RpcErrorNoDataFound.Error()
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
			instance.metadataCache.DeletePrefix(metaKey) // wipe metadata cache
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
	cacheKey := instance.getSeriesMetaKey(namespace, id) + strconv.FormatBool(ignoreExpiry)
	val, err := instance.metadataCache.Fetch(cacheKey, MetadataLocalCacheDuration, func() (interface{}, error) {
		return instance.getMetadataFromStorage(namespace, id, ignoreExpiry)
	})
	if err != nil {
		return
	}
	result = val.Value().(SeriesMetadata)
	return
}

func (instance *RedisBackend) getMetadataFromStorage(namespace Namespace, id uint64, ignoreExpiry bool) (result SeriesMetadata, err error) {
	metaKey := instance.getSeriesMetaKey(namespace, id)
	conn := instance.GetConnection(namespace)
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
			err = errors.Wrapf(types.RpcErrorNoDataFound.Error(), fmt.Sprintf("%s (id=%d)", types.RpcErrorSeriesExpired.String(), id))
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
				warnDuration := time.Duration(float64(duration) * 0.8) // 80% of time spent on locking
				var lastCompleted time.Time
				ticker := time.NewTicker(duration)
				for range ticker.C {
					if time.Since(lastCompleted) < duration {
						// prevent continuously locking mini redis causing high latency
						continue
					}
					startTime := time.Now()
					miniRedis.FastForward(duration)
					took := time.Since(startTime)
					lastCompleted = time.Now()
					if took > warnDuration {
						log.Printf("WARN mini redis fast forward took %s (this simulates expiry of testing redis)", took)
					}
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
		opts:               opts,
		expireWrittenCache: cache.New(60*time.Minute, 1*time.Minute),
		pipelines:          make(map[RequestId]redis.Pipeliner),
		metadataCache:      ccache.New(ccache.Configure().MaxSize(1000 * 1000)),
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
