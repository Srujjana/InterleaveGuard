package scenario

type SharedSlice struct {
    Values []int
}

func (s *SharedSlice) Append(value int) {
    s.Values = append(s.Values, value)
}
