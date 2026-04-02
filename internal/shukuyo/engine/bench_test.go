package engine

import "testing"

func BenchmarkMansionIndex(b *testing.B) {
	bd := mustParseDate("1977-01-15")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MansionIndexFromDate(bd)
	}
}

func BenchmarkGetRelation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetRelation(5, 15)
	}
}

func BenchmarkCompatibility(b *testing.B) {
	d1 := mustParseDate("1977-01-15")
	d2 := mustParseDate("1990-05-20")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Compatibility(d1, d2)
	}
}

func BenchmarkKuyouStar(b *testing.B) {
	bd := mustParseDate("1977-01-15")
	for i := 0; i < b.N; i++ {
		GetKuyouStar(bd, 2026)
	}
}

func BenchmarkSankiPosition(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetSankiPosition(5, 15)
	}
}
