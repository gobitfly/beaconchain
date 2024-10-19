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
		// loop to complete incomplete epochs
		err := d.handleIncompleteTransfers()
		if err != nil {
			d.log.Error(err, "failed to handle incomplete transfers", 0)
			time.Sleep(10 * time.Second)
			continue
		}
		err = d.handlePendingTransfers()
		if err != nil {
			d.log.Error(err, "failed to handle pending transfers", 0)
			time.Sleep(10 * time.Second)
			continue
		}
		time.Sleep(1 * time.Second)
	}
}

var TransferAtOnce = 2
var TransferInParallel = 3

func (d *dashboardData) handleIncompleteTransfers() error {
	incomplete, err := edb.GetIncompleteTransferEpochs()
	if err != nil {
		return errors.Wrap(err, "failed to get incomplete transfer epochs")
	}
	if len(incomplete) == 0 {
		d.log.Debugf("handleIncompleteTransfers, no incomplete transfer epochs")
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
	pending, err := edb.GetPendingTransferEpochs(int64(TransferAtOnce * TransferInParallel))
	if err != nil {
		return errors.Wrap(err, "failed to get pending transfer epochs")
	}
	if len(pending) == 0 {
		d.log.Debugf("handlePendingTransfers, no pending transfer epochs")
		return nil
	}
	d.log.Infof("handlePendingTransfers, found %d pending transfer epochs", len(pending))

	// allocate transfer batch ids
	batchIds := make([]uuid.UUID, 0)
	for i := range pending {
		if i%TransferAtOnce == 0 {
			id := uuid.New()
			batchIds = append(batchIds, id)
		}
		pending[i].TransferBatchId = &batchIds[len(batchIds)-1]
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

func (d *dashboardData) transferEpochs(epochs []edb.EpochMetadata) error {
	transferBatchEpochs := make(map[uuid.UUID][]edb.EpochMetadata)
	for _, e := range epochs {
		if e.TransferBatchId == nil {
			return fmt.Errorf("transfer batch id is nil for epoch %v", e)
		}
		id := *e.TransferBatchId
		transferBatchEpochs[id] = append(transferBatchEpochs[id], e)
	}
	eg := &errgroup.Group{}
	eg.SetLimit(3)
	for id, epochs := range transferBatchEpochs {
		epochs := epochs
		id := id
		eg.Go(func() error {
			d.log.Infof("doing transfer batch %s", id)
			err := edb.TransferEpochs(epochs)
			if err != nil {
				d.log.Error(err, "failed to transfer epochs", 0, log.Fields{"epochs": epochs})
				return errors.Wrap(err, "failed to transfer epochs")
			}

			now := time.Now()
			for i := range epochs {
				epochs[i].SuccessfulTransfer = &now
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
