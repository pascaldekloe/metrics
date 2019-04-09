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
