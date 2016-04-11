package conf

import (
	"bosun.org/cmd/bosun/search"
	"bosun.org/models"
	"bosun.org/opentsdb"
)

// TODO: remove this and merge it with Lookup
type ExprLookup struct {
	Tags       []string
	Entries    []*ExprEntry
	UnjoinedOK bool
}

type ExprEntry struct {
	AlertKey models.AlertKey
	Values   map[string]string
}

func (lookup *ExprLookup) Get(key string, tag opentsdb.TagSet) (value string, ok bool) {
	for _, entry := range lookup.Entries {
		value, ok = entry.Values[key]
		if !ok {
			continue
		}

		var nbMatch, unjoinedMatch int
		needMatch := len(entry.AlertKey.Group())
		for ak, av := range entry.AlertKey.Group() {
			tagv, found := tag[ak]
			if !found && lookup.UnjoinedOK {
				nbMatch++
				unjoinedMatch++
				continue
			}
			matches, err := search.Match(av, []string{tagv})
			if err != nil {
				return "", false
			}
			if len(matches) == 0 {
				break
			}

			nbMatch++
		}

		// If we're not fully matched keep going
		// If we're only matched through unjoined match keep going as well
		if nbMatch != needMatch || unjoinedMatch == needMatch {
			continue
		}
		return
	}
	return "", false
}
