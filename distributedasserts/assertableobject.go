package distributedasserts

import (
	"sync"
	"net"
	"github.com/arcaneiceman/GoVector/govec"
)

type AssertableObject struct {
    name string
    lock *sync.Mutex
    object interface{}
}

func (a AssertableObject) getObject() interface{} {
	a.lock.Lock()
	// Not sure if this really works, but trying to get a pointer to the object, 
	// but then dereference the pointer so GoLang actually copies the object 
	// instead of passing a pointer. Seems redundant, but that means that
	// copyObj := a.object implicitly does a copy of the struct object. 
	copyObj := *(&a.object)
	a.lock.Unlock()
	return copyObj
}

func (a AssertableObject) setObject(newObj interface{}) {
	a.lock.Lock()
	a.object = newObj
	a.lock.Unlock()
}

func (a AssertableObject) getName() string {
	return a.name
}

func (a AssertableObject) sendObject(writeTo func ([]byte,net.Addr) (int, error), msg string, addr net.Addr, LOG *govec.GoLog) (int, error) {
	a.lock.Lock()
	buf := LOG.PrepareSend(msg, a.object)
	n, err := writeTo(buf, addr)
	a.lock.Unlock()
	return n, err
}

func CreateAssertableObject(name string, obj interface{}) AssertableObject{
	lock := &sync.Mutex{}
	aObj := AssertableObject{ name: name, lock: lock, object: obj}
	return aObj
}