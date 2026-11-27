package scenario

type Entry struct {
    Value int
}

type SharedMap struct {
    Data map[int]*Entry
}

func (m *SharedMap) Write(key, value int) {
    m.Data[key].Value = value
}

func (m *SharedMap) Read(key int) int {
    return m.Data[key].Value
}
