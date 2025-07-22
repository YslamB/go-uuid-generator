package domain

type IDGenerator interface {
	NextID() (int64, error)
}
