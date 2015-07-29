package collectors

import (
	"reflect"
	"runtime"
	"sync"
	"time"

	"bosun.org/collect"
	"bosun.org/metadata"
	"bosun.org/opentsdb"
)

type StreamCollector struct {
	F      func() <-chan *opentsdb.MultiDataPoint
	Enable func() bool
	name   string
	init   func()

	// internal use
	sync.Mutex
	enabled bool

	TagOverride
}

func (s *StreamCollector) Init() {
	if s.init != nil {
		s.init()
	}
}

func (s *StreamCollector) Run(dpchan chan<- *opentsdb.DataPoint) {
	if s.Enable != nil {
		go func() {
			for {
				next := time.After(time.Minute * 5)
				s.Lock()
				s.enabled = s.Enable()
				s.Unlock()
				<-next
			}
		}()
	}

	inputChan := s.F()
	count := 0
	for md := range inputChan {
		if s.Enabled() {
			for _, dp := range *md {
				dpchan <- dp
				count++
			}

			if !collect.DisableDefaultCollectors {
				tags := opentsdb.TagSet{"collector": s.Name(), "os": runtime.GOOS}
				Add(md, "scollector.collector.count", count, tags, metadata.Counter, metadata.Count, "Counter of metrics passed through.")
			}
		}
	}
}

func (s *StreamCollector) Enabled() bool {
	if s.Enable == nil {
		return true
	}
	s.Lock()
	defer s.Unlock()
	return s.enabled
}

func (s *StreamCollector) Name() string {
	if s.name != "" {
		return s.name
	}
	v := runtime.FuncForPC(reflect.ValueOf(s.F).Pointer())
	return v.Name()

}
