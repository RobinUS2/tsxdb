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
	github.com/RobinUS2/tsxdb/rpc v0.0.0-20200824100554-7dd869f40ec7
	github.com/RobinUS2/tsxdb/telnet v0.0.0-20200824100554-7dd869f40ec7
	github.com/RobinUS2/tsxdb/tools v0.0.0-20200824100554-7dd869f40ec7
	github.com/alicebob/miniredis/v2 v2.13.2
	github.com/bsm/redislock v0.5.0
	github.com/go-redis/redis/v7 v7.4.0
	github.com/pkg/errors v0.9.1
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/apimachinery v0.0.0-20190515023456-b74e4c97951f
)
