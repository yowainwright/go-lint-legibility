package p

func total(scores []int) int {
	sum := 0
	for i := 0; i < len(scores); i++ {
		sum += scores[i]
	}

	return sum
}
