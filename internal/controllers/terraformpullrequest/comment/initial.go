package comment

type InitialComment struct {
}

func NewInitialComment() *InitialComment {
	return &InitialComment{}
}

func (c *InitialComment) Marker() string { return Marker("") }
