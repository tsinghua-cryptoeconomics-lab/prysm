package blocksave

import (
	"encoding/hex"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/interfaces"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v5/time/slots"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	chainTree = NewChainTree()
)

func ReceiveBlock(genesisTime time.Time, block interfaces.ReadOnlySignedBeaconBlock) {
	root, _ := block.Block().HashTreeRoot()
	parentRoot := block.Block().ParentRoot()

	logrus.WithFields(logrus.Fields{
		"root":   hex.EncodeToString(root[:]),
		"slot":   block.Block().Slot(),
		"parent": hex.EncodeToString(parentRoot[:]),
	}).Info("block save receive block")
	chainTree.AddBlock(block)
}

func ReceiveAttestation(genesisTime time.Time, attest *ethpb.Attestation) {
	curSlot := slots.CurrentSlot(uint64(genesisTime.Unix()))
	if attest.Data.Slot != curSlot {
		logrus.WithFields(logrus.Fields{
			"attest_slot": attest.Data.Slot,
			"cur_slot":    curSlot,
		}).Warn("ignore attestation slot because it is not equal to current slot")
		return
	}
	chainTree.AddAttestation(attest)
	// update block status (stabled or unstabled)
	chainTree.UpdateBlockStatus(attest)
}

// QueryBlockStatus query block status (stabled or unstabled)
func QueryBlockStatus(slot int64) bool {
	node := chainTree.GetBlockBySlot(slot)
	return node.stabled
}

func GetLatestHead(slot int64, checkpoint *ethpb.Checkpoint) *ChainNode {
	return chainTree.FilterLatestBlock(slot, checkpoint)
}

func GetLongestChainWithStableTransport(checkpoint *ethpb.Checkpoint) *ChainNode {
	return chainTree.GetLongestChainWithStableTransport(checkpoint)
}

func GetLongestChain(checkpoint *ethpb.Checkpoint) *ChainNode {
	return chainTree.GetLongestChainWithStableTransport(checkpoint)
}
