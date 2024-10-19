package main

type Piece byte

func (p Piece) value() int {
	// キングは最低でも Q 9個, R 2個, B 2個, N 2個の合計値より大きくする必要がある。
	// ちなみに合計値は 10500 程度なので 60000 は十分。
	// なお、map [] の直接取り出しで存在しないキーを入力した場合、ゼロ値が返る: https://go.dev/blog/maps
	return map[Piece]int{'P': 100, 'N': 280, 'B': 320, 'R': 479, 'Q': 929, 'K': 60000}[p]
}

func (p Piece) ours() bool {
	return p.value() > 0
}

// 自コマ（白）を大文字、敵コマ（黒）を小文字で表す。
// flip でひっくり返すことで、盤面の計算などはすべて白だけを考えればよくなる
func (p Piece) Flip() Piece {
	return map[Piece]Piece{'P': 'p', 'N': 'n', 'B': 'b', 'R': 'r', 'Q': 'q', 'K': 'k', 'p': 'P', 'n': 'N', 'b': 'B', 'r': 'R', 'q': 'Q', 'k': 'K', ' ': ' ', '.': '.'}[p]
}
