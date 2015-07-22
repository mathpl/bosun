package collectors

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"bosun.org/opentsdb"
)

func init() {
	collectors = append(collectors, &StreamCollector{F: c_local_listener})
}

func c_local_listener() <-chan *opentsdb.MultiDataPoint {
	listenAddr := "localhost:4240"

	pm := &putMetric{}
	pm.localMetrics = make(chan *opentsdb.MultiDataPoint, 1)

	mux := http.NewServeMux()
	mux.Handle("/api/put", pm)
	//router.Handle("/api/metadata/put", simpleJSON(PutMetadata))
	go http.ListenAndServe(listenAddr, mux)

	return pm.localMetrics
}

type putMetric struct {
	localMetrics chan *opentsdb.MultiDataPoint
}

func (pm *putMetric) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		bodyReader io.ReadCloser
		err        error
	)

	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}

	if r.Header.Get("Content-Encoding") == "gzip" {
		if bodyReader, err = gzip.NewReader(r.Body); err != nil {
			w.WriteHeader(500)
			return
		}
	} else {
		bodyReader = r.Body
	}

	if body, err := ioutil.ReadAll(bodyReader); err != nil {
		w.WriteHeader(500)
		return
	} else {
		bodyReader.Close()

		var (
			dp  *opentsdb.DataPoint
			mdp opentsdb.MultiDataPoint
		)

		if err := json.Unmarshal(body, &mdp); err == nil {
		} else if err = json.Unmarshal(body, &dp); err == nil {
			mdp = opentsdb.MultiDataPoint{dp}
		} else {
			fmt.Printf("Unable to decode: %s\n", err)
		}

		for _, dp := range mdp {
			dp.Tags = AddTags.Copy().Merge(dp.Tags)
		}

		pm.localMetrics <- &mdp

		w.WriteHeader(204)
	}
}

//func PutMetadata(w http.ResponseWriter, r *http.Request) (interface{}, error) {
//	d := json.NewDecoder(r.Body)
//	var ms []metadata.Metasend
//	if err := d.Decode(&ms); err != nil {
//		return nil, err
//	}
//	for _, m := range ms {
//		schedule.PutMetadata(metadata.Metakey{
//			Metric: m.Metric,
//			Tags:   m.Tags.Tags(),
//			Name:   m.Name,
//		}, m.Value)
//	}
//	w.WriteHeader(204)
//	return nil, nil
//}
