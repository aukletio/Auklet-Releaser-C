package main

import (
	"encoding/json"
	"testing"
)

func TestObjectify(t *testing.T) {
	test := []struct {
		b      []byte
		assert func(Object)
	}{
		{
			b:      []byte(`{"signal":11,"stack_trace":[{"fn":42,"cs":56},{"fn":-1987,"cs":32}]}`),
			assert: func(o Object) { _ = o.(*Event) },
		}, {
			b:      []byte(`{"nsamples":11,"callees":[{"fn":42,"cs":56,"ncalls":1,"nsamples":10},{"fn":1987,"cs":32,"ncalls":10,"nsamples":1}]}`),
			assert: func(o Object) { _ = o.(*Profile) },
		}, {
			b: []byte(`{"log":"hi mom"}`),
			assert: func(o Object) {
				if o != nil {
					t.Fail()
				}
			},
		},
	}
	for _, u := range test {
		o, err := Objectify(u.b)
		if err != nil {
			t.Error(err)
		}
		u.assert(o)
		b, err := json.Marshal(o)
		if err != nil {
			t.Error(err)
		}
		t.Log(string(b))
	}
}
