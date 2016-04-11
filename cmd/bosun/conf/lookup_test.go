package conf

import (
	"testing"

	"bosun.org/opentsdb"
)

func TestLookup(t *testing.T) {
	e1 := &Entry{
		Def:  "raw_text",
		Name: "service=infra/*,site=*",
		ExprEntry: &ExprEntry{
			AlertKey: "{service=infra/*,site=*}",
			Values:   map[string]string{"pd_service_key": "NOMATCH"},
		},
	}

	e2 := &Entry{
		Def:  "raw_text",
		Name: "service=my-service,site=boston-1|dallas-1",
		ExprEntry: &ExprEntry{
			AlertKey: "{service=my-service,site=boston-1|dallas-1}",
			Values:   map[string]string{"pd_service_key": "MATCH"},
		},
	}

	eDefault := &Entry{
		Def:  "raw_text",
		Name: "service=*,site=*",
		ExprEntry: &ExprEntry{
			AlertKey: "{service=*,site=*}",
			Values:   map[string]string{"pd_service_key": ""},
		},
	}

	l := &Lookup{Text: "raw_text", Name: "lookup_section", Tags: []string{"site", "service"}, Entries: []*Entry{e1, e2, eDefault}, UnjoinedOK: false}

	e := l.ToExpr()

	v, _ := e.Get("pd_service_key", opentsdb.TagSet{"service": "my-service", "site": "boston-1"})
	if v != "MATCH" {
		t.Error("Full match failed")
	}

	v, _ = e.Get("pd_service_key", opentsdb.TagSet{"service": "my-service", "site": "test-1"})
	if v != "" {
		t.Error("Wrong match shouldn't work:")
	}

	v, _ = e.Get("pd_service_key", opentsdb.TagSet{"service": "my-service"})
	if v != "" {
		t.Error("Partial match shouldn't work")
	}

	e.UnjoinedOK = true
	for i := 0; i < 100; i++ {
		v, _ = e.Get("pd_service_key", opentsdb.TagSet{"service": "my-service"})
		if v != "MATCH" {
			t.Error("Partial match should work")
		}

		v, _ = e.Get("pd_service_key", opentsdb.TagSet{"service": "my-service"})
		if v != "MATCH" {
			t.Error("Partial match should work")
		}

		v, _ = e.Get("pd_service_key", opentsdb.TagSet{"notag": "tag"})
		if v != "" {
			t.Error("fully unjoined match shouldn't work")
		}
		v, _ = e.Get("pd_service_key", opentsdb.TagSet{"service": "infra/test"})
		if v != "NOMATCH" {
			t.Error("Should match infra/test")
		}
	}
}
