module github.com/RobinUS2/tsxdb/tools

go 1.13

replace github.com/RobinUS2/tsxdb/rpc => ../rpc

replace github.com/RobinUS2/tsxdb/rpc/types => ../rpc/types

require (
	github.com/OneOfOne/xxhash v1.2.8
	github.com/RobinUS2/tsxdb/rpc v0.0.0-20200825111516-afed105674ab
	github.com/kr/pretty v0.1.0 // indirect
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/satori/go.uuid v1.2.0 => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
