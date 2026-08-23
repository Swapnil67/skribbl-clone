package game

const (
	MaxGuesserScore = 500
	MinGuesserScore = 500
	DrawerBaseBonus = 50
)

// * CalculateGuesserPoints awards higher points for faster guesses
func CalculateGuesserPoints(remainingSeconds, totalDurationSeconds int) int {
	if remainingSeconds <= 0 || totalDurationSeconds <= 0 {
		return MinGuesserScore
	}

	// * Linear score decay: remaining ratio * (Max - Min) + Min
	ratio := float64(remainingSeconds) / float64(totalDurationSeconds)
	score := int(float64(MinGuesserScore) + ratio*float64(MaxGuesserScore-MinGuesserScore))
	if score > MaxGuesserScore {
		return MaxGuesserScore
	}
	return score
}

// * CalculateDrawerBonus awards points to the drawer for each player who guessed
func CalculateDrawerBonus(numGuessedSoFar int) int {
	return numGuessedSoFar * DrawerBaseBonus
}
