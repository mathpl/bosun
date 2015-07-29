package collectors

import (
	"log"

	"bosun.org/opentsdb"
)

type TagOverride struct {
	tags opentsdb.TagSet
}

func (to *TagOverride) AddTags(t opentsdb.TagSet) {
	if to.tags == nil {
		to.tags = t
	} else {
		log.Printf("Adding %+v", t)
	}
}

func (to *TagOverride) ApplyTags(t opentsdb.TagSet) {
	if to.tags != nil {
		t.Merge(to.tags)
	}
}
