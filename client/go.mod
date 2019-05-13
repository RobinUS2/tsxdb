module github.com/RobinUS2/tsxdb/client

go 1.12

replace github.com/RobinUS2/tsxdb/client => ../client

replace github.com/RobinUS2/tsxdb/rpc => ../rpc

replace github.com/RobinUS2/tsxdb/server => ../server

replace github.com/RobinUS2/tsxdb/tools => ../tools

require (
	github.com/OneOfOne/xxhash v1.2.5 // indirect
	github.com/RobinUS2/tsxdb/rpc v0.0.0-00010101000000-000000000000
	github.com/RobinUS2/tsxdb/server v0.0.0-00010101000000-000000000000
	github.com/RobinUS2/tsxdb/tools v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.8.1
	github.com/satori/go.uuid v1.2.0 // indirect
)
