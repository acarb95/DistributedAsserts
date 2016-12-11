How to use!

Any node that intends to be asserted over must call:

InitDistributedAssert(addr string, neighbours []string, processName string)


Where:

- addr: a free port that this process can recieve on.

- neighbours: a list of the ip:ports chosen by other processes to recieve on.

- processName: the name used for the log files


Then, any variables that the node intends to expose must have the following called:

AddAssertable(name string, pointer interface{}, f processFunction)


Where:

- name: the string associated with the variable

- pointer: a pointer to the variable's address

- f: a function which takes in the pointer, and returns a value. If this is nil, then the function is simply the identity function.


To make an assertion, call the following:

Assert(outerFunc func(map[string]map[string]interface{})bool, requestedValues map[string][]string)


Where:

- requestedValues: this is a map from the ports listed in the neighbours array (from Init) to a list of variable names. The assert will go to each ip:port listed and request the variables in the array.

- outerFunc: this is a function that takes a map from ip:port to a map from variable name to variable value, and returns false if your assertion is violated (if it's in a "bad" state) 