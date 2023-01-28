package table

type stack struct {
	v []*field
}

func NewFieldStack() *stack {
	s := &stack{
		v: make([]*field, 0),
	}
	return s
}

func (s *stack) Push(v field) {
	s.v = append(s.v, &v)
}

func (s *stack) Pop() any {
	if len(s.v) == 0 {
		return nil
	}
	v := s.v[len(s.v)-1]
	s.v = s.v[:len(s.v)-1]
	return *v
}

func (s *stack) Len() int {
	return len(s.v)
}
