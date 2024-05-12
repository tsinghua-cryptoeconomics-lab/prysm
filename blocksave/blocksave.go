package blocksave

import (
	"encoding/hex"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/interfaces"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/sirupsen/logrus"
)

var (
	chainTree = NewChainTree()
)

func ReceiveBlock(block interfaces.ReadOnlySignedBeaconBlock) {
	root, _ := block.Block().HashTreeRoot()
	parentRoot := block.Block().ParentRoot()

	logrus.WithFields(logrus.Fields{
		"root":   hex.EncodeToString(root[:]),
		"slot":   block.Block().Slot(),
		"parent": hex.EncodeToString(parentRoot[:]),
	}).Info("block save receive block")
	chainTree.AddBlock(block)
}

func ReceiveAttestation(attest *ethpb.Attestation) {
	chainTree.AddAttestation(attest)
	// todo: update block status (stabled or unstabled)
}

// todo: add api to query block status (stabled or unstabled)
func QueryBlockStatus(slot int64) byte {
	node := chainTree.GetBlockBySlot(slot)
	custom := node.block.Block().Body().Graffiti()
	return custom[0]
}
