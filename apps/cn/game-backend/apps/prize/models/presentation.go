package models

import "strings"

// PrizeText selects each field independently so older untranslated records remain readable.
func PrizeText(zh, en, lang string) string {
	if lang == "en" && strings.TrimSpace(en) != "" {
		return en
	}
	return zh
}

// Display excludes redemption keys, including from the shared history cache.
func (prize PrizeModel) Display() PrizeDisplay {
	return PrizeDisplay{Title: prize.Title, TitleEn: prize.TitleEn, Platform: prize.Platform, PlatformEn: prize.PlatformEn, Count: len(prize.Keys)}
}

func (prize PrizeDisplay) Localized(lang string) PrizeDisplay {
	return PrizeDisplay{Title: PrizeText(prize.Title, prize.TitleEn, lang), Platform: PrizeText(prize.Platform, prize.PlatformEn, lang), Count: prize.Count}
}

func (history PrizeWinnerCacheSaveModel) Localized(lang string) PrizeWinnerCacheSaveModel {
	if history.Prize == nil {
		return history
	}
	items := make([]PrizeCacheSaveModel, len(history.Prize))
	for i, item := range history.Prize {
		item.Name = PrizeText(item.Name, item.NameEn, lang)
		item.Desc = PrizeText(item.Desc, item.DescEn, lang)
		item.NameEn, item.DescEn = "", ""
		item.Prize = item.Prize.Localized(lang)
		items[i] = item
	}
	history.Prize = items
	return history
}
