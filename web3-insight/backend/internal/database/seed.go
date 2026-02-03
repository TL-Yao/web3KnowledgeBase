package database

import (
	"github.com/google/uuid"
	"github.com/user/web3-insight/internal/model"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	// Create parent categories first
	parentCategories := []model.Category{
		{Name: "Layer 1", NameEn: "Layer 1", Slug: "layer-1", Icon: "layers", SortOrder: 1},
		{Name: "Layer 2", NameEn: "Layer 2", Slug: "layer-2", Icon: "layers", SortOrder: 2},
		{Name: "DeFi", NameEn: "DeFi", Slug: "defi", Icon: "coins", SortOrder: 3},
		{Name: "NFT", NameEn: "NFT", Slug: "nft", Icon: "image", SortOrder: 4},
		{Name: "钱包与安全", NameEn: "Wallet & Security", Slug: "wallet-security", Icon: "shield", SortOrder: 5},
	}

	// Map to store category IDs by slug
	categoryIDs := make(map[string]uuid.UUID)

	for _, cat := range parentCategories {
		var existing model.Category
		result := db.Where("slug = ?", cat.Slug).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			db.Create(&cat)
			categoryIDs[cat.Slug] = cat.ID
		} else {
			categoryIDs[cat.Slug] = existing.ID
		}
	}

	// Create child categories with parent references
	layer1ID := categoryIDs["layer-1"]
	layer2ID := categoryIDs["layer-2"]
	defiID := categoryIDs["defi"]

	childCategories := []model.Category{
		// Layer 1 children
		{Name: "Ethereum", NameEn: "Ethereum", Slug: "ethereum", ParentID: &layer1ID, Icon: "diamond", SortOrder: 1},
		{Name: "Solana", NameEn: "Solana", Slug: "solana", ParentID: &layer1ID, Icon: "zap", SortOrder: 2},
		{Name: "Cosmos", NameEn: "Cosmos", Slug: "cosmos", ParentID: &layer1ID, Icon: "globe", SortOrder: 3},
		// Layer 2 children
		{Name: "ZK Rollup", NameEn: "ZK Rollup", Slug: "zk-rollup", ParentID: &layer2ID, Icon: "lock", SortOrder: 1},
		{Name: "Optimistic Rollup", NameEn: "Optimistic Rollup", Slug: "optimistic-rollup", ParentID: &layer2ID, Icon: "clock", SortOrder: 2},
		// DeFi children
		{Name: "DEX", NameEn: "DEX", Slug: "dex", ParentID: &defiID, Icon: "repeat", SortOrder: 1},
		{Name: "Lending", NameEn: "Lending", Slug: "lending", ParentID: &defiID, Icon: "percent", SortOrder: 2},
		{Name: "Staking", NameEn: "Staking", Slug: "staking", ParentID: &defiID, Icon: "lock", SortOrder: 3},
	}

	for _, cat := range childCategories {
		var existing model.Category
		result := db.Where("slug = ?", cat.Slug).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			db.Create(&cat)
			categoryIDs[cat.Slug] = cat.ID
		} else {
			categoryIDs[cat.Slug] = existing.ID
		}
	}

	return nil
}
