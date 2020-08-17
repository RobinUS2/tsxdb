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
	github.com/RobinUS2/tsxdb/rpc v0.0.0-20190924121018-2ae17e334935
	github.com/RobinUS2/tsxdb/telnet v0.0.0-20190523130403-80b35d96abae
	github.com/RobinUS2/tsxdb/tools v0.0.0-20190518132317-4b1ff7c46623
	github.com/bsm/redislock v0.5.0
	github.com/go-redis/redis/v7 v7.2.0
	github.com/pkg/errors v0.8.1
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/apimachinery v0.0.0-20190515023456-b74e4c97951f
)
