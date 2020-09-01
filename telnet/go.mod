module github.com/RobinUS2/tsxdb/telnet

go 1.13

replace github.com/RobinUS2/tsxdb/telnet => ../telnet

replace github.com/RobinUS2/tsxdb/client => ../client

replace github.com/RobinUS2/tsxdb/server => ../server

replace github.com/RobinUS2/tsxdb/rpc => ../rpc

require (
	github.com/RobinUS2/tsxdb/client v0.0.0-20200831110925-b62f451e618d
	github.com/RobinUS2/tsxdb/server v0.0.0-20190523121601-0130f23bf035
	github.com/pkg/errors v0.9.1
	github.com/reiver/go-oi v1.0.0
	github.com/reiver/go-telnet v0.0.0-20180421082511-9ff0b2ab096e
)
