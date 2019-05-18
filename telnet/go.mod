module github.com/RobinUS2/tsxdb/telnet

go 1.12

replace github.com/RobinUS2/tsxdb/telnet => ../telnet

replace github.com/RobinUS2/tsxdb/client => ../client

replace github.com/RobinUS2/tsxdb/server => ../server

require (
	github.com/RobinUS2/tsxdb/client v0.0.0-20190517163515-dd67f46aab79
	github.com/RobinUS2/tsxdb/server v0.0.0-20190517163515-dd67f46aab79
	github.com/pkg/errors v0.8.1
	github.com/reiver/go-oi v0.0.0-20160325061615-431c83978379
	github.com/reiver/go-telnet v0.0.0-20180421082511-9ff0b2ab096e
)
