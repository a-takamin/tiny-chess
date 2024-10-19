package main

import (
	"errors"
	"strings"
)

// Board は 120 個の Piece で埋められた配列だと考える
// 8 × 8 = 64 では？と思うかもしれないが、上下に余白を 2 行ずつ増やし、左右に余白を 1 行ずつ増やしているため、(2+8+2) × (1+8+1) = 120 になっている。
// この余白によって、ピースのイリーガルムーブを判定する
// 2 行の余白は、ナイトが最大 2 行分一気に移動するから。
// 左右が 2 ではなく 1 列の余白で間に合っているのは、配列において右端（例えば 9）と左端（例えば 10）はつながっているため。
const BOARD_ARRAY_LENGTH = 120

type Board [BOARD_ARRAY_LENGTH]Piece

func (b Board) Flip() (bb Board) {
	for i := BOARD_ARRAY_LENGTH - 1; i >= 0; i-- {
		bb[i] = b[BOARD_ARRAY_LENGTH-i-1].Flip()
	}
	return b
}

func (b Board) String() (s string) {
	s = "\n"
	// 8 × 8 の部分のみ考える
	for row := 2; row < 10; row++ {
		for col := 1; col < 9; col++ {
			s = s + string(b[row*10+col])
		}
	}
	return s
}

// FEN から Board を作る
// FEN とはある盤面の状態を文字列で表したもの
// https://www.chess.com/terms/fen-chess
func NewBoardFromFEN(fen string) (b Board, err error) {
	parts := strings.Split(fen, " ")
	rows := strings.Split(parts[0], "/")
	if len(rows) != 8 {
		return b, errors.New("FEN should have 8 rows")
	}
	// コマがない状態で初期化
	for i := 0; i < BOARD_ARRAY_LENGTH; i++ {
		b[i] = ' '
	}
	for i := 0; i < 8; i++ {
		// 最初の余白2行をすっ飛ばしている
		index := i*10 + 21
		// string の range loop は char が取り出される
		for _, c := range rows[i] {
			// TODO: なんで q ?
			q := Piece(c)
			// コマが無いかどうか判定（FEN ではコマが無いマスは数字で表される
			if q >= '1' && q <= '8' {
				for j := Piece(0); q-j >= '1'; j++ {
					b[index] = '.'
					index++
				}
			} else if q.value() == 0 && q.Flip().value() == 0 {
				return b, errors.New("invalid piece value: " + string(c))
			} else {
				b[index] = q
				index++
			}
		}
		// その行に 8 個のコマを正しく入れ終わったら 10N + 9 番目にいるはずなので。
		if index%10 != 9 {
			return b, errors.New("invalid row length")
		}
	}
	// FEN では空白で区切られた 2 つ目の部分で次に動かす側を表す。w か b
	if len(parts) > 1 && parts[1] == "b" {
		b = b.Flip()
	}
	return b, nil
}
