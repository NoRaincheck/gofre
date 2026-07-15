package obj

type Counter struct {
	value int64
}

//export GetCounter
func GetCounter() *Counter {
	return &Counter{value: 0}
}

// Increment is not exported via //export
func (c *Counter) Increment() {
	c.value++
}
