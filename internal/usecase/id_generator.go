package usecase

import (
	"id-generator/internal/domain"
	"time"
)

type IDUsecase struct {
	gen domain.IDGenerator
}

func NewIDUsecase(gen domain.IDGenerator) *IDUsecase {
	return &IDUsecase{gen: gen}
}

func (uc *IDUsecase) GenerateID() (int64, error) {
	return uc.gen.NextID()
}

func (uc *IDUsecase) GenerateIDs(count int) (*[]int64, int64, error) {

	ids := make([]int64, count)
	t := time.Now()

	for i := range count {
		id, err := uc.gen.NextID()

		if err != nil {
			return nil, 0, err
		}
		ids[i] = id
	}

	duration := time.Since(t)

	return &ids, duration.Microseconds(), nil
}
