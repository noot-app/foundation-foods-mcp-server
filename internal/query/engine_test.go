package query

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/noot-app/foundation-foods-mcp-server/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEngine(t *testing.T) {
	t.Run("creates engine successfully with valid data file", func(t *testing.T) {
		// This test would need the actual data file to run
		t.Skip("Skipping integration test - requires actual data file")

		logger := config.NewTestLogger(io.Discard, "debug")

		engine, err := NewEngine("../../data/foundationfoods_2025-04-24.json", logger)

		require.NoError(t, err)
		assert.NotNil(t, engine)
		assert.NotNil(t, engine.data)
		assert.Greater(t, len(engine.data.FoundationFoods), 0)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		logger := config.NewTestLogger(io.Discard, "debug")

		engine, err := NewEngine("non-existent-file.json", logger)

		assert.Error(t, err)
		assert.Nil(t, engine)
		assert.Contains(t, err.Error(), "failed to read Foundation Foods data file")
	})
}

func TestEngine_SearchFoodsByName(t *testing.T) {
	// Create a test engine with mock data
	testData := &FoundationFoodsData{
		FoundationFoods: []FoundationFood{
			{
				Description:  "Milk, whole, 3.25% milkfat",
				FdcId:        1,
				FoodCategory: FoodCategory{Description: "Dairy and Egg Products"},
			},
			{
				Description:  "Cheese, cottage, lowfat, 2% milkfat",
				FdcId:        2,
				FoodCategory: FoodCategory{Description: "Dairy and Egg Products"},
			},
			{
				Description:  "Eggs, whole, raw, fresh",
				FdcId:        3,
				FoodCategory: FoodCategory{Description: "Dairy and Egg Products"},
			},
			{
				Description:  "Bread, white, commercially prepared",
				FdcId:        4,
				FoodCategory: FoodCategory{Description: "Baked Products"},
			},
		},
	}

	logger := config.NewTestLogger(io.Discard, "debug")
	engine := &Engine{
		data:   testData,
		logger: logger,
	}

	ctx := context.Background()

	t.Run("finds milk matches with proper prioritization", func(t *testing.T) {
		results, err := engine.SearchFoodsByName(ctx, "milk", 3)

		require.NoError(t, err)
		assert.Greater(t, len(results), 0)
		// First result should be the one that starts with "Milk"
		assert.Equal(t, "Milk, whole, 3.25% milkfat", results[0].Description)
	})

	t.Run("finds partial matches", func(t *testing.T) {
		results, err := engine.SearchFoodsByName(ctx, "egg", 3)

		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Eggs, whole, raw, fresh", results[0].Description)
	})

	t.Run("prioritizes better matches", func(t *testing.T) {
		// "milk" should find "Milk, whole..." before "Cheese, cottage... milkfat"
		results, err := engine.SearchFoodsByName(ctx, "milk", 2)

		require.NoError(t, err)
		assert.Greater(t, len(results), 0)
		assert.Equal(t, "Milk, whole, 3.25% milkfat", results[0].Description)
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		results, err := engine.SearchFoodsByName(ctx, "a", 2) // Should match multiple items

		require.NoError(t, err)
		assert.LessOrEqual(t, len(results), 2)
	})

	t.Run("handles case insensitive search", func(t *testing.T) {
		results, err := engine.SearchFoodsByName(ctx, "MILK", 3)

		require.NoError(t, err)
		assert.Greater(t, len(results), 0)
		assert.Equal(t, "Milk, whole, 3.25% milkfat", results[0].Description)
	})

	t.Run("returns empty for no matches", func(t *testing.T) {
		results, err := engine.SearchFoodsByName(ctx, "xyz123nonexistent", 3)

		require.NoError(t, err)
		assert.Len(t, results, 0)
	})
}

func TestEngine_GetFoodByFdcId(t *testing.T) {
	testData := &FoundationFoodsData{
		FoundationFoods: []FoundationFood{
			{
				Description:  "Milk, whole, 3.25% milkfat",
				FdcId:        12345,
				FoodCategory: FoodCategory{Description: "Dairy and Egg Products"},
			},
		},
	}

	logger := config.NewTestLogger(io.Discard, "debug")
	engine := &Engine{
		data:   testData,
		logger: logger,
	}

	ctx := context.Background()

	t.Run("finds food by FDC ID", func(t *testing.T) {
		food, err := engine.GetFoodByFdcId(ctx, 12345)

		require.NoError(t, err)
		assert.NotNil(t, food)
		assert.Equal(t, "Milk, whole, 3.25% milkfat", food.Description)
		assert.Equal(t, 12345, food.FdcId)
	})

	t.Run("returns error for non-existent FDC ID", func(t *testing.T) {
		food, err := engine.GetFoodByFdcId(ctx, 99999)

		assert.Error(t, err)
		assert.Nil(t, food)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestEngine_Health(t *testing.T) {
	logger := config.NewTestLogger(io.Discard, "debug")
	ctx := context.Background()

	t.Run("healthy when data is loaded", func(t *testing.T) {
		testData := &FoundationFoodsData{
			FoundationFoods: []FoundationFood{
				{Description: "Test food", FdcId: 1},
			},
		}

		engine := &Engine{
			data:   testData,
			logger: logger,
		}

		err := engine.Health(ctx)
		assert.NoError(t, err)
	})

	t.Run("unhealthy when data is not loaded", func(t *testing.T) {
		engine := &Engine{
			data:   nil,
			logger: logger,
		}

		err := engine.Health(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not loaded")
	})

	t.Run("unhealthy when data is empty", func(t *testing.T) {
		testData := &FoundationFoodsData{
			FoundationFoods: []FoundationFood{},
		}

		engine := &Engine{
			data:   testData,
			logger: logger,
		}

		err := engine.Health(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})
}

func TestCalculateRelevanceScore(t *testing.T) {
	testCases := []struct {
		name          string
		description   string
		query         string
		expectGreater float64 // Should be greater than this score
	}{
		{
			name:          "exact match gets highest score",
			description:   "Milk, whole",
			query:         "milk, whole",
			expectGreater: 900,
		},
		{
			name:          "prefix match gets high score",
			description:   "Milk, whole, 3.25% milkfat",
			query:         "milk",
			expectGreater: 400,
		},
		{
			name:          "substring match gets moderate score",
			description:   "Cheese, cottage, lowfat, 2% milkfat",
			query:         "milk",
			expectGreater: 30, // Reduced because food-specific adjustments reduce score for incidental matches
		},
		{
			name:          "no match gets zero score",
			description:   "Bread, white, sliced",
			query:         "xyz",
			expectGreater: -1, // Should be 0
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			normalizedQuery := normalizeString(tc.query)
			queryWords := strings.Fields(normalizedQuery)

			score := calculateRelevanceScore(tc.description, normalizedQuery, queryWords)

			if tc.expectGreater == -1 {
				assert.Equal(t, 0.0, score)
			} else {
				assert.Greater(t, score, tc.expectGreater,
					"Score %.2f should be greater than %.2f for description '%s' and query '%s'",
					score, tc.expectGreater, tc.description, tc.query)
			}
		})
	}
}

func TestNormalizeString(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"Milk, Whole", "milk whole"},
		{"Bread (White)", "bread white"},
		{"Cheese, cottage, lowfat, 2% milkfat.", "cheese cottage lowfat 2% milkfat"},
		{"  EGGS  ", "eggs"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := normalizeString(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
