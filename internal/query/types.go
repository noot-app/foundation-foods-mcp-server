package query

import (
	"context"
)

// FoundationFoodsData represents the root structure of the USDA Foundation Foods dataset
type FoundationFoodsData struct {
	FoundationFoods []FoundationFood `json:"FoundationFoods"`
}

// FoundationFood represents a single food item in the Foundation Foods dataset
type FoundationFood struct {
	FoodClass                 string         `json:"foodClass"`
	Description               string         `json:"description"`
	FoodNutrients             []FoodNutrient `json:"foodNutrients"`
	FoodAttributes            []interface{}  `json:"foodAttributes"`
	NutrientConversionFactors []interface{}  `json:"nutrientConversionFactors"`
	IsHistoricalReference     bool           `json:"isHistoricalReference"`
	NdbNumber                 int            `json:"ndbNumber"`
	DataType                  string         `json:"dataType"`
	FoodCategory              FoodCategory   `json:"foodCategory"`
	FdcId                     int            `json:"fdcId"`
	FoodPortions              []FoodPortion  `json:"foodPortions"`
	PublicationDate           string         `json:"publicationDate"`
	InputFoods                []InputFood    `json:"inputFoods"`
}

// FoodNutrient represents nutritional information for a food item
type FoodNutrient struct {
	Type                   string                 `json:"type"`
	Id                     int                    `json:"id"`
	Nutrient               Nutrient               `json:"nutrient"`
	DataPoints             int                    `json:"dataPoints,omitempty"`
	FoodNutrientDerivation FoodNutrientDerivation `json:"foodNutrientDerivation"`
	Max                    float64                `json:"max,omitempty"`
	Min                    float64                `json:"min,omitempty"`
	Median                 float64                `json:"median,omitempty"`
	Amount                 float64                `json:"amount"`
}

// Nutrient represents a specific nutrient
type Nutrient struct {
	Id       int    `json:"id"`
	Number   string `json:"number"`
	Name     string `json:"name"`
	Rank     int    `json:"rank"`
	UnitName string `json:"unitName"`
}

// FoodNutrientDerivation represents how a nutrient value was derived
type FoodNutrientDerivation struct {
	Code               string             `json:"code"`
	Description        string             `json:"description"`
	FoodNutrientSource FoodNutrientSource `json:"foodNutrientSource"`
}

// FoodNutrientSource represents the source of nutrient data
type FoodNutrientSource struct {
	Id          int    `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

// FoodCategory represents the category a food belongs to
type FoodCategory struct {
	Id          int    `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

// FoodPortion represents a food portion/serving size
type FoodPortion struct {
	Id              int         `json:"id"`
	Value           float64     `json:"value"`
	MeasureUnit     MeasureUnit `json:"measureUnit"`
	GramWeight      float64     `json:"gramWeight"`
	SequenceNumber  int         `json:"sequenceNumber"`
	Amount          float64     `json:"amount"`
	MinYearAcquired int         `json:"minYearAcquired"`
}

// MeasureUnit represents a unit of measurement
type MeasureUnit struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
}

// InputFood represents an input food used in composite foods
type InputFood struct {
	Id              int             `json:"id"`
	FoodDescription string          `json:"foodDescription"`
	InputFood       InputFoodDetail `json:"inputFood"`
}

// InputFoodDetail represents detailed information about an input food
type InputFoodDetail struct {
	FoodClass       string       `json:"foodClass"`
	Description     string       `json:"description"`
	DataType        string       `json:"dataType"`
	FoodCategory    FoodCategory `json:"foodCategory"`
	FdcId           int          `json:"fdcId"`
	PublicationDate string       `json:"publicationDate"`
}

// SearchProductsResponse represents the response from a food search
type SearchProductsResponse struct {
	Found    bool             `json:"found"`
	Count    int              `json:"count"`
	Products []FoundationFood `json:"products"`
}

// SearchResult represents a single search result with relevance score
type SearchResult struct {
	Food  FoundationFood
	Score float64
}

// QueryEngine defines the interface for querying Foundation Foods data
type QueryEngine interface {
	// SearchFoodsByName searches for foods by their description/name
	SearchFoodsByName(ctx context.Context, query string, limit int) ([]FoundationFood, error)

	// GetFoodByFdcId retrieves a specific food by its FDC ID
	GetFoodByFdcId(ctx context.Context, fdcId int) (*FoundationFood, error)

	// Health checks if the query engine is ready and operational
	Health(ctx context.Context) error
}
