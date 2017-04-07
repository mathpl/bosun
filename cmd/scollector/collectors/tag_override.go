package collectors

import (
	"fmt"
	"regexp"
	"strings"

	"bosun.org/opentsdb"
)

type replaceRe struct {
	re  *regexp.Regexp
	dst string
}

type TagOverride struct {
	matchedTags map[string]*regexp.Regexp
	tags        opentsdb.TagSet
	replace     map[string]replaceRe
}

func (to *TagOverride) AddTagOverrides(sources map[string]string, t opentsdb.TagSet, replace map[string][]string) error {
	if to.matchedTags == nil {
		to.matchedTags = make(map[string]*regexp.Regexp)
	}

	if to.replace == nil {
		to.replace = make(map[string]replaceRe)
	}

	var err error
	for tag, re := range sources {
		to.matchedTags[tag], err = regexp.Compile(re)
		if err != nil {
			return fmt.Errorf("invalid regexp: %s error: %s", re, err)
		}
	}

	if to.tags == nil {
		to.tags = t.Copy()
	} else {
		to.tags = to.tags.Merge(t)
	}

	for tag, params := range replace {
		if len(params) != 2 {
			return fmt.Errorf("invalid replace for %s, must be 2 parameters long: %s", tag, strings.Join(params, ","))
		}
		re, err := regexp.Compile(params[0])
		if err != nil {
			return fmt.Errorf("invalid regexp: %s error: %s", re, err)
		}
		to.replace[tag] = replaceRe{re: re, dst: params[1]}
	}

	return nil
}

func (to *TagOverride) ApplyTagOverrides(t opentsdb.TagSet) {
	namedMatchGroup := make(map[string]string)
	for tag, re := range to.matchedTags {
		if v, ok := t[tag]; ok {
			matches := re.FindStringSubmatch(v)

			if len(matches) > 1 {
				for i, match := range matches[1:] {
					matchedTag := re.SubexpNames()[i+1]
					if match != "" && matchedTag != "" {
						namedMatchGroup[matchedTag] = match
					}
				}
			}
		}
	}

	for tag, v := range namedMatchGroup {
		t[tag] = v
	}

	for tag, v := range to.tags {
		if v == "" {
			delete(t, tag)
		} else {
			t[tag] = v
		}
	}

	for tag, replace := range to.replace {
		if v, ok := t[tag]; ok {
			t[tag] = replace.re.ReplaceAllString(v, replace.dst)
		}
	}
}
