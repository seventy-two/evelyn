package letterboxd

import (
	"fmt"
	"testing"
)

func TestScrapeReviews(t *testing.T) {
	// Test with a well-known movie
	response, err := ScrapeReviews("revolutionary road")

	if err != nil {
		t.Logf("Warning: Error occurred: %v", err)
		// Don't fail the test as network issues are common
		return
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	fmt.Printf("Movie: %s\n", response.Movie)
	fmt.Printf("Page: %d\n", response.Page)
	fmt.Printf("Reviews found: %d\n", len(response.Reviews))

	for i, review := range response.Reviews {
		if i >= 3 { // Only show first 3 for testing
			break
		}
		fmt.Printf("\n--- Review %d ---\n", i+1)
		fmt.Printf("Username: %s\n", review.Username)
		fmt.Printf("Rating: %s\n", review.Rating)
		fmt.Printf("Date: %s\n", review.DateWatched)
		fmt.Printf("Rewatch: %t\n", review.IsRewatch)
		fmt.Printf("Liked: %t\n", review.IsLiked)
		fmt.Printf("Like Count: %s\n", review.LikeCount)
		fmt.Printf("Review: %.100s...\n", review.ReviewText)
	}
}

func TestCleanMovieTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"The Matrix", "the-matrix"},
		{"Spider-Man: No Way Home", "spider-man-no-way-home"},
		{"Avengers: Endgame", "avengers-endgame"},
		{"The Lord of the Rings: The Fellowship of the Ring", "the-lord-of-the-rings-the-fellowship-of-the-ring"},
		{"2001: A Space Odyssey", "2001-a-space-odyssey"},
		{"Revolutionary Road", "revolutionary-road"},
		{"Face/Off", "face-off"},
		{"Mission: Impossible – The Final Reckoning", "mission-impossible-the-final-reckoning"},
		{"Thunderbolts*", "thunderbolts"},
		{"Guardians of the Galaxy Vol. 3", "guardians-of-the-galaxy-vol-3"},
	}

	for _, test := range tests {
		result := cleanMovieTitle(test.input)
		if result != test.expected {
			t.Errorf("cleanMovieTitle(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}
