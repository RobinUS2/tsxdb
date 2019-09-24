package backend

import (
	"errors"
	"github.com/RobinUS2/tsxdb/rpc/types"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

const maxPaddingSize = 0.1
const MemoryType = TypeBackend("memory")

// @todo partition by timestamp!!!

type MemoryBackend struct {
	// data
	data    map[Namespace]map[Series]map[Timestamp]float64
	dataMux sync.RWMutex

	// metadata
	seriesIdCounter uint64
	series          map[Series]*SeriesMetadata
	seriesMux       sync.RWMutex

	AbstractBackend
}

func (instance *MemoryBackend) Type() TypeBackend {
	return MemoryType
}

func (instance *MemoryBackend) Write(context ContextWrite, timestamps []uint64, values []float64) error {
	if len(timestamps) != len(values) {
		return errors.New("mismatch pairs")
	}

	namespace := Namespace(context.Namespace)
	seriesId := Series(context.Series)

	// obtain write lock
	instance.dataMux.Lock()

	// init maps
	instance.__notLockedInitMaps(context.Context, true)

	// execute writes
	for idx, timestamp := range timestamps {
		value := values[idx]

		// pad timestamp with random number to make sure we can have multiple per actual time "padded decimals"
		tsWithRand := float64(timestamp) + (rand.Float64() * maxPaddingSize)

		// write
		instance.data[namespace][seriesId][Timestamp(tsWithRand)] = value
	}

	// unlock
	instance.dataMux.Unlock()

	return nil
}

func (instance *MemoryBackend) GetSeriesMeta(s Series) *SeriesMetadata {
	instance.seriesMux.RLock()
	v := instance.series[s]
	instance.seriesMux.RUnlock()
	return v
}

// this does NOT lock the instance.data variable
func (instance *MemoryBackend) __notLockedInitMaps(context Context, autoCreate bool) (available bool) {
	namespace := Namespace(context.Namespace)
	if _, found := instance.data[namespace]; !found {
		if !autoCreate {
			return false
		}
		instance.data[namespace] = make(map[Series]map[Timestamp]float64)
	}
	series := Series(context.Series)
	if _, found := instance.data[namespace][series]; !found {
		if !autoCreate {
			return false
		}
		instance.data[namespace][series] = make(map[Timestamp]float64)
	}

	// data exists, fetch metadata
	meta := instance.GetSeriesMeta(series)
	if meta == nil {
		// this could happen in case of a restart of the server (while the client still believes the series is already initialized)
		panic(types.RpcErrorBackendMetadataNotFound)
	}

	// ttl of series
	if meta.TtlExpire > 0 {
		nowSeconds := nowSeconds()
		if meta.TtlExpire < nowSeconds {
			// expired, remove it
			res := instance.ReverseApi().DeleteSeries(&DeleteSeries{
				Series: []types.SeriesIdentifier{
					{
						Namespace: context.Namespace,
						Id:        context.Series,
					},
				},
			})
			if res.Error != nil {
				// @todo deal with in other way
				panic(res.Error)
			}
			return false
		}
	}

	return true
}

func nowSeconds() uint64 {
	return uint64(time.Now().Unix())
}

var ErrNoDataFound = errors.New("no data found")

func (instance *MemoryBackend) Read(context ContextRead) (res ReadResult) {
	namespace := Namespace(context.Namespace)
	seriesId := Series(context.Series)
	instance.dataMux.RLock()
	if !instance.__notLockedInitMaps(context.Context, false) {
		// not available in the store
		res.Error = ErrNoDataFound
		instance.dataMux.RUnlock()
		return
	}
	series := instance.data[namespace][seriesId]

	// prune
	var pruned map[uint64]float64
	fromFloat := float64(context.From)
	toFloat := float64(context.To) + maxPaddingSize // add a bit here since that's the maximum value of the padding
	for tsF, value := range series {
		ts := float64(tsF)
		if ts < fromFloat || ts > toFloat {
			continue
		}
		if pruned == nil {
			// lazy init map, since it could be very well that we have no data
			pruned = make(map[uint64]float64)
		}
		// truncate timestamp to get rid of the padded decimals
		pruned[uint64(ts)] = value
	}

	// unlock series data
	instance.dataMux.RUnlock()

	res.Results = pruned

	return
}

func (instance *MemoryBackend) __notLockedGetSeriesByNameSpaceAndName(namespace Namespace, name string) *SeriesMetadata {
	for _, serie := range instance.series {
		if serie.Namespace != namespace {
			continue
		}
		if serie.Name == name {
			return serie
		}
	}
	return nil
}

func (instance *MemoryBackend) CreateOrUpdateSeries(create *CreateSeries) (result *CreateSeriesResult) {
	result = &CreateSeriesResult{
		Results: make(map[types.SeriesCreateIdentifier]types.SeriesMetadataResponse),
	}

	var newSeries []types.SeriesCreateMetadata
	instance.seriesMux.RLock()
	for _, serie := range create.Series {
		existing := instance.__notLockedGetSeriesByNameSpaceAndName(Namespace(serie.Namespace), serie.Name)
		if existing != nil {
			// @todo support changes (e.g. adding tags)
			// return existing metadata
			result.Results[serie.SeriesCreateIdentifier] = types.SeriesMetadataResponse{
				Id:                     uint64(existing.Id),
				Error:                  nil,
				SeriesCreateIdentifier: serie.SeriesCreateIdentifier,
				New:                    false,
			}
			continue
		}
		if newSeries == nil {
			newSeries = make([]types.SeriesCreateMetadata, 0)
		}
		newSeries = append(newSeries, serie)
	}
	instance.seriesMux.RUnlock()

	// add
	if newSeries != nil {
		// write lock
		instance.seriesMux.Lock()
		for _, serie := range newSeries {
			// check existing again, now with write barrier globally
			existing := instance.__notLockedGetSeriesByNameSpaceAndName(Namespace(serie.Namespace), serie.Name)
			if existing != nil {
				continue
			}

			// increment id
			id := atomic.AddUint64(&instance.seriesIdCounter, 1)

			// add to memory
			var ttlExpire uint64
			if serie.Ttl > 0 {
				ttlExpire = nowSeconds() + uint64(serie.Ttl)
			}
			instance.series[Series(id)] = &SeriesMetadata{
				Namespace: Namespace(serie.Namespace),
				Name:      serie.Name,
				Id:        Series(id),
				Tags:      serie.Tags,
				TtlExpire: ttlExpire,
			}

			// result
			result.Results[serie.SeriesCreateIdentifier] = types.SeriesMetadataResponse{
				Id:                     id,
				Error:                  nil,
				SeriesCreateIdentifier: serie.SeriesCreateIdentifier,
				New:                    true,
			}
		}
		instance.seriesMux.Unlock()
	}

	return
}

func (instance *MemoryBackend) SearchSeries(search *SearchSeries) (result *SearchSeriesResult) {
	result = &SearchSeriesResult{
		Series: nil, // lazy init
	}
	if search.And != nil {
		result.Error = errors.New("no AND support yet")
		return
	}
	if search.Or != nil {
		result.Error = errors.New("no OR support yet")
		return
	}
	if search.Comparator != SearchSeriesComparatorEquals {
		result.Error = errors.New("only EQUALS support")
		return
	}
	if search.Tag != "" {
		result.Error = errors.New("not tag support yet")
		return
	}

	// search
	instance.seriesMux.RLock()
	for _, serie := range instance.series {
		if serie.Namespace != Namespace(search.Namespace) {
			continue
		}
		if serie.Name == search.Name {
			// @todo support tags, etc
			// match

			// init result set
			if result.Series == nil {
				result.Series = make([]types.SeriesIdentifier, 0)
			}
			result.Series = append(result.Series, types.SeriesIdentifier{
				Namespace: int(serie.Namespace),
				Id:        uint64(serie.Id),
			})
		}
	}
	instance.seriesMux.RUnlock()

	return
}

func (instance *MemoryBackend) DeleteSeries(ops *DeleteSeries) (result *DeleteSeriesResult) {
	result = &DeleteSeriesResult{}
	instance.seriesMux.Lock()
	defer instance.seriesMux.Unlock()
	for _, deleteOperation := range ops.Series {
		// check correct namespace
		key := Series(deleteOperation.Id)
		if val, found := instance.series[key]; found {
			if val.Namespace != Namespace(deleteOperation.Namespace) {
				result.Error = errors.New("invalid namespace")
				return
			}
		} else {
			// not found
			result.Error = errors.New("not found")
			return
		}
		delete(instance.series, key)
	}
	return
}

func (instance *MemoryBackend) Clear() error {
	instance.seriesMux.Lock()
	instance.dataMux.Lock()
	instance.data = map[Namespace]map[Series]map[Timestamp]float64{}
	instance.series = map[Series]*SeriesMetadata{}
	instance.seriesIdCounter = 0
	instance.dataMux.Unlock()
	instance.seriesMux.Unlock()
	return nil
}

func (instance *MemoryBackend) Init() error {
	return nil
}

func NewMemoryBackend() *MemoryBackend {
	m := &MemoryBackend{
		data:   make(map[Namespace]map[Series]map[Timestamp]float64),
		series: make(map[Series]*SeriesMetadata),
	}
	if err := m.Clear(); err != nil {
		// clear should always work for in-memory
		panic(err)
	}
	return m
}
