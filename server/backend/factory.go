package backend

import (
	"fmt"
	"gopkg.in/yaml.v2"
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
	details := make(map[Namespace]RedisConnectionDetails)
	namespacesIf, ok := opts["redis"]
	if !ok {
		panic("no redis opts")
	}
	namespaces := namespacesIf.([]interface{}) // allow to panic, converting should succeed, panic otherwise
	for idx, namespace := range namespaces {
		// namespace is of type map[interface{}]interface{}, convert back to yaml before unmarshalling again to
		// the correct RedisConnectionDetails type
		yamlBytes, err := yaml.Marshal(namespace)
		if err != nil {
			panic(err)
		}
		var connectionDetails RedisConnectionDetails
		err = yaml.Unmarshal(yamlBytes, &connectionDetails)
		if err != nil {
			panic(err)
		}

		details[Namespace(idx)] = connectionDetails
	}

	if len(details) == 0 {
		panic("no namespaces found")
	}
	return &RedisOpts{
		details,
	}
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
