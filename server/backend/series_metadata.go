package backend

type Namespace int
type Series uint64
type Timestamp float64

type SeriesMetadata struct {
	Id        Series
	Namespace Namespace
	Name      string
	Tags      []string `json:",omitempty"`
	TtlExpire uint64   `json:",omitempty"` //  0 OR time in the future in seconds
}
