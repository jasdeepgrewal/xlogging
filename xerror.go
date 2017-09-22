package xlogging

type stdError struct {
	s string
}

func (e stdError) Error() string {
	return e.s
}
