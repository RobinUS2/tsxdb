module github.com/RobinUS2/tsxdb/server

go 1.13

replace github.com/RobinUS2/tsxdb/rpc => ../rpc

replace github.com/RobinUS2/tsxdb/rpc/types => ../rpc/types

replace github.com/RobinUS2/tsxdb/server => ../server

replace github.com/RobinUS2/tsxdb/server/backend => ../server/backend

replace github.com/RobinUS2/tsxdb/server/rollup => ../server/rollup

replace github.com/RobinUS2/tsxdb/tools => ../tools

replace github.com/RobinUS2/tsxdb/telnet => ../telnet

require (
	github.com/RobinUS2/tsxdb/rpc v0.0.0-20200831110925-b62f451e618d
	github.com/RobinUS2/tsxdb/telnet v0.0.0-20200901125404-22137cdbe6ba
	github.com/RobinUS2/tsxdb/tools v0.0.0-20200901125404-22137cdbe6ba
	github.com/alicebob/miniredis/v2 v2.30.0
	github.com/bsm/redislock v0.7.2
	github.com/go-redis/redis/v8 v8.11.5
	github.com/hashicorp/golang-lru v0.5.0
	github.com/jinzhu/now v1.1.5
	github.com/karlseguin/ccache/v2 v2.0.8
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apimachinery v0.15.7
)
