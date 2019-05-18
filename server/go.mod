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
	github.com/RobinUS2/tsxdb/rpc v0.0.0-20190513114607-96dd167a5920
	github.com/RobinUS2/tsxdb/telnet v0.0.0-20190517163515-dd67f46aab79
	github.com/RobinUS2/tsxdb/tools v0.0.0-20190517163515-dd67f46aab79
	github.com/bsm/redis-lock v8.0.0+incompatible
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/pkg/errors v0.8.1
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/apimachinery v0.0.0-20190511063452-5b67e417bf61
)
