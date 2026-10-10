package letterboxd

import (
	"testing"
)

func TestGetRandomReview(t *testing.T) {
	// Test with a well-known movie
	review := GetRandomReview("revolutionary road")

	if review == nil {
		t.Log("No review found - this is okay as it depends on network and availability")
		return
	}

	// If we got a review, check it has basic fields
	if review.Username == "" {
		t.Error("Expected username to be populated")
	}

	t.Logf("Got random review from user: %s", review.Username)
	if review.Rating != "" {
		t.Logf("Rating: %s", review.Rating)
	}
	if review.ReviewText != "" {
		t.Logf("Review (first 100 chars): %.100s", review.ReviewText)
	}
}
