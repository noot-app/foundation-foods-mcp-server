package mcpgo

import (
	"context"
	"io"
	"testing"

	"github.com/noot-app/foundation-foods-mcp-server/internal/auth"
	"github.com/noot-app/foundation-foods-mcp-server/internal/config"
	"github.com/noot-app/foundation-foods-mcp-server/internal/query"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	t.Run("creates server successfully", func(t *testing.T) {
		logger := config.NewTestLogger(io.Discard, "debug")

		// Create a test data structure (though we can't create Engine directly without a file)
		testData := &query.FoundationFoodsData{
			FoundationFoods: []query.FoundationFood{
				{Description: "Test food", FdcId: 1},
			},
		}

		// Create a mock engine (this would be better with a proper interface)
		mockEngine := &testQueryEngine{data: testData}
		authenticator := auth.NewBearerTokenAuth("test-token")

		server := NewServer(mockEngine, authenticator, logger)

		assert.NotNil(t, server)
		assert.NotNil(t, server.mcpServer)
		assert.Equal(t, mockEngine, server.queryEngine)
		assert.Equal(t, authenticator, server.auth)
		assert.Equal(t, logger, server.log)
	})
}

// testQueryEngine is a mock implementation for testing
type testQueryEngine struct {
	data *query.FoundationFoodsData
}

func (t *testQueryEngine) SearchFoodsByName(ctx context.Context, query string, limit int) ([]query.FoundationFood, error) {
	return t.data.FoundationFoods, nil
}

func (t *testQueryEngine) GetFoodByFdcId(ctx context.Context, fdcId int) (*query.FoundationFood, error) {
	for _, food := range t.data.FoundationFoods {
		if food.FdcId == fdcId {
			return &food, nil
		}
	}
	return nil, nil
}

func (t *testQueryEngine) Health(ctx context.Context) error {
	return nil
}
