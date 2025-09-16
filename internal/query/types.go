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
	UnitName   string  `json:"unitName"`
	Amount     float64 `json:"amount"`
	DataPoints int     `json:"dataPoints,omitempty"`
	Max        float64 `json:"max,omitempty"`
	Min        float64 `json:"min,omitempty"`
	Median     float64 `json:"median,omitempty"`
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
	"Water",
	"Energy",
	"Energy (Atwater General Factors)",
	"Energy (Atwater Specific Factors)",
	"Protein",
	"Total lipid (fat)",
	"Ash",
	"Nitrogen",

	// Carbohydrates and sugars
	"Carbohydrate, by difference",
	"Fiber, total dietary",
	"Starch",
	"Sugars, Total", // Total sugars (126 foods)
	"Total Sugars",  // Alternative sugar naming variant (5 foods)
	"Fructose",      // Individual sugar components
	"Glucose",
	"Sucrose",
	"Lactose",
	"Maltose",
	"Galactose",

	// Fats and fatty acids
	"Fatty acids, total saturated",
	"Fatty acids, total trans",
	"Fatty acids, total monounsaturated",
	"Fatty acids, total polyunsaturated",
	"Cholesterol",
	"Total fat (NLEA)", // NLEA compliant total fat
	// Saturated fatty acids (SFA)
	"SFA 4:0",  // Butyric acid
	"SFA 6:0",  // Caproic acid
	"SFA 8:0",  // Caprylic acid
	"SFA 10:0", // Capric acid
	"SFA 12:0", // Lauric acid
	"SFA 14:0", // Myristic acid
	"SFA 15:0", // Pentadecanoic acid
	"SFA 16:0", // Palmitic acid
	"SFA 17:0", // Margaric acid
	"SFA 18:0", // Stearic acid
	"SFA 20:0", // Arachidic acid
	"SFA 24:0", // Lignoceric acid
	// Monounsaturated fatty acids (MUFA)
	"MUFA 14:1 c", // Myristoleic acid
	"MUFA 16:1 c", // Palmitoleic acid
	"MUFA 18:1 c", // Oleic acid
	"MUFA 20:1 c", // Gadoleic acid
	// Polyunsaturated fatty acids (PUFA)
	"PUFA 18:2 c",               // Linoleic acid (alternative naming)
	"PUFA 18:2 n-6 c,c",         // Linoleic acid (specific naming)
	"PUFA 18:3 c",               // Alpha-linolenic acid (alternative naming)
	"PUFA 18:3 n-3 c,c,c (ALA)", // Alpha-linolenic acid (specific naming)
	"PUFA 20:3 c",               // Dihomo-gamma-linolenic acid
	"PUFA 20:3 n-6",             // Dihomo-gamma-linolenic acid (n-6)
	"PUFA 20:4",                 // Arachidonic acid
	"PUFA 20:5 n-3 (EPA)",       // Eicosapentaenoic acid
	"PUFA 22:5 n-3 (DPA)",       // Docosapentaenoic acid
	"PUFA 22:6 n-3 (DHA)",       // Docosahexaenoic acid

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
