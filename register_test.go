package metrics

import "testing"

func TestRegisterLabelClash(t *testing.T) {
	var labels = []string{"first", "another"}

	// test any order of the labels
	for _, label1 := range labels {
		for _, label2 := range labels {
			if label1 == label2 {
				// undefined in the specification
				continue
			}

			f := NewRegister().Must2LabelCounter
			f("test_metric", labels[0], labels[1])
			func() {
				defer func() {
					if recover() == nil {
						t.Errorf("no panic for reuse with (%q, %q)", label1, label2)
					}
				}()
				f("test_metric", label1, label2)
			}()
		}
	}

	labels = append(labels, "last")
	for _, label1 := range labels {
		for _, label2 := range labels {
			for _, label3 := range labels {
				if label1 == label2 || label1 == label3 || label2 == label3 {
					// undefined in the specification
					continue
				}

				f := NewRegister().Must3LabelCounter
				f("test_metric", labels[0], labels[1], labels[2])
				func() {
					defer func() {
						if recover() == nil {
							t.Errorf("no panic for reuse with (%q, %q, %q)", label1, label2, label3)
						}
					}()
					f("test_metric", label1, label2, label3)
				}()
			}
		}
	}
}

func TestSort3(t *testing.T) {
	golden := []struct {
		S1, S2, S3          string
		Order               int
		Want1, Want2, Want3 string
	}{
		{"a", "b", "c", order123, "a", "b", "c"},
		{"a", "c", "b", order132, "a", "b", "c"},
		{"b", "a", "c", order213, "a", "b", "c"},
		{"b", "c", "a", order231, "a", "b", "c"},
		{"c", "a", "b", order312, "a", "b", "c"},
		{"c", "b", "a", order321, "a", "b", "c"},
	}

	for _, gold := range golden {
		s1, s2, s3 := gold.S1, gold.S2, gold.S3
		order := sort3(&s1, &s2, &s3)
		if order != gold.Order {
			t.Errorf("(%q, %q, %q) got order %d, want %d",
				gold.S1, gold.S2, gold.S3, order, gold.Order)
		}
		if s1 != gold.Want1 || s2 != gold.Want2 || s3 != gold.Want3 {
			t.Errorf("(%q, %q, %q) got (%q, %q, %q), want (%q, %q, %q)",
				gold.S1, gold.S2, gold.S3,
				s1, s2, s3, gold.Want1, gold.Want2, gold.Want3)
		}
	}
}
