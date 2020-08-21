module github.com/RobinUS2/tsxdb/client

go 1.13

replace github.com/RobinUS2/tsxdb/client => ../client

replace github.com/RobinUS2/tsxdb/rpc => ../rpc

replace github.com/RobinUS2/tsxdb/rpc/types => ../rpc/types

replace github.com/RobinUS2/tsxdb/tools => ../tools

require (
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/RobinUS2/tsxdb/rpc v0.0.0-20200821115332-b962b83da4f6
	github.com/RobinUS2/tsxdb/tools v0.0.0-20200821115332-b962b83da4f6
	github.com/karlseguin/ccache/v2 v2.0.6
	github.com/pkg/errors v0.9.1
	gopkg.in/yaml.v2 v2.3.0 // indirect
)
