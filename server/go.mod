module github.com/RobinUS2/tsxdb/server

go 1.12

replace github.com/RobinUS2/tsxdb/rpc => ../rpc

replace github.com/RobinUS2/tsxdb/rpc/types => ../rpc/types

replace github.com/RobinUS2/tsxdb/server => ../server

replace github.com/RobinUS2/tsxdb/server/backend => ../server/backend

replace github.com/RobinUS2/tsxdb/server/rollup => ../server/rollup

replace github.com/RobinUS2/tsxdb/tools => ../tools

replace github.com/RobinUS2/tsxdb/telnet => ../telnet

require (
	github.com/RobinUS2/tsxdb/rpc v0.0.0-20190522063027-1933ca59773d
	github.com/RobinUS2/tsxdb/telnet v0.0.0-20190523121536-3a83a2461764
	github.com/RobinUS2/tsxdb/tools v0.0.0-20190518132317-4b1ff7c46623
	github.com/bsm/redis-lock v8.0.0+incompatible
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/pkg/errors v0.8.1
	k8s.io/apimachinery v0.0.0-20190515023456-b74e4c97951f
)
