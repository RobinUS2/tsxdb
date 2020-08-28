module github.com/RobinUS2/tsxdb/integration

go 1.13

replace github.com/RobinUS2/tsxdb/client => ../client

replace github.com/RobinUS2/tsxdb/rpc => ../rpc

replace github.com/RobinUS2/tsxdb/rpc/types => ../rpc/types

replace github.com/RobinUS2/tsxdb/server => ../server

replace github.com/RobinUS2/tsxdb/tools => ../tools

replace github.com/RobinUS2/tsxdb/telnet => ../telnet

require (
	github.com/RobinUS2/tsxdb/client v0.0.0-20200828094023-5f020cf85e95
	github.com/RobinUS2/tsxdb/rpc v0.0.0-20200828082527-116be78017fa
	github.com/RobinUS2/tsxdb/server v0.0.0-20190523121601-0130f23bf035
)
