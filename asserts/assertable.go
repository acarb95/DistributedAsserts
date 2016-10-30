package assert

import (
	"sync"
)

type Assertable interface {
    assert() interface{}
    evaluate_results(results []interface{})
}

type AssertableObject struct {
    Name string
    Lock sync.Mutex
    Object interface{}
}

func (ao *AssertableObject) assert() {
    return nil
}

func (ao *AssertableObject) evaluate_results(results []interface{}) {
    // Perform evaluation
}

func create_assertable_object(name string, obj interface{}) {
	return AssertableObject {
		Name: name
		Lock: &sync.Mutex{}
		Object: obj
	}
}