package blocksave

import (
	"encoding/hex"
	"github.com/prysmaticlabs/prysm/v5/cache/lru"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/interfaces"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/sirupsen/logrus"
)

var (
	blockRootCache = lru.New(100000)      // key is root, and value is block
	blockCache     = lru.New(100000)      // key is slot, and value is block
	attestCache    = lru.New(100000 * 10) // key is root, and value is blockAtt
)

type BlockAtt struct {
	attestations map[string]*ethpb.Attestation //
}

type BlockInfo struct {
	root         string
	attestations map[string]*ethpb.Attestation
	blockdata    interface{}
}

func ReceiveBlock(block interfaces.ReadOnlySignedBeaconBlock) {
	root, _ := block.Block().HashTreeRoot()
	blockRootCache.Add(hex.EncodeToString(root[:]), block)
	blockCache.Add(block.Block().Slot(), block)
	logrus.WithFields(logrus.Fields{
		"root": hex.EncodeToString(root[:]),
		"slot": block.Block().Slot(),
	}).Info("block save receive block")
}

func ReceiveAttestation(attest *ethpb.Attestation) {
	root := attest.Data.BeaconBlockRoot
	if _, ok := attestCache.Get(hex.EncodeToString(root)); !ok {
		attestCache.Add(hex.EncodeToString(root), &BlockAtt{attestations: make(map[string]*ethpb.Attestation)})
	}
	blockAtt, _ := attestCache.Get(hex.EncodeToString(root))
	blockAtt.(*BlockAtt).attestations[hex.EncodeToString(attest.Signature)] = attest
	logrus.WithFields(logrus.Fields{
		"root": hex.EncodeToString(root),
		"slot": attest.Data.Slot,
	}).Info("block save receive attestation")
}

// FetchHeadBlock fetches the head block from the block cache.
func CheckBlock(slot uint64) bool {
	curSlot := slot

	if _, ok := blockCache.Get(slot); ok {
		return true
	}
	return false
}
