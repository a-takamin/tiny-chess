package main

// 符号付きの整数を右へシフトすると、それは論理シフトではなく算術シフトなので、最上位は符号の値（0 or 1）で埋められる
// つまり >>63 すると、すべて 0 かすべて 1 の整数になる
// それとの排他的論理和を取るので、負だった場合はすべてが反転する
// ビットを反転させて 1 を引くと符号が反転するので、絶対値を取れている
// ※ 正の場合は 000... との排他的論理和をとって 000... を引くので何も起こらない
func abs(n int) int { return int((int64(n) ^ int64(n)>>63) - int64(n)>>63) }

// Position は今のゲームの状態を表す構造体
type Position struct {
	board Board
	score int
	wc    [2]bool // 白のキャスリング（0: キングサイド, 1: クイーンサイド）権
	bc    [2]bool // 黒のキャスリング権
	ep    Square  // アンパッサンの対象マス
	kp    Square  // キングの位置
}

func (pos Position) Flip() Position {
	np := Position{
		score: -pos.score,
		wc:    [2]bool{pos.bc[0], pos.bc[1]},
		bc:    [2]bool{pos.wc[0], pos.wc[1]},
		ep:    pos.ep.Flip(),
		kp:    pos.kp.Flip(),
	}
	np.board = pos.board.Flip()
	return np
}

// 今の Position における可能な Move をすべて算出する
func (pos Position) Moves() (moves []Move) {
	var directions = map[Piece][]Square{
		'P': {N, N + N, N + W, N + E},
		'N': {N + N + E, N + N + W, E + E + N, E + E + S, S + S + E, S + S + W, W + W + S, W + W + N},
		'B': {N + E, N + W, S + E, S + W},
		'R': {N, E, S, W},
		'Q': {N, E, S, W, N + E, N + W, S + E, S + W},
		'K': {N, E, S, W, N + E, N + W, S + E, S + W},
	}
	for index, p := range pos.board {
		if !p.ours() {
			continue
		}
		i := Square(index)
		for _, d := range directions[p] {
			for j := i + d; ; j = j + d {
				// q は to の座標
				q := pos.board[j]
				// ' ' は余白部分。'.' はコマが無い。つまり、味方コマにぶつかった
				if q == ' ' || (q != '.' && q.ours()) {
					break
				}
				if p == 'P' {
					// コマが既にある
					if (d == N || d == N+N) && q != '.' {
						break
					}
					// 初手の 2 マス動きについて。初手であれば i は A1 + N 以上なはず。後は 1 マス先にコマが既にある
					if d == N+N && (i < A1+N || pos.board[i+N] != '.') {
						break
					}
				}
				moves = append(moves, Move{from: i, to: j})
				// P, N, K は判定は 1 回だけ。※ B, R, Q はさらに次のマスを考える
				if p == 'P' || p == 'N' || p == 'K' {
					break
				}
				// B, R, Q は敵コマにぶつかったのでこれ以上同じ方角へはいけない
				if q != ' ' && q != '.' && !q.ours() {
					break
				}
				// ルークの Move の視点でクイーンサイドキャスリングの K の動きを追加する（i が A1 のときに、to+E が K になる動きができるのは R か Q しかいない）
				// j+E = K ってことはあいだの K, B がいなかったことが決まっている。美しい条件だ。
				// TODO: これは i = A1 = Q のときでもクイーンサイドキャスリングが選択肢に入ってしまうのでは？後でFEN がそれの Board を作って Moves() を呼ぶテストを書きたい
				//// 下の Move() を読むと A1 が動いたら wc[0] = false になるので大丈夫そうだけど、テスト次第では R かどうかの条件も入れたい。
				if i == A1 && pos.board[j+E] == 'K' && pos.wc[0] {
					moves = append(moves, Move{
						from: j + E,
						to:   j + W,
					})
				}
				// キングサイドキャスリング
				if i == H1 && pos.board[j+W] == 'K' && pos.wc[1] {
					moves = append(moves, Move{
						from: j + W,
						to:   j + E,
					})
				}
			}
		}
	}
	return moves
}

// Move を実行し、新しい Position を返す
func (pos Position) Move(m Move) (np Position) {
	i, j, p := m.from, m.to, pos.board[m.from]
	np = pos
	np.ep = 0
	np.kp = 0
	np.score = pos.score + pos.value(m)
	np.board[m.to] = pos.board[m.from]
	np.board[m.from] = '.'
	// R が動いたらキャスリング不可
	if i == A1 {
		np.wc[0] = false
	}
	if i == H1 {
		np.wc[1] = false
	}
	if i == A8 {
		np.bc[1] = false
	}
	if i == H8 {
		np.bc[0] = false
	}
	// K が動いたら
	if p == 'K' {
		// キャスリング不可
		np.wc[0], np.wc[1] = false, false
		// キャスリングの場合
		if abs(int(j-i)) == 2 {
			if j < i {
				np.board[H1] = '.'
			} else {
				np.board[A1] = '.'
			}
			np.board[(i+j)/2] = 'R'
		}
	}
	if p == 'P' {
		if A8 <= j && j <= H8 {
			// TODO: アンダープロモーションは実装されていないようだ
			np.board[j] = 'Q'
		}
		if j-i == N+N {
			np.ep = i + N
		}
		if j == pos.ep {
			np.board[j+S] = '.'
		}
	}
	return np.Flip()
}

// Move による評価値の揺れを返す
func (pos Position) value(m Move) int {
	// 盤面の位置により評価が決まっているらしい。
	// もちろんこの評価は固定ではない。
	pst := map[Piece][120]int{
		'P': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 178, 183, 186, 173, 202, 182, 185, 190, 0, 0, 107, 129, 121, 144, 140, 131, 144, 107, 0, 0, 83, 116, 98, 115, 114, 0, 115, 87, 0, 0, 74, 103, 110, 109, 106, 101, 0, 77, 0, 0, 78, 109, 105, 89, 90, 98, 103, 81, 0, 0, 69, 108, 93, 63, 64, 86, 103, 69, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		'N': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 214, 227, 205, 205, 270, 225, 222, 210, 0, 0, 277, 274, 380, 244, 284, 342, 276, 266, 0, 0, 290, 347, 281, 354, 353, 307, 342, 278, 0, 0, 304, 304, 325, 317, 313, 321, 305, 297, 0, 0, 279, 285, 311, 301, 302, 315, 282, 0, 0, 0, 262, 290, 293, 302, 298, 295, 291, 266, 0, 0, 257, 265, 282, 0, 282, 0, 257, 260, 0, 0, 206, 257, 254, 256, 261, 245, 258, 211, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		'B': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 261, 242, 238, 244, 297, 213, 283, 270, 0, 0, 309, 340, 355, 278, 281, 351, 322, 298, 0, 0, 311, 359, 288, 361, 372, 310, 348, 306, 0, 0, 345, 337, 340, 354, 346, 345, 335, 330, 0, 0, 333, 330, 337, 343, 337, 336, 0, 327, 0, 0, 334, 345, 344, 335, 328, 345, 340, 335, 0, 0, 339, 340, 331, 326, 327, 326, 340, 336, 0, 0, 313, 322, 305, 308, 306, 305, 310, 310, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		'R': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 514, 508, 512, 483, 516, 512, 535, 529, 0, 0, 534, 508, 535, 546, 534, 541, 513, 539, 0, 0, 498, 514, 507, 512, 524, 506, 504, 494, 0, 0, 0, 484, 495, 492, 497, 475, 470, 473, 0, 0, 451, 444, 463, 458, 466, 450, 433, 449, 0, 0, 437, 451, 437, 454, 454, 444, 453, 433, 0, 0, 426, 441, 448, 453, 450, 436, 435, 426, 0, 0, 449, 455, 461, 484, 477, 461, 448, 447, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		'Q': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 935, 930, 921, 825, 998, 953, 1017, 955, 0, 0, 943, 961, 989, 919, 949, 1005, 986, 953, 0, 0, 927, 972, 961, 989, 1001, 992, 972, 931, 0, 0, 930, 913, 951, 946, 954, 949, 916, 923, 0, 0, 915, 914, 927, 924, 928, 919, 909, 907, 0, 0, 899, 923, 916, 918, 913, 918, 913, 902, 0, 0, 893, 911, 0, 910, 914, 914, 908, 891, 0, 0, 890, 899, 898, 916, 898, 893, 895, 887, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		'K': {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 60004, 60054, 60047, 59901, 59901, 60060, 60083, 59938, 0, 0, 59968, 60010, 60055, 60056, 60056, 60055, 60010, 60003, 0, 0, 59938, 60012, 59943, 60044, 59933, 60028, 60037, 59969, 0, 0, 59945, 60050, 60011, 59996, 59981, 60013, 0, 59951, 0, 0, 59945, 59957, 59948, 59972, 59949, 59953, 59992, 59950, 0, 0, 59953, 59958, 59957, 59921, 59936, 59968, 59971, 59968, 0, 0, 59996, 60003, 59986, 59950, 59943, 59982, 60013, 60004, 0, 0, 60017, 60030, 59997, 59986, 60006, 59999, 60040, 60018, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}

	i, j := m.from, m.to
	p, q := Piece(pos.board[i]), Piece(pos.board[j])

	// to - from での評価値の揺れがベース
	score := pst[p][j] - pst[p][i]

	// 取った敵コマの分をこちらに加算する
	if q != '.' && q != ' ' && !q.ours() {
		score += pst[q.Flip()][j.Flip()]
	}
	// キャスリングによるキングの位置の評価値の揺れを加算
	if abs(int(j-pos.kp)) < 2 {
		score += pst['K'][j.Flip()]
	}
	// キャスリングによるルークの位置の評価値の揺れを加算
	if p == 'K' && (abs(int(i-j))) == 2 {
		score = score + pst['R'][(i+j)/2]
		if j < i {
			score = score - pst['R'][A1]
		} else {
			score = score - pst['R'][H1]
		}
	}
	if p == 'P' {
		// プロモーションによる評価値の揺れを加算
		if A8 <= j && j <= H8 {
			score += pst['Q'][j] - pst['P'][j]
		}
		// アンパッサン
		if j == pos.ep {
			score += pst['P'][(j + S).Flip()]
		}
	}
	return score
}
