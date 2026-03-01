package selector

type BlockSelector interface {
	Query() (string, []any)
}

type BlocksAlongY struct {
	X, Z int
}

func (s BlocksAlongY) Query() (string, []any) {
	return "SELECT x, y, z, data FROM blocks WHERE x=$1 and z=$2 ORDER BY y", []any{
		s.X, s.Z,
	}
}
