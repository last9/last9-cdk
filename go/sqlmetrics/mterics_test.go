package sqlmetrics

import (
	dto "github.com/prometheus/client_model/go"
)

// labelSetContains scans for labels and a map of expected values;
// returning true IF all of the expected keys and alues are present in labels.
// returning false for all other cases.
func labelSetContains(
	labels []*dto.LabelPair, expected map[string]string,
) bool {
	for k, v := range expected {
		var found bool
		for _, l := range labels {
			if l.GetName() == k && l.GetValue() == v {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
