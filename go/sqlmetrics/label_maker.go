package sqlmetrics

type LabelSet map[string]string

// Merge accepts a set of LabelSets and merge them into this LabelSet.
func (l LabelSet) Merge(ls ...LabelSet) LabelSet {
	for _, m := range ls {
		for k, v := range m {
			if v != "" {
				l[k] = v
			}
		}
	}

	return l
}

// ToMap converts a LabelSet to map[string]string
func (l LabelSet) ToMap() map[string]string {
	return map[string]string(l)
}

// A type that user can extend to Parse a query and extract less verbose
// or more relevant labels out of it.
type LabelMaker func(string) LabelSet

const idealLabelLen = 20

// The default labelSet to be exported is just a query, that too trimmed down
// to 140 charachters only. Queries can be large and can really bring down
// the metric server to its knees if left untapped. If this behaviour is not
// desired, a user can anwyay implement their own QToLabelSet and emit the
// raw query as-it-is.
func defaultLabelMaker(q string) LabelSet {
	if len(q) > idealLabelLen {
		q = q[:idealLabelLen] + "..."
	}

	return LabelSet{"per": q}
}

// queryStatus is an enumeration.
type queryStatus int

const (
	success queryStatus = iota
	failure
)

// String method on queryStatus comes handy when printing or
// creating labels.
func (q queryStatus) String() string {
	return [...]string{"success", "failure"}[q]
}
