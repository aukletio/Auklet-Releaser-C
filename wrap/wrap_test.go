package main

import (
	"encoding/json"
	"syscall"
	"testing"
)

func wait() syscall.WaitStatus {
	return syscall.WaitStatus(0)
}

func TestObjectify(t *testing.T) {
	send := func(o object) error {
		b, err := json.Marshal(o)
		if err != nil {
			t.Log(string(b))
			return err
		}
		return nil
	}

	test := []struct {
		b      []byte
		assert func(object)
	}{
		{
			b: []byte(`{
					"type":"profile",
					"data":{
						"tree":{
							"nsamples":11,
							"callees":[{
								"fn":42,
								"cs":56,
								"ncalls":1,
								"nsamples":10
							},{
								"fn":1987,
								"cs":32,
								"ncalls":10,
								"nsamples":1
							}]
						}
					}
				}`),
			assert: func(o object) { _ = o.(*profile) },
		},
		{
			b: []byte(`{
					"type":"log",
					"data":{
						"level":"info",
						"message":"hi mom"
					}
				}`),
			assert: func(o object) {
				if o != nil {
					t.Fail()
				}
			},
		},
		{
			b: []byte(`{
					"type":"event",
					"data":{
						"signal":11,
						"stack_trace":[{
							"fn":42,
							"cs":56
						},{
							"fn":-1987,
							"cs":32
						}]
					}
				}`),
			assert: func(o object) { _ = o.(*event) },
		},
	}
	var done bool
	for _, u := range test {
		var err error
		done, err = objectify(u.b, wait, send)
		if err != nil {
			t.Error(err)
		}
		if done {
			break
		}
	}
	if !done {
		t.Error("processed an event, but not done")
	}
}
