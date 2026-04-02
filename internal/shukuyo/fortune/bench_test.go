package fortune

import "testing"

func BenchmarkCalculateDaily(b *testing.B) {
	birth := bd("1977-01-15")
	target := bd("2026-04-02")
	for i := 0; i < b.N; i++ {
		CalculateDaily(birth, target, "zh-TW")
	}
}

func BenchmarkCalculateWeekly(b *testing.B) {
	birth := bd("1977-01-15")
	target := bd("2026-04-02")
	for i := 0; i < b.N; i++ {
		CalculateWeekly(birth, target, "zh-TW")
	}
}

func BenchmarkCalculateMonthly(b *testing.B) {
	birth := bd("1977-01-15")
	for i := 0; i < b.N; i++ {
		CalculateMonthly(birth, 2026, 4, "zh-TW")
	}
}

func BenchmarkCalculateYearly(b *testing.B) {
	birth := bd("1977-01-15")
	for i := 0; i < b.N; i++ {
		CalculateYearly(birth, 2026, "zh-TW")
	}
}
