package events

type Bus struct {
}

type Handler func()

func (b Bus) Register(h Handler) error {
	return nil
}
