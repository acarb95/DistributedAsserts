package asserts

import (
	"sync"
)

type Assertable interface {
    assert() interface{}
    evaluate_results(results []interface{})
}

type AssertableObject struct {
    Name string
    Lock *sync.Mutex
    Object interface{}
}

func (ao *AssertableObject) assert() interface{} {
    return nil
}

func (ao *AssertableObject) evaluate_results(results []interface{}) {
    // Perform evaluation
}

func CreateAssertableObject(name string, obj interface{}) AssertableObject{
	lock := &sync.Mutex{}
	aObj := AssertableObject{ Name: name, Lock: lock, Object: obj}
	return aObj
}