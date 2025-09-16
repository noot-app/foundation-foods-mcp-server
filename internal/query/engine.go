package query

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
)

// Engine implements the QueryEngine interface for Foundation Foods data
type Engine struct {
	data   *FoundationFoodsData
	logger *slog.Logger
}

// NewEngine creates a new query engine and loads the Foundation Foods data
func NewEngine(jsonFilePath string, logger *slog.Logger) (*Engine, error) {
	logger.Info("Loading Foundation Foods data", "path", jsonFilePath)

	// Read the JSON file
	data, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Foundation Foods data file: %w", err)
	}

	// Parse the JSON
	var foundationFoodsData FoundationFoodsData
	if err := json.Unmarshal(data, &foundationFoodsData); err != nil {
		return nil, fmt.Errorf("failed to parse Foundation Foods JSON data: %w", err)
	}

	logger.Info("Foundation Foods data loaded successfully",
		"food_count", len(foundationFoodsData.FoundationFoods))

	return &Engine{
		data:   &foundationFoodsData,
		logger: logger,
	}, nil
}

// SearchFoodsByName searches for foods by their description using intelligent scoring
func (e *Engine) SearchFoodsByName(ctx context.Context, query string, limit int) ([]FoundationFood, error) {
	if e.data == nil {
		return nil, fmt.Errorf("foundation Foods data not loaded")
	}

	if limit <= 0 {
		limit = 3
	}
	if limit > 10 {
		limit = 10
	}

	e.logger.Debug("Searching Foundation Foods",
		"query", query,
		"limit", limit,
		"total_foods", len(e.data.FoundationFoods))

	// Normalize the search query
	normalizedQuery := normalizeString(query)
	queryWords := strings.Fields(normalizedQuery)

	var results []SearchResult

	// Search through all foods
	for _, food := range e.data.FoundationFoods {
		score := calculateRelevanceScore(food.Description, normalizedQuery, queryWords)
		if score > 0 {
			results = append(results, SearchResult{
				Food:  food,
				Score: score,
			})
		}
	}

	// Sort by score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Extract top results
	var foods []FoundationFood
	for i, result := range results {
		if i >= limit {
			break
		}
		foods = append(foods, result.Food)

		e.logger.Debug("Search result",
			"rank", i+1,
			"score", result.Score,
			"description", result.Food.Description)
	}

	e.logger.Debug("Search complete",
		"query", query,
		"results_found", len(results),
		"results_returned", len(foods))

	return foods, nil
}

// GetFoodByFdcId retrieves a specific food by its FDC ID
func (e *Engine) GetFoodByFdcId(ctx context.Context, fdcId int) (*FoundationFood, error) {
	if e.data == nil {
		return nil, fmt.Errorf("foundation Foods data not loaded")
	}

	for _, food := range e.data.FoundationFoods {
		if food.FdcId == fdcId {
			return &food, nil
		}
	}

	return nil, fmt.Errorf("food with FDC ID %d not found", fdcId)
}

// Health checks if the query engine is ready and operational
func (e *Engine) Health(ctx context.Context) error {
	if e.data == nil {
		return fmt.Errorf("foundation Foods data not loaded")
	}

	if len(e.data.FoundationFoods) == 0 {
		return fmt.Errorf("foundation Foods data is empty")
	}

	return nil
}

// SearchFoodsByNameSimplified searches for foods and returns simplified nutrient information
func (e *Engine) SearchFoodsByNameSimplified(ctx context.Context, query string, limit int, nutrientsToInclude []string) (*SimplifiedNutrientResponse, error) {
	// Use the existing search functionality
	foods, err := e.SearchFoodsByName(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	// Convert to simplified format
	simplifiedFoods := make([]SimplifiedFood, 0, len(foods))
	for _, food := range foods {
		simplifiedFood := SimplifiedFood{
			Name:         food.Description,
			Nutrients:    make([]SimplifiedNutrient, 0, len(food.FoodNutrients)),
			FoodPortions: make([]SimplifiedFoodPortion, 0, len(food.FoodPortions)),
		}

		// Convert nutrients to simplified format with filtering
		for _, nutrient := range food.FoodNutrients {
			// Skip Energy in kJ - we only want kcal
			if strings.ToLower(strings.TrimSpace(nutrient.Nutrient.Name)) == "energy" &&
				strings.ToLower(strings.TrimSpace(nutrient.Nutrient.UnitName)) == "kj" {
				continue
			}

			// Check if this nutrient should be included
			if e.shouldIncludeNutrient(nutrient.Nutrient.Name, nutrientsToInclude) {
				simplifiedNutrient := SimplifiedNutrient{
					Name:       nutrient.Nutrient.Name,
					UnitName:   nutrient.Nutrient.UnitName,
					Amount:     nutrient.Amount,
					DataPoints: nutrient.DataPoints,
					Max:        nutrient.Max,
					Min:        nutrient.Min,
					Median:     nutrient.Median,
				}
				simplifiedFood.Nutrients = append(simplifiedFood.Nutrients, simplifiedNutrient)
			}
		}

		// Convert food portions to simplified format
		for _, portion := range food.FoodPortions {
			simplifiedPortion := SimplifiedFoodPortion{
				Value: portion.Value,
				MeasureUnit: SimplifiedMeasureUnit{
					Name:         portion.MeasureUnit.Name,
					Abbreviation: portion.MeasureUnit.Abbreviation,
				},
				GramWeight: portion.GramWeight,
				Amount:     portion.Amount,
			}
			simplifiedFood.FoodPortions = append(simplifiedFood.FoodPortions, simplifiedPortion)
		}

		simplifiedFoods = append(simplifiedFoods, simplifiedFood)
	}

	return &SimplifiedNutrientResponse{
		Found: len(simplifiedFoods) > 0,
		Count: len(simplifiedFoods),
		Foods: simplifiedFoods,
	}, nil
}

// normalizeString normalizes a string for better searching
func normalizeString(s string) string {
	// Convert to lowercase and trim whitespace
	s = strings.ToLower(strings.TrimSpace(s))

	// Remove common punctuation that doesn't affect meaning
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")

	return s
}

// calculateRelevanceScore calculates how relevant a food description is to a search query
func calculateRelevanceScore(description, normalizedQuery string, queryWords []string) float64 {
	normalizedDesc := normalizeString(description)
	descWords := strings.Fields(normalizedDesc)

	// No match if no words to compare
	if len(queryWords) == 0 || len(descWords) == 0 {
		return 0
	}

	var score float64

	// 1. Exact match (highest priority)
	if normalizedDesc == normalizedQuery {
		score += 1000
	}

	// 2. Query appears as substring at the beginning of description
	if strings.HasPrefix(normalizedDesc, normalizedQuery) {
		score += 500
	}

	// 3. Query appears as substring anywhere
	if strings.Contains(normalizedDesc, normalizedQuery) {
		score += 100
	}

	// 4. Word-level matching
	matchedWords := 0
	totalQueryWords := len(queryWords)

	for _, queryWord := range queryWords {
		bestWordScore := 0.0

		for i, descWord := range descWords {
			wordScore := 0.0

			// Exact word match
			if descWord == queryWord {
				wordScore = 50
				// Bonus for position (earlier words are more important)
				if i < 3 {
					wordScore += float64(3-i) * 10
				}
			} else if strings.HasPrefix(descWord, queryWord) && len(queryWord) >= 3 {
				// Prefix match (for partial words)
				wordScore = 25
				if i < 3 {
					wordScore += float64(3-i) * 5
				}
			} else if strings.Contains(descWord, queryWord) && len(queryWord) >= 4 {
				// Substring match (less reliable)
				wordScore = 10
			}

			if wordScore > bestWordScore {
				bestWordScore = wordScore
			}
		}

		if bestWordScore > 0 {
			matchedWords++
			score += bestWordScore
		}
	}

	// 5. Bonus for matching multiple words
	if totalQueryWords > 1 {
		matchRatio := float64(matchedWords) / float64(totalQueryWords)
		score *= (1 + matchRatio) // Boost score based on word match ratio
	}

	// 6. Penalty for very long descriptions that match incidentally
	if len(descWords) > 10 && matchedWords < totalQueryWords {
		score *= 0.8
	}

	// 7. Specific food search improvements
	score = adjustScoreForFoodContext(description, normalizedQuery, queryWords, score)

	return score
}

// adjustScoreForFoodContext applies food-specific scoring adjustments
func adjustScoreForFoodContext(description, normalizedQuery string, queryWords []string, currentScore float64) float64 {
	normalizedDesc := normalizeString(description)

	// Boost simple, direct food names
	descWords := strings.Fields(normalizedDesc)
	if len(descWords) <= 3 && len(queryWords) == 1 {
		// Simple food names like "milk" or "eggs" should rank higher
		if strings.Contains(descWords[0], queryWords[0]) {
			currentScore *= 1.5
		}
	}

	// Handle common food search patterns
	for _, queryWord := range queryWords {
		switch queryWord {
		case "milk":
			// Prefer "milk, whole" over "cheese, cottage, lowfat, 2% milkfat"
			if strings.HasPrefix(normalizedDesc, "milk") {
				currentScore *= 2.0
			} else if strings.Contains(normalizedDesc, "milkfat") || strings.Contains(normalizedDesc, "milk fat") {
				currentScore *= 0.3 // Reduce score for incidental mentions
			}
		case "cheese":
			if strings.HasPrefix(normalizedDesc, "cheese") {
				currentScore *= 1.5
			}
		case "chicken", "beef", "pork":
			if strings.HasPrefix(normalizedDesc, queryWord) {
				currentScore *= 1.3
			}
		case "bread":
			if strings.HasPrefix(normalizedDesc, "bread") || strings.Contains(normalizedDesc, "bread") {
				currentScore *= 1.2
			}
		}
	}

	// Penalize very specific branded or technical descriptions when searching for generic terms
	if len(queryWords) == 1 && len(descWords) > 6 {
		// Check if description contains lots of brand names, codes, or technical terms
		brandIndicators := []string{"brand", "store", "composite", "mixed", "frozen", "canned"}
		for _, indicator := range brandIndicators {
			if strings.Contains(normalizedDesc, indicator) {
				currentScore *= 0.7
				break
			}
		}
	}

	return currentScore
}

// shouldIncludeNutrient checks if a nutrient should be included based on the filter list
func (e *Engine) shouldIncludeNutrient(nutrientName string, nutrientsToInclude []string) bool {
	// If no filter is specified, include all nutrients
	if len(nutrientsToInclude) == 0 {
		return true
	}

	normalizedNutrientName := strings.ToLower(strings.TrimSpace(nutrientName))

	for _, includeName := range nutrientsToInclude {
		normalizedIncludeName := strings.ToLower(strings.TrimSpace(includeName))

		// Direct exact match
		if normalizedNutrientName == normalizedIncludeName {
			return true
		}

		// Enhanced matching for alternative names
		if e.isAlternativeNutrientName(normalizedNutrientName, normalizedIncludeName) {
			return true
		}
	}

	return false
}

// isAlternativeNutrientName checks if two nutrient names refer to the same nutrient
func (e *Engine) isAlternativeNutrientName(dataName, filterName string) bool {
	// Handle legacy fatty acid naming - check if filter name without PUFA prefix matches data name with PUFA prefix
	if strings.HasPrefix(filterName, "pufa ") {
		withoutPrefix := strings.TrimPrefix(filterName, "pufa ")
		if dataName == withoutPrefix {
			return true
		}
	}

	// Handle reverse case - data has PUFA prefix but filter doesn't
	if strings.HasPrefix(dataName, "pufa ") {
		withoutPrefix := strings.TrimPrefix(dataName, "pufa ")
		if filterName == withoutPrefix {
			return true
		}
	}

	// Note: Sugar variants are treated as separate nutrients - no alternative mapping

	// Handle vitamin C variations
	if (filterName == "vitamin c, total ascorbic acid" && dataName == "vitamin c") ||
		(filterName == "vitamin c" && dataName == "vitamin c, total ascorbic acid") {
		return true
	}

	return false
}
