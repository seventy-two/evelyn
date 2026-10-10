package letterboxd

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
)

// Review represents a single Letterboxd review
type Review struct {
	Username    string `json:"username"`
	Rating      string `json:"rating"`
	ReviewText  string `json:"review_text"`
	DateWatched string `json:"date_watched"`
	IsRewatch   bool   `json:"is_rewatch"`
	IsLiked     bool   `json:"is_liked"`
	LikeCount   string `json:"like_count"`
}

// ReviewsResponse represents the response containing reviews
type ReviewsResponse struct {
	Reviews []Review `json:"reviews"`
	Page    int      `json:"page"`
	Movie   string   `json:"movie"`
}

// ScrapeReviews scrapes reviews for a given movie title from Letterboxd
// It tries random pages 1-200, decreasing the range on each failed attempt
// Returns reviews when found, or falls back to page 1 after 5 attempts
func ScrapeReviews(movieTitle string) (*ReviewsResponse, error) {
	baseURL := "https://letterboxd.com/film/%s/reviews/by/added/page/%d/"

	// Clean movie title for URL (replace spaces with hyphens, lowercase, remove special chars)
	cleanTitle := cleanMovieTitle(movieTitle)

	maxAttempts := 5
	pageRange := 100

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Generate random page number within current range
		var page int
		if attempt == maxAttempts {
			// Last attempt: use page 1
			page = 1
		} else {
			page = rand.Intn(pageRange) + 1
		}

		url := fmt.Sprintf(baseURL, cleanTitle, page)
		fmt.Println("Fetching URL:", url)
		// Make HTTP request
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error fetching URL %s: %v\n", url, err)
			continue
		}
		if page == 1 {
			fmt.Println(resp)
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			// Decrease page range for next attempt
			pageRange = pageRange * 2 / 4
			continue
		}

		// Read the HTML
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			fmt.Printf("Error reading response body from %s: %v\n", url, err)
			continue
		}

		// Extract reviews
		reviews := extractReviews(string(body))

		// If we found reviews, return them
		if len(reviews) > 0 {
			return &ReviewsResponse{
				Reviews: reviews,
				Page:    page,
				Movie:   movieTitle,
			}, nil
		}

		// If no reviews found and not the last attempt, decrease range and try again
		if attempt < maxAttempts {
			pageRange = pageRange * 3 / 4
			if pageRange < 1 {
				pageRange = 1
			}
		}
	}
	fmt.Println("All attempts to scrape reviews failed.")
	// If we get here, we couldn't find any reviews
	return &ReviewsResponse{
		Reviews: []Review{},
		Page:    1,
		Movie:   movieTitle,
	}, fmt.Errorf("no reviews found for movie: %s", movieTitle)
}

// extractReviews parses the HTML string and extracts review data using regex
func extractReviews(html string) []Review {
	var reviews []Review

	// Find all review containers using regex
	reviewPattern := `(?s)<article class="production-viewing -viewing"[^>]*>(.*?)</article>`
	reviewRegex := regexp.MustCompile(reviewPattern)
	reviewMatches := reviewRegex.FindAllStringSubmatch(html, -1)

	for _, match := range reviewMatches {
		if len(match) < 2 {
			continue
		}

		reviewHTML := match[1]
		review := Review{}

		// Extract username from avatar link
		usernamePattern := `<a class="avatar[^"]*" href="/([^/]+)/"`
		usernameRegex := regexp.MustCompile(usernamePattern)
		if usernameMatch := usernameRegex.FindStringSubmatch(reviewHTML); len(usernameMatch) > 1 {
			review.Username = usernameMatch[1]
		}

		// Extract rating from span class
		ratingPattern := `<span class="[^"]*rating[^"]*rated-(\d+)[^"]*"`
		ratingRegex := regexp.MustCompile(ratingPattern)
		if ratingMatch := ratingRegex.FindStringSubmatch(reviewHTML); len(ratingMatch) > 1 {
			ratingNum := ratingMatch[1]
			// Convert numeric rating to stars (assuming 1-10 scale to 0.5-5 stars)
			switch ratingNum {
			case "2":
				review.Rating = "★"
			case "4":
				review.Rating = "★★"
			case "5":
				review.Rating = "★★½"
			case "6":
				review.Rating = "★★★"
			case "7":
				review.Rating = "★★★½"
			case "8":
				review.Rating = "★★★★"
			case "9":
				review.Rating = "★★★★½"
			case "10":
				review.Rating = "★★★★★"
			default:
				review.Rating = "★½"
			}
		}

		// Extract review text from the review body
		reviewTextPattern := `(?s)<div class="body-text[^"]*js-review-body[^"]*"[^>]*>(.*?)</div>`
		reviewTextRegex := regexp.MustCompile(reviewTextPattern)
		if reviewTextMatch := reviewTextRegex.FindStringSubmatch(reviewHTML); len(reviewTextMatch) > 1 {
			reviewContent := reviewTextMatch[1]

			// Check for collapsed text first
			collapsedPattern := `(?s)<div class="collapsed-text[^"]*"[^>]*>(.*?)</div>`
			collapsedRegex := regexp.MustCompile(collapsedPattern)
			if collapsedMatch := collapsedRegex.FindStringSubmatch(reviewContent); len(collapsedMatch) > 1 {
				review.ReviewText = cleanText(collapsedMatch[1])
			} else {
				review.ReviewText = cleanText(reviewContent)
			}
		}

		// Extract date watched
		datePattern := `<time class="timestamp"[^>]*datetime="([^"]+)"`
		dateRegex := regexp.MustCompile(datePattern)
		if dateMatch := dateRegex.FindStringSubmatch(reviewHTML); len(dateMatch) > 1 {
			review.DateWatched = dateMatch[1]
		}

		// Check if it's a rewatch
		review.IsRewatch = strings.Contains(reviewHTML, "Rewatched by")

		// Check if the review is liked
		review.IsLiked = strings.Contains(reviewHTML, `icon-liked`)

		// Extract like count
		likeCountPattern := `data-count="([^"]+)"`
		likeCountRegex := regexp.MustCompile(likeCountPattern)
		if likeMatch := likeCountRegex.FindStringSubmatch(reviewHTML); len(likeMatch) > 1 {
			review.LikeCount = likeMatch[1]
		}

		// Only add review if we have at least a username
		if review.Username != "" {
			reviews = append(reviews, review)
		}
	}

	return reviews
}

// cleanMovieTitle cleans a movie title for use in Letterboxd URLs
func cleanMovieTitle(title string) string {
	// Convert to lowercase
	cleaned := strings.ToLower(title)

	// Replace slashes and various dash types with hyphens
	cleaned = strings.ReplaceAll(cleaned, "/", "-")
	cleaned = strings.ReplaceAll(cleaned, "–", "-") // en dash
	cleaned = strings.ReplaceAll(cleaned, "—", "-") // em dash
	cleaned = strings.ReplaceAll(cleaned, ":", "-") // colon
	cleaned = strings.ReplaceAll(cleaned, ".", "-") // period

	// Remove common punctuation and special characters (except hyphens)
	specialChars := regexp.MustCompile(`[^\w\s-]`)
	cleaned = specialChars.ReplaceAllString(cleaned, "")

	// Replace spaces and multiple hyphens with single hyphens
	spaceRegex := regexp.MustCompile(`[\s_]+`)
	cleaned = spaceRegex.ReplaceAllString(cleaned, "-")

	// Remove multiple consecutive hyphens
	hyphenRegex := regexp.MustCompile(`-+`)
	cleaned = hyphenRegex.ReplaceAllString(cleaned, "-")

	// Trim leading/trailing hyphens
	cleaned = strings.Trim(cleaned, "-")

	return cleaned
}

// cleanText removes HTML tags and cleans up text content
func cleanText(text string) string {
	// Remove HTML tags
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	cleaned := tagRegex.ReplaceAllString(text, "")

	// Clean up whitespace
	cleaned = strings.ReplaceAll(cleaned, "\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "\t", " ")

	// Replace multiple spaces with single space
	spaceRegex := regexp.MustCompile(`\s+`)
	cleaned = spaceRegex.ReplaceAllString(cleaned, " ")

	return strings.TrimSpace(cleaned)
}
