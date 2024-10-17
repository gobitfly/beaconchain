package modules

import (
	"fmt"
	"time"

	"github.com/gobitfly/beaconchain/pkg/commons/log"
	edb "github.com/gobitfly/beaconchain/pkg/exporter/db"
	"github.com/google/uuid"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// | handleIncompleteTransfers
// | - GetIncompleteTransferEpochs
// | - TransferEpochs  A
// | - PushEpochMetadata (successful_transfer)  B
// | handlePendingTransfers
// | - GetPendingTransferEpochs
// | - allocate transfer batch ids
// | - PushEpochMetadata (transfer_batch_id)
// | - TransferEpochs  A
// | - PushEpochMetadata (successful_transfer)  B

func (d *dashboardData) maintenanceTask() {
	for {
		// wait
		time.Sleep(10 * time.Second)
		// loop to complete incomplete epochs
		err := d.handleIncompleteTransfers()
		if err != nil {
			d.log.Error(err, "failed to handle incomplete transfers", 0)
			continue
		}
		err = d.handlePendingTransfers()
		if err != nil {
			d.log.Error(err, "failed to handle pending transfers", 0)
			continue
		}
	}
}

func (d *dashboardData) handleIncompleteTransfers() error {
	incomplete, err := edb.GetIncompleteTransferEpochs()
	if err != nil {
		return errors.Wrap(err, "failed to get incomplete transfer epochs")
	}
	if len(incomplete) == 0 {
		return nil
	}
	d.log.Infof("handleIncompleteTransfers, found %d incomplete transfer epochs", len(incomplete))

	err = d.transferEpochs(incomplete)
	if err != nil {
		return errors.Wrap(err, "failed to transfer incomplete epochs")
	}
	return nil
}

func (d *dashboardData) handlePendingTransfers() error {
	batchSize := int64(3)       // how many epochs we want to transfer in a single query
	bundleSize := 3 * batchSize // how many epochs we want to transfer in one overall attempt
	pending, err := edb.GetPendingTransferEpochs(bundleSize)
	if err != nil {
		return errors.Wrap(err, "failed to get pending transfer epochs")
	}
	if len(pending) == 0 {
		return nil
	}
	d.log.Infof("handlePendingTransfers, found %d pending transfer epochs", len(pending))

	// allocate transfer batch ids
	var currentBatchId uuid.UUID
	for i, e := range pending {
		if i%int(batchSize) == 0 {
			currentBatchId = uuid.New()
		}
		e.TransferBatchId = &currentBatchId
	}

	err = edb.PushEpochMetadata(pending)
	if err != nil {
		return errors.Wrap(err, "failed to push epoch metadata")
	}

	err = d.transferEpochs(pending)
	if err != nil {
		return errors.Wrap(err, "failed to transfer pending epochs")
	}
	return nil
}

func (d *dashboardData) transferEpochs(incomplete []edb.EpochMetadata) error {
	transferBatchEpochs := make(map[uuid.UUID][]edb.EpochMetadata)
	for _, e := range incomplete {
		if e.TransferBatchId == nil {
			return fmt.Errorf("transfer batch id is nil for epoch %s", e)
		}
		id := *e.TransferBatchId
		transferBatchEpochs[id] = append(transferBatchEpochs[id], e)
	}
	eg := &errgroup.Group{}
	eg.SetLimit(1)
	for id, epochs := range transferBatchEpochs {
		epochs := epochs
		eg.Go(func() error {
			d.log.Infof("doing transfer batch %s", id)
			err := edb.TransferEpochs(epochs)
			if err != nil {
				d.log.Error(err, "failed to transfer epochs", 0, log.Fields{"epochs": epochs})
				return errors.Wrap(err, "failed to transfer epochs")
			}

			now := time.Now()
			for _, e := range epochs {
				e.SuccessfulTransfer = &now
			}
			err = edb.PushEpochMetadata(epochs)
			if err != nil {
				d.log.Error(err, "failed to push epoch metadata", 0, log.Fields{"epochs": epochs})
				return errors.Wrap(err, "failed to push epoch metadata")
			}
			return nil
		})
	}
	err := eg.Wait()
	if err != nil {
		return errors.Wrap(err, "failed to complete transfer batches")
	}
	return nil
}
