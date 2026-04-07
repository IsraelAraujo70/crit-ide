package highlight

// sortTokens sorts tokens by start offset using insertion sort (small slices).
func sortTokens(tokens []Token) {
	for i := 1; i < len(tokens); i++ {
		key := tokens[i]
		j := i - 1
		for j >= 0 && tokens[j].Start > key.Start {
			tokens[j+1] = tokens[j]
			j--
		}
		tokens[j+1] = key
	}
}
