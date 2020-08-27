module github.com/RobinUS2/tsxdb/integration

go 1.13

replace github.com/RobinUS2/tsxdb/client => ../client

replace github.com/RobinUS2/tsxdb/rpc => ../rpc

replace github.com/RobinUS2/tsxdb/rpc/types => ../rpc/types

replace github.com/RobinUS2/tsxdb/server => ../server

replace github.com/RobinUS2/tsxdb/tools => ../tools

replace github.com/RobinUS2/tsxdb/telnet => ../telnet

require (
	github.com/RobinUS2/tsxdb/client v0.0.0-20200827184906-63c34ab0136a
	github.com/RobinUS2/tsxdb/server v0.0.0-20190523121601-0130f23bf035
)
