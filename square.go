package main

// マスの座標
type Square int

const A1, H1, A8, H8 Square = 91, 98, 21, 28

func (s Square) Flip() Square {
	// BOARD は 0~119 の要素が 120 個の配列
	return 119 - s
}

func (s Square) String() string {
	// これなるほどって感じ
	return string([]byte{" abcdefgh "[s%10], "  87654321  "[s/10]})
}
