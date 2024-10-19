package main

const N, E, S, W = -10, 1, 10, 1

type Move struct {
	from Square
	to   Square
}

// 差し手を e2e4 のような形式で表す
func (m Move) String() string {
	return m.from.String() + m.to.String()
}
