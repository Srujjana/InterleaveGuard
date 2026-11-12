package scenario

//shared var :c.value
/*it is accessed concurrently by multiple goroutines
without synchronization, leading to a data race.*/
type Counter struct {
	Value int
}

func (c *Counter) Increment() {
	c.Value++
}

func (c *Counter) Read() int {
	return c.Value
}
