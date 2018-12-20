package statistics

type Statistics struct {
}

type Counts struct {
	A int
	B int
}

func Run() (stats *Statistics) {
	return &Statistics{}
}

func (s *Statistics) GetCounts() Counts {
	return Counts{
		A: 1,
		B: 42,
	}
}
