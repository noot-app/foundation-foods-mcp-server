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

	// SearchFoodsByNameSimplified searches for foods and returns simplified nutrient information
	SearchFoodsByNameSimplified(ctx context.Context, query string, limit int, nutrientsToInclude []string) (*SimplifiedNutrientResponse, error)

	// GetFoodByFdcId retrieves a specific food by its FDC ID
	GetFoodByFdcId(ctx context.Context, fdcId int) (*FoundationFood, error)

	// Health checks if the query engine is ready and operational
	Health(ctx context.Context) error
}

// SimplifiedNutrient represents a nutrient with only essential information
type SimplifiedNutrient struct {
	Name       string  `json:"name"`
	Unit       string  `json:"unit"`
	Amount     float64 `json:"amount"`
	DataPoints int     `json:"dataPoints"`
}

// SimplifiedMeasureUnit represents a simplified measure unit
type SimplifiedMeasureUnit struct {
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
}

// SimplifiedFoodPortion represents a simplified food portion
type SimplifiedFoodPortion struct {
	Value       float64               `json:"value"`
	MeasureUnit SimplifiedMeasureUnit `json:"measureUnit"`
	Modifier    string                `json:"modifier,omitempty"`
	GramWeight  float64               `json:"gramWeight"`
	Amount      float64               `json:"amount"`
}

// SimplifiedFood represents a food item with simplified nutrient information
type SimplifiedFood struct {
	Name         string                  `json:"name"`
	Nutrients    []SimplifiedNutrient    `json:"nutrients"`
	FoodPortions []SimplifiedFoodPortion `json:"foodPortions"`
}

// SimplifiedNutrientResponse represents the response for simplified nutrient searches
type SimplifiedNutrientResponse struct {
	Found bool             `json:"found"`
	Count int              `json:"count"`
	Foods []SimplifiedFood `json:"foods"`
}

// DefaultNutrients contains the standard set of nutrients to return by default
// Optimized based on comprehensive analysis of USDA Foundation Foods data
var DefaultNutrients = []string{
	// Basic composition
	"Energy",
	"Protein",
	"Total lipid (fat)",

	// Carbohydrates and sugars
	"Carbohydrate, by difference", // total_carbs_g
	"Fiber, total dietary",        // dietary_fiber_g
	"Sugars, Total",               // total_sugars_g (126 foods)
	"Total Sugars",                // total_sugars_g (alternative naming - 5 foods)

	// Fats and fatty acids
	"Total lipid (fat)",                  // total_fat_g
	"Total fat (NLEA)",                   // NLEA compliant total fat
	"Fatty acids, total saturated",       // saturated_fat_g
	"Fatty acids, total trans",           // trans_fat_g
	"Fatty acids, total monounsaturated", // monounsaturated_fat_g
	"Fatty acids, total polyunsaturated", // polyunsaturated_fat_g
	"Cholesterol",                        // cholesterol_mg
	// Essential omega fatty acids only
	"PUFA 18:3 n-3 c,c,c (ALA)", // omega3_ala_g (Alpha-linolenic acid)
	"PUFA 20:5 n-3 (EPA)",       // omega3_epa_g (Eicosapentaenoic acid)
	"PUFA 22:6 n-3 (DHA)",       // omega3_dha_g (Docosahexaenoic acid)
	"PUFA 18:2 n-6 c,c",         // omega6_g (Linoleic acid - primary omega-6)

	// Minerals
	"Sodium, Na",
	"Calcium, Ca",
	"Iron, Fe",
	"Magnesium, Mg",
	"Phosphorus, P",
	"Potassium, K",
	"Zinc, Zn",
	"Copper, Cu",
	"Manganese, Mn",
	"Selenium, Se",
	"Iodine, I",
	"Molybdenum, Mo",

	// Vitamins
	"Vitamin A, RAE",
	"Vitamin C, total ascorbic acid",
	"Vitamin D (D2 + D3)",
	"Vitamin E (alpha-tocopherol)",
	"Tocopherol, beta",  // Beta-tocopherol
	"Tocopherol, gamma", // Gamma-tocopherol
	"Tocopherol, delta", // Delta-tocopherol
	"Vitamin K (phylloquinone)",
	"Vitamin K (Dihydrophylloquinone)", // Dihydrophylloquinone
	"Vitamin K (Menaquinone-4)",        // Menaquinone-4
	"Thiamin",
	"Riboflavin",
	"Niacin",
	"Vitamin B-6",
	"Folate, total",
	"Vitamin B-12",
	"Biotin",
	"Pantothenic acid",
	"Choline, total",
}
