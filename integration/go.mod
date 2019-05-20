module github.com/RobinUS2/tsxdb/integration

go 1.12

replace github.com/RobinUS2/tsxdb/client => ../client

replace github.com/RobinUS2/tsxdb/rpc => ../rpc

replace github.com/RobinUS2/tsxdb/rpc/types => ../rpc/types

replace github.com/RobinUS2/tsxdb/server => ../server

replace github.com/RobinUS2/tsxdb/tools => ../tools

replace github.com/RobinUS2/tsxdb/telnet => ../telnet

require (
	github.com/RobinUS2/tsxdb/client v0.0.0-20190520190210-0fe97f2fc0ef
	github.com/RobinUS2/tsxdb/rpc v0.0.0-20190518132317-4b1ff7c46623
	github.com/RobinUS2/tsxdb/server v0.0.0-20190518132317-4b1ff7c46623
	github.com/RobinUS2/tsxdb/tools v0.0.0-20190518132317-4b1ff7c46623
	github.com/pkg/errors v0.8.1
)
