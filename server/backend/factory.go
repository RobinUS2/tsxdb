package backend

import (
	"fmt"
)

func InstanceFactory(typeStr string, opts map[string]interface{}) IAbstractBackend {
	switch typeStr {
	case MemoryType.String():
		return NewMemoryBackend()
	case RedisType.String():
		return NewRedisBackend(ExtractRedisOpts(opts))
	default:
		panic(fmt.Sprintf("backend %s not supported", typeStr))
	}
}

func ExtractRedisOpts(opts map[string]interface{}) *RedisOpts {
	// @todo connection details etc
	return &RedisOpts{}
}

func StrategyInstanceFactory(typeStr string, opts map[string]interface{}) AbstractStrategy {
	switch typeStr {
	case SimpleStrategyType.String():
		fallthrough
	case "": // empty
		return NewSimpleStrategy()
	default:
		panic(fmt.Sprintf("backend  strategy %s not supported", typeStr))
	}
}
