package sqlmetrics

// A type that user can extend to Parser a query and extract meaningful
// metrics out of it.
type LabelMaker func(string) map[string]string

// The default labelSet to be exported is just a query, that too trimmed down
// to 140 charachters only. Queries can be large and can really bring down
// the metric server to its knees if left untapped. If this behaviour is not
// desired, a user can anwyay implement their own QToLabelSet and emit the
// raw query as-it-is.
func defaultLabelMaker(q string) map[string]string {
	if len(q) > 140 {
		q = q[:140] + "..."
	}

	return map[string]string{"per": q}
}
