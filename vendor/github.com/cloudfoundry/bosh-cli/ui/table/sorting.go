package table

type Sorting struct {
	SortBy []ColumnSort
	Rows   [][]Value
}

func (s Sorting) Len() int { return len(s.Rows) }

func (s Sorting) Less(i, j int) bool {
	var leftScore, rightScore int

	for ci, cs := range s.SortBy {
		var left, right Value

		left = s.Rows[i][cs.Column].Value()
		right = s.Rows[j][cs.Column].Value()

		c := left.Compare(right)

		if c == 0 {
			leftScore += (10 - ci) * 10
			rightScore += (10 - ci) * 10
		} else {
			if (cs.Asc && c == -1) || (!cs.Asc && c == 1) {
				leftScore += (10 - ci) * 10
			} else {
				rightScore += (10 - ci) * 10
			}
		}
	}

	return leftScore > rightScore
}

func (s Sorting) Swap(i, j int) { s.Rows[i], s.Rows[j] = s.Rows[j], s.Rows[i] }
