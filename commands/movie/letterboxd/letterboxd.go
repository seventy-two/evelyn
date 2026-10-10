package letterboxd

import (
	"math/rand"
)

// GetRandomReview gets a single random review for a movie title
func GetRandomReview(movieTitle string) *Review {
	response, err := ScrapeReviews(movieTitle)
	if err != nil || len(response.Reviews) == 0 {
		return nil
	}

	// Return a random review from the results
	if len(response.Reviews) > 0 {
		randomIndex := rand.Intn(len(response.Reviews))
		return &response.Reviews[randomIndex]
	}

	return nil
}
