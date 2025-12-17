package game

import "math"

// MaxHPForLevel computes MaxHP for a given level (1..999) using piecewise-linear interpolation.
func MaxHPForLevel(lv int) int {
	if lv < 1 {
		lv = 1
	}
	if lv <= 30 {
		v := 30.0 + math.Round(270.0*float64(lv-1)/29.0)
		return int(v)
	}
	v := 300.0 + math.Round(699.0*float64(lv-30)/969.0)
	return int(v)
}

// AllowedMisses computes allowed misses M for a session length N.
func AllowedMisses(N int) int {
	if N <= 0 {
		return 1
	}
	mBase := int(math.Ceil(float64(N) * 0.20))
	if N <= 10 {
		if 2 > mBase {
			return 2
		}
		return mBase
	}
	if 1 > mBase {
		return 1
	}
	return mBase
}

// DamagePerMiss computes damage per miss so that M+1 misses kill the player.
func DamagePerMiss(maxHP, M int) int {
	if M < 0 {
		M = 0
	}
	return int(math.Ceil(float64(maxHP) / float64(M+1)))
}

// TierForLevel returns tier index (1..6) and multiplier.
func TierForLevel(lv int) (int, float64) {
	switch {
	case lv >= 1 && lv <= 19:
		return 1, 1.0
	case lv >= 20 && lv <= 49:
		return 2, 1.2
	case lv >= 50 && lv <= 99:
		return 3, 1.5
	case lv >= 100 && lv <= 199:
		return 4, 1.9
	case lv >= 200 && lv <= 399:
		return 5, 2.4
	default:
		return 6, 3.0
	}
}

// QExpFor computes per-question EXP given baseExp, tier multiplier and rarity.
func QExpFor(baseExp int, tierMul float64, rare bool) int {
	rarityMul := 1.0
	if rare {
		rarityMul = 2.0
	}
	v := math.Round(float64(baseExp) * tierMul * rarityMul)
	return int(v)
}

// ClearBonus computes the clear bonus for a session.
func ClearBonus(N int, baseExp int, tierMul float64) int {
	v := math.Round(float64(N*baseExp) * tierMul * 0.5)
	return int(v)
}

// PerfectBonusMul returns the perfect bonus multiplier for N questions.
func PerfectBonusMul(N int) float64 {
	return 1.20 + 0.01*float64(N)
}

// SessionExpClear computes total SessionExp on clear; if allCorrect is true and perfectEnabled is true, applies perfect multiplier.
func SessionExpClear(sumCorrectExp int, clearBonus int, allCorrect bool, N int, perfectEnabled bool) int {
	res := sumCorrectExp + clearBonus
	if allCorrect && perfectEnabled {
		res = int(math.Round(float64(res) * PerfectBonusMul(N)))
	}
	return res
}

// SessionExpFail computes SessionExp on failure.
func SessionExpFail(sumCorrectExp int, failFactor float64) int {
	if failFactor < 0 {
		failFactor = 0
	}
	return int(math.Floor(float64(sumCorrectExp) * failFactor))
}

// ExpToNext computes experience required to reach the next level.
func ExpToNext(level int) int {
	if level <= 99 {
		return 30 + 5*(level-1)
	}
	return 500 + 10*(level-100)
}
