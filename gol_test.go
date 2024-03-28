package main

import (
	"testing"
)

var num = 1

func BenchmarkNextGeneration(b *testing.B) {
	m := model{}
	m.field = make([][]int, 100)
	for i := 0; i < 100; i++ {
		m.field[i] = make([]int, 100)
	}
	m = generateField(m)
	for i := 0; i < b.N; i++ {
		m = nextGeneration(m)
	}
}
