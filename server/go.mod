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
	github.com/RobinUS2/tsxdb/rpc v0.0.0-20200827170956-bf7df6ad9a7f
	github.com/RobinUS2/tsxdb/telnet v0.0.0-20200827184906-63c34ab0136a
	github.com/RobinUS2/tsxdb/tools v0.0.0-20200827170956-bf7df6ad9a7f
	github.com/alicebob/miniredis/v2 v2.13.2
	github.com/bsm/redislock v0.5.0
	github.com/go-redis/redis/v7 v7.4.0
	github.com/hashicorp/golang-lru v0.5.0
	github.com/karlseguin/ccache/v2 v2.0.6
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/apimachinery v0.0.0-20190515023456-b74e4c97951f
)
