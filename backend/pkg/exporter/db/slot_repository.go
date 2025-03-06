package db

import "github.com/jmoiron/sqlx"

type SlotRepository interface {
	GetAllSlots(tx *sqlx.Tx) ([]uint64, error)
	GetLastSlot(tx *sqlx.Tx) (uint64, error)
	SetSlotFinalizationAndStatus(slot uint64, finalized bool, status string, tx *sqlx.Tx) error
	GetAllNonFinalizedSlots() ([]*GetAllNonFinalizedSlotsRow, error)
}

type slotRepository struct{}

func NewSlotRepository() SlotRepository {
	return &slotRepository{}
}

func (r *slotRepository) GetAllSlots(tx *sqlx.Tx) ([]uint64, error) {
	return GetAllSlots(tx)
}

func (r *slotRepository) GetLastSlot(tx *sqlx.Tx) (uint64, error) {
	return GetLastSlot(tx)
}

func (r *slotRepository) SetSlotFinalizationAndStatus(slot uint64, finalized bool, status string, tx *sqlx.Tx) error {
	return SetSlotFinalizationAndStatus(slot, finalized, status, tx)
}

func (r *slotRepository) GetAllNonFinalizedSlots() ([]*GetAllNonFinalizedSlotsRow, error) {
	return GetAllNonFinalizedSlots()
}
