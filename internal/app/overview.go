package app

import (
	"math/big"
	"strings"
)

// @impl WSH-005-AC6 WLS-011-AC3: use the same completeness rule for cards and overview counts.
func completeWish(w wishCard) bool {
	return w.Title != "" && w.Price != "" && len(w.Links) > 0
}

// @impl WLS-011-AC2 WLS-011-AC3 WLS-011-AC6
// Prices have already passed validateWish. Integer hundredths preserve every
// decimal, including prices larger than floating-point or machine integers.
func summarizeWishes(cards []wishCard) (total string, claimed, unclaimed int) {
	cents := new(big.Int)
	for _, card := range cards {
		if card.Price != "" {
			whole, fraction, _ := strings.Cut(card.Price, ".")
			value, _ := new(big.Int).SetString(whole+(fraction + "00")[:2], 10)
			cents.Add(cents, value)
		}
		if completeWish(card) {
			if len(card.Claims) > 0 {
				claimed++
			} else {
				unclaimed++
			}
		}
	}
	digits := cents.String()
	if len(digits) < 3 {
		digits = strings.Repeat("0", 3-len(digits)) + digits
	}
	return digits[:len(digits)-2] + "." + digits[len(digits)-2:], claimed, unclaimed
}
