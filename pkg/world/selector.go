package world

type BlockSelector interface {
	Query() (string, []any)
}

type BlocksAlongY struct {
	X, Z int
}

func (s BlocksAlongY) Query() (string, []any) {
	return "SELECT posx, posy, posz, data FROM blocks WHERE posx=$1 and posz=$2 ORDER BY posy", []any{
		s.X, s.Z,
	}
}
