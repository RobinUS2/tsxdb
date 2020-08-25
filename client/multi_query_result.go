package client

type MultiQueryResult struct {
	Error   error
	Results []QueryResult // indexed by the queries provided
}
