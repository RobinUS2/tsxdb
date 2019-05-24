module github.com/RobinUS2/tsxdb/client

go 1.12

replace github.com/RobinUS2/tsxdb/client => ../client

replace github.com/RobinUS2/tsxdb/rpc => ../rpc

replace github.com/RobinUS2/tsxdb/rpc/types => ../rpc/types

replace github.com/RobinUS2/tsxdb/tools => ../tools

require (
	github.com/RobinUS2/tsxdb/rpc v0.0.0-20190524075157-5e90bb459d1e
	github.com/RobinUS2/tsxdb/tools v0.0.0-20190517163515-dd67f46aab79
	github.com/pkg/errors v0.8.1
)
