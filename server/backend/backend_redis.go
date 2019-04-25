package backend

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"math"
)

type RedisBackend struct {
	opts *RedisOpts

	connections []*redis.Client
}

func (instance *RedisBackend) Type() TypeBackend {
	return TypeBackend("redis")
}

func (instance *RedisBackend) Write(context ContextWrite, timestamps []uint64, values []float64) error {
	return errors.New("not implemented")
}

func (instance *RedisBackend) Read(context ContextRead) (res ReadResult) {
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

var RedisDefaultConnectionNamespace = Namespace(0)
