package integration_test

import (
	"fmt"
	"github.com/RobinUS2/tsxdb/client"
	"github.com/RobinUS2/tsxdb/integration"
	"github.com/RobinUS2/tsxdb/server"
	"github.com/RobinUS2/tsxdb/server/backend"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

const token = "verySecure123@#$"

// @todo func TestBatchWritePerformanceRedis(t *testing.T) {
// @todo func TestBatchReadPerformance(t *testing.T) {
// @todo func TestBatchReadPerformanceRedis(t *testing.T) {

func TestRun(t *testing.T) {
	if err := integration.Run(); err != nil {
		t.Error(err)
	}
}

func TestNew(t *testing.T) {
	// start server
	s := NewTestServer(true, true)
	basicTestSuite(t, s)
}

func TestNewRedis(t *testing.T) {
	// start server
	s := NewTestServerRedis(true, true)
	basicTestSuite(t, s)
}

func basicTestSuite(t *testing.T, s *server.Instance) {
	// client
	c := NewTestClient(s)
	if c == nil {
		t.Error()
		return
	}

	// new series
	series := c.Series("mySeries")

	// timestamp
	now := c.Now()
	const oneMinute = 60 * 1000
	const writeValue = 10.1

	// write
	{
		result := series.Write(now, writeValue)
		if result.Error != nil {
			t.Error(result.Error)
		}
	}

	// batch
	{
		b := c.NewBatchWriter()
		if err := b.AddToBatch(c.Series("a"), now+1, 10.0); err != nil {
			t.Error(err)
		}
		if err := b.AddToBatch(c.Series("b"), now+2, 11.1); err != nil {
			t.Error(err)
		}
		if err := b.AddToBatch(c.Series("a"), now+3, 22.2); err != nil {
			t.Error(err)
		}
		result := b.Execute()
		if result.Error != nil {
			t.Error(result.Error)
		}
		if result.NumPersisted != 3 {
			t.Error(result.NumPersisted)
		}
	}

	// read
	{
		result := series.QueryBuilder().From(now - oneMinute).To(now + oneMinute).Execute()
		if result.Error != nil {
			t.Error(result.Error)
		}
		if result.Results == nil {
			t.Error()
		}
		if len(result.Results) != 1 {
			t.Error(result.Results)
			return
		}
		var ts uint64
		var value float64
		for ts, value = range result.Results {
			// no need to do something
		}
		if ts != now {
			t.Error(ts, now)
		}
		if value != writeValue {
			t.Error(value)
		}
		//t.Log(ts, value)
	}

	// batch read
	{
		multiQuery := c.MultiQueryBuilder()
		{
			someSeries := c.Series("a")
			qb := someSeries.QueryBuilder().From(now - oneMinute).To(now + oneMinute)
			if err := multiQuery.AddQuery(qb); err != nil {
				t.Error(err)
			}
		}
		{
			someSeries := c.Series("b")
			qb := someSeries.QueryBuilder().From(now - oneMinute).To(now + oneMinute)
			if err := multiQuery.AddQuery(qb); err != nil {
				t.Error(err)
			}
		}
		res := multiQuery.Execute()
		if res.Error != nil {
			t.Error(res.Error)
		}
		if len(res.Results) != 2 {
			t.Errorf("expected 2 results %+v", res.Results)
		}
	}

	// empty name
	{
		series := c.Series("")
		id, err := series.Create()
		if id != 0 || err == nil {
			t.Error("should error", id, err)
		}
	}

	// invalid name
	{
		series := c.Series("with whitespace")
		id, err := series.Create()
		if id != 0 || err == nil {
			t.Error("should error", id, err)
		}
	}

	{
		// auto batch
		syncFlush := client.NewAutoBatchOptAsyncFlush(false)
		b := c.NewAutoBatchWriter(5, 100*time.Millisecond, syncFlush)
		if b == nil {
			t.Error()
		}
		var flushCount uint64
		b.SetPostFlushFn(func() {
			atomic.AddUint64(&flushCount, 1)
		})

		// series
		series := c.Series("testSeries")
		if err := b.AddToBatch(series, rand.Uint64(), rand.Float64()); err != nil {
			t.Error(err)
		}

		// allow ticker to tick
		time.Sleep(60 * time.Millisecond)

		// check flush count, should still be zero
		if b.FlushCount() != 0 {
			t.Error(b.FlushCount())
		}

		// allow ticker to tick
		time.Sleep(60 * time.Millisecond)

		// check flush count, should still be higher
		if b.FlushCount() != 1 {
			t.Error(b.FlushCount())
		}

		// flushCount?
		if atomic.LoadUint64(&flushCount) != 1 {
			t.Error("should have flushCount", flushCount)
		}

		// hit limit
		for i := 0; i < 10; i++ {
			if err := b.AddToBatch(series, rand.Uint64(), rand.Float64()); err != nil {
				t.Error(err)
			}
		}

		// flushCount?
		if atomic.LoadUint64(&flushCount) != 3 {
			t.Error("should have flushCount", flushCount)
		}

		// close
		if err := b.Close(); err != nil {
			t.Error(err)
		}
	}

	// ASYNC default
	{
		// auto batch
		b := c.NewAutoBatchWriter(5, 100*time.Millisecond)
		if b == nil {
			t.Error()
		}
		if !b.AsyncFlush() {
			t.Error("should be async by default")
		}
		var flushCount uint64
		b.SetPostFlushFn(func() {
			atomic.AddUint64(&flushCount, 1)
		})

		// series
		series := c.Series("testSeries")
		if err := b.AddToBatch(series, rand.Uint64(), rand.Float64()); err != nil {
			t.Error(err)
		}

		// allow ticker to tick
		time.Sleep(60 * time.Millisecond)

		// check flush count, should still be zero
		if b.FlushCount() != 0 {
			t.Error(b.FlushCount())
		}

		// allow ticker to tick
		time.Sleep(60 * time.Millisecond)

		// check flush count, should still be higher
		if b.FlushCount() != 1 {
			t.Error(b.FlushCount())
		}

		// flushCount?
		if atomic.LoadUint64(&flushCount) != 1 {
			t.Error("should have flushCount", flushCount)
		}

		// hit limit
		for i := 0; i < 10; i++ {
			if err := b.AddToBatch(series, rand.Uint64(), rand.Float64()); err != nil {
				t.Error(err)
			}
		}

		time.Sleep(1 * time.Second) // give time to complete async flush @todo actually wait for it, but this is fine for now

		// flushCount?
		if atomic.LoadUint64(&flushCount) != 3 {
			t.Error("should have flushCount", flushCount)
		}

		// close
		if err := b.Close(); err != nil {
			t.Error(err)
		}
	}

	c.Close()
	_ = s.Shutdown()
}

func TestNewNamespace(t *testing.T) {
	// start server
	s := NewTestServer(true, true)
	c := NewTestClient(s)
	if c == nil {
		t.Error()
		return
	}

	// new series
	series := c.Series("mySeries", client.NewSeriesNamespace(1))

	// timestamp
	now := c.Now()
	const oneMinute = 60 * 1000
	const writeValue = 12.3

	// write
	{
		result := series.Write(now, writeValue)
		if result.Error != nil {
			t.Error(result.Error)
		}
	}

	// read
	{
		result := series.QueryBuilder().From(now - oneMinute).To(now + oneMinute).Execute()
		if result.Error != nil {
			t.Error(result.Error)
		}
		if result.Results == nil {
			t.Error()
		}
		if len(result.Results) != 1 {
			t.Error(result.Results)
			return
		}
		var ts uint64
		var value float64
		for ts, value = range result.Results {
			// no need to do something
		}
		if ts != now {
			t.Error(ts, now)
		}
		if value != writeValue {
			t.Error(value)
		}
		//t.Log(ts, value)
	}

	c.Close()
	_ = s.Shutdown()
}

func TestWritePerformance(t *testing.T) {
	// start server
	s := NewTestServer(true, true)
	c := NewTestClient(s)
	now := c.Now()
	startTime := time.Now()
	const minTime = 1 * time.Second
	const minIters = 100
	const writeValue = 10.1
	series := c.Series("benchmarkSeriesWrite")
	var i int
	for i = 0; i < 1000*1000; i++ {
		result := series.Write(now+uint64(i), writeValue)
		if result.Error != nil {
			t.Error(result.Error)
		}
		if i > minIters && i%100 == 0 {
			if time.Since(startTime).Seconds() > minTime.Seconds() {
				break
			}
		}
	}
	tookMs := float64(time.Since(startTime).Nanoseconds()) / 1000000.0
	tookMsEach := tookMs / float64(i)
	perSecond := 1000.0 / tookMsEach
	t.Logf("write avg time %f.2ms (%d iterations - %.0f/second)", tookMsEach, i, perSecond)

	c.Close()
	_ = s.Shutdown()
}

func TestReadPerformance(t *testing.T) {
	// start server
	s := NewTestServer(true, true)
	c := NewTestClient(s)
	now := c.Now()
	series := c.Series("benchmarkSeriesRead")
	const writeValue = 10.1

	// write one value to prevent errors from no data found
	{
		result := series.Write(now, writeValue)
		if result.Error != nil {
			t.Error(result.Error)
		}
	}

	startTime := time.Now()
	const minTime = 1 * time.Second
	const minIters = 100
	const oneMinute = 60 * 1000

	var i int
	for i = 0; i < 1000*1000; i++ {
		result := series.QueryBuilder().From(now - oneMinute).To(now + oneMinute).Execute()
		if result.Error != nil {
			t.Error(result.Error)
		}
		if i > minIters && i%100 == 0 {
			if time.Since(startTime).Seconds() > minTime.Seconds() {
				break
			}
		}
	}
	tookMs := float64(time.Since(startTime).Nanoseconds()) / 1000000.0
	tookMsEach := tookMs / float64(i)
	perSecond := 1000.0 / tookMsEach
	t.Logf("read avg time %f.2ms (%d iterations - %.0f/second)", tookMsEach, i, perSecond)

	c.Close()
	_ = s.Shutdown()
}

func TestNoOpPerformance(t *testing.T) {
	// start server
	s := NewTestServer(true, true)
	c := NewTestClient(s)
	series := c.Series("benchmarkSeriesNoOp")

	startTime := time.Now()
	const minTime = 1 * time.Second
	const minIters = 100

	var i int
	for i = 0; i < 1000*1000; i++ {
		err := series.NoOp()
		if err != nil {
			t.Error(err)
		}
		if i > minIters && i%100 == 0 {
			if time.Since(startTime).Seconds() > minTime.Seconds() {
				break
			}
		}
	}
	tookMs := float64(time.Since(startTime).Nanoseconds()) / 1000000.0
	tookMsEach := tookMs / float64(i)
	t.Logf("noop avg time %f.2ms (%d iterations)", tookMsEach, i)

	c.Close()
	_ = s.Shutdown()
}

func TestInitSeriesPerformance(t *testing.T) {
	// start server
	s := NewTestServer(true, true)
	c := NewTestClient(s)
	now := c.Now()
	startTime := time.Now()
	const minTime = 1 * time.Second
	const minIters = 100
	const writeValue = 10.1
	var i int
	seriesMap := make(map[string]bool)
	numSeriesInited := 0
	numSeriesCreated := 0
	for i = 0; i < 1000*1000; i++ {
		seriesId := i - (i % 10)
		series := c.Series(fmt.Sprintf("benchmarkSeriesInitPerformance-%d", seriesId))
		if _, found := seriesMap[series.Name()]; found {
			numSeriesInited++
		} else {
			numSeriesCreated++
			seriesMap[series.Name()] = true
		}
		result := series.Write(now+uint64(i), writeValue)
		if result.Error != nil {
			t.Error(result.Error)
		}
		if i > minIters && i%100 == 0 {
			if time.Since(startTime).Seconds() > minTime.Seconds() {
				break
			}
		}
	}
	tookMs := float64(time.Since(startTime).Nanoseconds()) / 1000000.0
	tookMsEach := tookMs / float64(i)
	perSecond := 1000.0 / tookMsEach
	t.Logf("init series (+1 write) avg time %f.2ms (%d iterations - %.0f/second)", tookMsEach, i, perSecond)
	stats := s.Statistics()
	t.Logf("%+v", stats)
	if stats.NumSeriesCreated() != uint64(numSeriesCreated) {
		t.Errorf("expected %d was %d", numSeriesCreated, stats.NumSeriesCreated())
	}

	c.Close()
	_ = s.Shutdown()
}

func TestBatchWritePerformance(t *testing.T) {
	// start server
	s := NewTestServer(true, true)
	c := NewTestClient(s)
	startTime := time.Now()
	totalValuesWritten := 0
	const minTime = 1 * time.Second
	const minIters = 100
	const batchSize = 1000 // tuning this number increases throughput, seems to max out at around 100K value with throughput of 1.7MM/sec on 1 core @ MacBook Pro Feb '18, although that is not realistic so we leave it at 1000 for now
	series := c.Series("benchmarkSeriesWriteBatch")
	var i int
	for i = 0; i < 1000*1000; i++ {
		b := c.NewBatchWriter()
		// batches
		for j := 0; j < batchSize; j++ {
			totalValuesWritten++
			if err := b.AddToBatch(series, rand.Uint64(), rand.Float64()); err != nil {
				t.Error(err)
			}
		}
		result := b.Execute()
		if result.Error != nil {
			t.Error(result.Error)
		}
		if i > minIters && i%100 == 0 {
			if time.Since(startTime).Seconds() > minTime.Seconds() {
				break
			}
		}
	}
	tookMs := float64(time.Since(startTime).Nanoseconds()) / 1000000.0
	tookMsEach := tookMs / float64(i*batchSize)
	perSecond := 1000.0 / tookMsEach
	numIterations := i + 1
	t.Logf("write avg time %f.2ms (%d iterations - %.0f/second)", tookMsEach, numIterations, perSecond)

	time.Sleep(1 * time.Second) // wait for async to flush @todo wait for it to really complete

	stats := s.Statistics()
	t.Logf("%+v totalValuesWritten %d", stats, totalValuesWritten)
	if stats.NumValuesWritten() != uint64(totalValuesWritten) {
		t.Errorf("lost writes %d vs %d", stats.NumValuesWritten(), totalValuesWritten)
	}
	if stats.NumSeriesCreated() != 1 {
		t.Errorf("1 serie")
	}
	if stats.NumSeriesInitialised() != 0 {
		t.Errorf("init is only if we redo an existing one")
	}
	if stats.NumAuthentications() < uint64(numIterations) {
		t.Errorf("at least 1 auth per flush %d vs %d", stats.NumAuthentications(), numIterations)
	}
	if stats.NumReads() != 0 {
		t.Errorf("no reads")
	}

	c.Close()
	_ = s.Shutdown()
}

func runBatchWritePerformanceMultiSeries(t *testing.T, s *server.Instance) {
	c := NewTestClient(s)
	startTime := time.Now()
	totalValuesWritten := 0
	const minTime = 1 * time.Second
	const minIters = 100
	const batchSize = 1000 // tuning this number increases throughput, seems to max out at around 100K value with throughput of 1.7MM/sec on 1 core @ MacBook Pro Feb '18, although that is not realistic so we leave it at 1000 for now
	var i int
	for i = 0; i < 1000*1000; i++ {
		b := c.NewBatchWriter()
		// batches
		for j := 0; j < batchSize; j++ {
			totalValuesWritten++
			seriesId := i % 100
			series := c.Series(fmt.Sprintf("benchmarkSeriesWriteBatchMultiSeries-%d", seriesId))
			if err := b.AddToBatch(series, rand.Uint64(), rand.Float64()); err != nil {
				t.Error(err)
			}
		}
		startExecute := time.Now()
		result := b.Execute()
		t.Logf("execute took %s", time.Since(startExecute))
		if result.Error != nil {
			t.Error(result.Error)
		}

		// evict series cache
		if i%20 == 0 {
			// simulate retransmission of metadata
			if c.SeriesPool().EvictCache() < 1 {
				t.Error("should evict")
			}
		}

		if i > minIters && i%10 == 0 {
			if time.Since(startTime).Seconds() > minTime.Seconds() {
				break
			}
		}
	}
	tookMs := float64(time.Since(startTime).Nanoseconds()) / 1000000.0
	tookMsEach := tookMs / float64(i*batchSize)
	perSecond := 1000.0 / tookMsEach
	numIterations := i + 1
	t.Logf("write avg time %f.2ms (%d iterations - %.0f/second)", tookMsEach, numIterations, perSecond)

	time.Sleep(1 * time.Second) // wait for async to flush @todo wait for it to really complete

	stats := s.Statistics()
	t.Logf("%+v totalValuesWritten %d", stats, totalValuesWritten)
	if stats.NumValuesWritten() != uint64(totalValuesWritten) {
		t.Errorf("lost writes %d vs %d", stats.NumValuesWritten(), totalValuesWritten)
	}
	if stats.NumSeriesCreated() != 100 {
		t.Errorf("100 series expected was %d", stats.NumSeriesCreated())
	}
	if stats.NumSeriesInitialised() < 1 {
		t.Errorf("init should be done a few times")
	}
	if stats.NumAuthentications() < uint64(numIterations) {
		t.Errorf("at least 1 auth per flush %d vs %d", stats.NumAuthentications(), numIterations)
	}
	if stats.NumReads() != 0 {
		t.Errorf("no reads")
	}

	c.Close()
	_ = s.Shutdown()
}

func TestBatchWritePerformanceMultiSeries(t *testing.T) {
	// start server
	s := NewTestServer(true, true)
	runBatchWritePerformanceMultiSeries(t, s)
}

func TestBatchWritePerformanceMultiSeriesRedis(t *testing.T) {
	// start server
	s := NewTestServerRedis(true, true)
	runBatchWritePerformanceMultiSeries(t, s)
}

// during a restart of the (memory) server it could be that metadata is lost, in such a way that clients need to re-transmit this
func TestServerRestartClientResendMetadata(t *testing.T) {
	// start server
	s := NewTestServer(true, true)
	c := NewTestClient(s)

	series := c.Series("TestServerRestartClientResendMetadata")
	result := series.Write(1, 1.0)
	if result.Error != nil {
		t.Error(result.Error)
	}

	// clear metadata
	m := s.MetaStore()
	if err := m.Clear(); err != nil {
		t.Error(err)
	}

	// write again
	{
		result := series.Write(1, 1.0)
		if result.Error != nil {
			t.Error(result.Error)
		}
	}

	c.Close()
	_ = s.Shutdown()
}

func NewTestClient(server *server.Instance) *client.Instance {
	opts := client.NewOpts()
	if server != nil {
		opts.ListenPort = server.Opts().ListenPort
		opts.ListenHost = server.Opts().ListenHost
		opts.AuthToken = server.Opts().AuthToken
		opts.OptsConnection.Debug = true
	}
	c := client.New(opts)
	return c
}

var lastPort uint64 = 1234

func NewTestServer(init bool, listen bool) *server.Instance {
	port := atomic.AddUint64(&lastPort, 1)
	opts := server.NewOpts()
	opts.ListenPort = int(port)
	opts.AuthToken = token
	s := server.New(opts)
	if init {
		if err := s.Init(); err != nil {
			panic(err)
		}
	}
	if listen {
		if err := s.StartListening(); err != nil {
			panic(err)
		}
	}
	return s
}

func NewTestServerRedis(init bool, listen bool) *server.Instance {
	port := atomic.AddUint64(&lastPort, 1)
	opts := server.NewOpts()
	opts.ListenPort = int(port)
	opts.AuthToken = token
	opts.Backends = []server.BackendOpts{
		{
			Type:       "redis",
			Identifier: backend.DefaultIdentifier,
			Metadata:   true,
			Options: map[string]interface{}{
				backend.RedisOptsKey: []interface{}{
					backend.RedisConnectionDetails{
						Addr: "127.0.0.1",
						Port: 6379,
						Type: backend.RedisMemory,
					},
				},
			},
		},
	}
	s := server.New(opts)
	if init {
		if err := s.Init(); err != nil {
			panic(err)
		}
	}
	if listen {
		if err := s.StartListening(); err != nil {
			panic(err)
		}
	}
	return s
}
