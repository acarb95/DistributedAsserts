package asserts

import (
	"sync"
)

type Assertable struct {
    NodesVariables map[string][]string
    Evaluate func(variables ... map[string]interface{}) bool
}