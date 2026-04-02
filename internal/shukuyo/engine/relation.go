package engine

// GetRelation determines the relationship between two mansion indices.
// Pure T21n1299 三九秘法: forward distance → position name → paired group.
func GetRelation(idx1, idx2 int) Relation {
	fwd := (idx2 - idx1 + 27) % 27
	direction := SankuPositionNames[fwd]
	group := positionToGroup[direction]
	inverse := DirectionInverse[direction]

	return Relation{
		Group:            group,
		Direction:        direction,
		InverseDirection: inverse,
		ForwardDistance:   fwd,
	}
}
