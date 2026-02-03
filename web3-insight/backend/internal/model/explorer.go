package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// ExplorerResearch stores research data about blockchain explorers
type ExplorerResearch struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ChainName       string         `gorm:"size:100;not null;index" json:"chainName"`
	ChainType       string         `gorm:"size:50" json:"chainType"` // L1, L2, sidechain
	ExplorerName    string         `gorm:"size:100;not null" json:"explorerName"`
	ExplorerURL     string         `gorm:"size:500;not null" json:"explorerUrl"`
	ExplorerType    string         `gorm:"size:50" json:"explorerType"` // official, third-party
	Features        datatypes.JSON `gorm:"type:jsonb" json:"features"`   // Feature checklist
	UIFeatures      datatypes.JSON `gorm:"type:jsonb" json:"uiFeatures"` // UI/UX features
	APIFeatures     datatypes.JSON `gorm:"type:jsonb" json:"apiFeatures"` // API capabilities
	Screenshots     pq.StringArray `gorm:"type:text[]" json:"screenshots"`
	Analysis        string         `gorm:"type:text" json:"analysis"`
	Strengths       pq.StringArray `gorm:"type:text[]" json:"strengths"`
	Weaknesses      pq.StringArray `gorm:"type:text[]" json:"weaknesses"`
	PopularityScore float64        `gorm:"default:0" json:"popularityScore"`
	ResearchStatus  string         `gorm:"size:20;default:'pending'" json:"researchStatus"` // pending, in_progress, completed
	ResearchNotes   string         `gorm:"type:text" json:"researchNotes"`
	LastUpdated     time.Time      `gorm:"default:now()" json:"lastUpdated"`
	CreatedAt       time.Time      `json:"createdAt"`
}

func (ExplorerResearch) TableName() string {
	return "explorer_research"
}

// ExplorerFeature represents a feature to check for in explorers
type ExplorerFeature struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Category    string    `gorm:"size:50;not null" json:"category"`    // core, advanced, defi, nft, api
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Importance  string    `gorm:"size:20;default:'medium'" json:"importance"` // high, medium, low
	SortOrder   int       `gorm:"default:0" json:"sortOrder"`
}

func (ExplorerFeature) TableName() string {
	return "explorer_features"
}

// FeatureChecklistItem represents a single feature check result
type FeatureChecklistItem struct {
	FeatureID   string `json:"featureId"`
	FeatureName string `json:"featureName"`
	Category    string `json:"category"`
	Supported   bool   `json:"supported"`
	Notes       string `json:"notes,omitempty"`
	Quality     string `json:"quality,omitempty"` // excellent, good, basic, poor
}

// StandardFeatureCategories defines the categories for feature checklist
var StandardFeatureCategories = []string{
	"core",     // Basic blockchain data viewing
	"advanced", // Advanced analysis features
	"defi",     // DeFi-specific features
	"nft",      // NFT-specific features
	"api",      // API capabilities
	"ux",       // User experience features
}

// StandardFeatures defines the standard features to check
var StandardFeatures = []ExplorerFeature{
	// Core features
	{Category: "core", Name: "Transaction Search", Description: "Search transactions by hash", Importance: "high", SortOrder: 1},
	{Category: "core", Name: "Address Lookup", Description: "View address balance and history", Importance: "high", SortOrder: 2},
	{Category: "core", Name: "Block Explorer", Description: "Browse blocks and block details", Importance: "high", SortOrder: 3},
	{Category: "core", Name: "Token Tracking", Description: "View token balances and transfers", Importance: "high", SortOrder: 4},
	{Category: "core", Name: "Contract Verification", Description: "Verify and view source code", Importance: "high", SortOrder: 5},
	{Category: "core", Name: "Contract Interaction", Description: "Read/write contract functions", Importance: "medium", SortOrder: 6},

	// Advanced features
	{Category: "advanced", Name: "Internal Transactions", Description: "View internal/trace calls", Importance: "medium", SortOrder: 10},
	{Category: "advanced", Name: "Token Approvals", Description: "View and revoke approvals", Importance: "medium", SortOrder: 11},
	{Category: "advanced", Name: "Gas Tracker", Description: "Real-time gas price tracking", Importance: "medium", SortOrder: 12},
	{Category: "advanced", Name: "Charts & Analytics", Description: "Historical data visualization", Importance: "low", SortOrder: 13},
	{Category: "advanced", Name: "Watchlist", Description: "Save addresses to watch", Importance: "low", SortOrder: 14},
	{Category: "advanced", Name: "Address Labels", Description: "Known address labeling", Importance: "medium", SortOrder: 15},

	// DeFi features
	{Category: "defi", Name: "DEX Trades", Description: "Decode DEX transactions", Importance: "medium", SortOrder: 20},
	{Category: "defi", Name: "Liquidity Pools", Description: "LP token and pool tracking", Importance: "medium", SortOrder: 21},
	{Category: "defi", Name: "Yield Farming", Description: "Farming position tracking", Importance: "low", SortOrder: 22},
	{Category: "defi", Name: "Token Price", Description: "Show token USD values", Importance: "medium", SortOrder: 23},

	// NFT features
	{Category: "nft", Name: "NFT Gallery", Description: "View NFTs in address", Importance: "medium", SortOrder: 30},
	{Category: "nft", Name: "NFT Metadata", Description: "Display NFT attributes", Importance: "medium", SortOrder: 31},
	{Category: "nft", Name: "Collection Stats", Description: "Collection-level analytics", Importance: "low", SortOrder: 32},

	// API features
	{Category: "api", Name: "REST API", Description: "Public REST API access", Importance: "high", SortOrder: 40},
	{Category: "api", Name: "API Documentation", Description: "Comprehensive API docs", Importance: "high", SortOrder: 41},
	{Category: "api", Name: "Rate Limits", Description: "Reasonable rate limits", Importance: "medium", SortOrder: 42},
	{Category: "api", Name: "Webhooks", Description: "Event notification webhooks", Importance: "low", SortOrder: 43},
	{Category: "api", Name: "GraphQL", Description: "GraphQL API support", Importance: "low", SortOrder: 44},

	// UX features
	{Category: "ux", Name: "Mobile Responsive", Description: "Works well on mobile", Importance: "medium", SortOrder: 50},
	{Category: "ux", Name: "Dark Mode", Description: "Dark theme support", Importance: "low", SortOrder: 51},
	{Category: "ux", Name: "Multi-language", Description: "Multiple language support", Importance: "low", SortOrder: 52},
	{Category: "ux", Name: "Search Suggestions", Description: "Auto-complete in search", Importance: "low", SortOrder: 53},
}
