package collectors

import "bosun.org/opentsdb"

type TagOverride struct {
	tags opentsdb.TagSet
}

func (c *TagOverride) AddTags(t opentsdb.TagSet) {
	c.tags = t
}
