package collectors

import "bosun.org/opentsdb"

type TagOverride struct {
	tags opentsdb.TagSet
}

func (to *TagOverride) AddTags(t opentsdb.TagSet) {
	if to.tags == nil {
		to.tags = t.Copy()
	} else {
		to.tags = to.tags.Merge(t)
	}
}

func (to *TagOverride) ApplyTags(t opentsdb.TagSet) {
	if to.tags != nil {
		t = t.Merge(to.tags)
	}
}
