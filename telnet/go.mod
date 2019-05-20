module github.com/RobinUS2/tsxdb/telnet

go 1.12

replace github.com/RobinUS2/tsxdb/telnet => ../telnet

replace github.com/RobinUS2/tsxdb/client => ../client

replace github.com/RobinUS2/tsxdb/server => ../server

require (
	github.com/RobinUS2/tsxdb/client v0.0.0-20190520190210-0fe97f2fc0ef
	github.com/RobinUS2/tsxdb/server v0.0.0-20190518132317-4b1ff7c46623
	github.com/pkg/errors v0.8.1
	github.com/reiver/go-oi v0.0.0-20160325061615-431c83978379
	github.com/reiver/go-telnet v0.0.0-20180421082511-9ff0b2ab096e
)
