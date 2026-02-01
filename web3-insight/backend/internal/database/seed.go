package database

import (
	"github.com/user/web3-insight/internal/model"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	categories := []model.Category{
		{Name: "基础技术", NameEn: "Fundamentals", Slug: "fundamentals", Icon: "book", SortOrder: 1},
		{Name: "扩容方案", NameEn: "Scaling Solutions", Slug: "scaling", Icon: "layers", SortOrder: 2},
		{Name: "跨链技术", NameEn: "Cross-chain", Slug: "cross-chain", Icon: "link", SortOrder: 3},
		{Name: "生态系统", NameEn: "Ecosystems", Slug: "ecosystems", Icon: "globe", SortOrder: 4},
		{Name: "Explorer 技术", NameEn: "Explorer Tech", Slug: "explorer-tech", Icon: "search", SortOrder: 5},
		{Name: "行业动态", NameEn: "Industry News", Slug: "news", Icon: "newspaper", SortOrder: 6},
	}

	for _, cat := range categories {
		db.FirstOrCreate(&cat, model.Category{Slug: cat.Slug})
	}

	return nil
}
