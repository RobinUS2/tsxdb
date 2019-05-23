package integration_test

import (
	"fmt"
	"github.com/RobinUS2/tsxdb/client"
	"github.com/RobinUS2/tsxdb/integration"
	"github.com/RobinUS2/tsxdb/server"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

const token = "verySecure123@#$"

func TestRun(t *testing.T) {
	if err := integration.Run(); err != nil {
		t.Error(err)
	}
}

func TestNew(t *testing.T) {
	// start server
	s := NewTestServer(true, true)
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
		if err := b.AddToBatch(c.Series("a"), 12345, 10.0); err != nil {
			t.Error(err)
		}
		if err := b.AddToBatch(c.Series("b"), 54321, 11.1); err != nil {
			t.Error(err)
		}
		if err := b.AddToBatch(c.Series("a"), 12345, 22.2); err != nil {
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
		b := c.NewAutoBatchWriter(5, 100*time.Millisecond)
		if b == nil {
			t.Error()
		}
		flushed := false
		b.SetPostFlushFn(func() {
			flushed = true
		})

		// series
		series := c.Series("testSeries")
		if err := b.AddToBatch(series, rand.Uint64(), rand.Float64()); err != nil {
			t.Error(err)
		}

		// allow ticker to tick
		time.Sleep(50 * time.Millisecond)

		// check flush count
		if b.FlushCount() < 1 {
			t.Error(b.FlushCount())
		}

		// flushed?
		if !flushed {
			t.Error("should have flushed")
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
	for i = 0; i < 1000*1000; i++ {
		seriesId := i - (i % 10)
		series := c.Series(fmt.Sprintf("benchmarkSeriesInitPerformance-%d", seriesId))
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

	c.Close()
	_ = s.Shutdown()
}

func TestBatchWritePerformance(t *testing.T) {
	// start server
	s := NewTestServer(true, true)
	c := NewTestClient(s)
	startTime := time.Now()
	const minTime = 1 * time.Second
	const minIters = 100
	const batchSize = 1000 // tuning this number increases throughput, seems to max out at around 100K value with throughput of 1.7MM/sec on 1 core @ MacBook Pro Feb '18, although that is not realistic so we leave it at 1000 for now
	series := c.Series("benchmarkSeriesWriteBatch")
	var i int
	for i = 0; i < 1000*1000; i++ {
		b := c.NewBatchWriter()
		// batches
		for j := 0; j < batchSize; j++ {
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
	t.Logf("write avg time %f.2ms (%d iterations - %.0f/second)", tookMsEach, i, perSecond)

	c.Close()
	_ = s.Shutdown()
}

func NewTestClient(server *server.Instance) *client.Instance {
	opts := client.NewOpts()
	if server != nil {
		opts.ListenPort = server.Opts().ListenPort
		opts.ListenHost = server.Opts().ListenHost
		opts.AuthToken = server.Opts().AuthToken
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
