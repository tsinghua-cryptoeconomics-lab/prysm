package blocksave

import (
	"encoding/hex"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/interfaces"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/sirupsen/logrus"
	"sync"
)

type ChainNode struct {
	forked map[string]*ChainNode
	root   []byte
	block  interfaces.ReadOnlySignedBeaconBlock
	parent *ChainNode
	attest map[string]*ethpb.Attestation
	mux    sync.Mutex
}

type ChainTree struct {
	rootNode       *ChainNode
	blockCache     map[string]*ChainNode // key is block root, and value is ChainNode
	blockSlotCache map[int64]*ChainNode  // key is slot, and value ChainNode
	mux            sync.Mutex
}

func NewChainTree() *ChainTree {
	return &ChainTree{
		rootNode:   nil,
		blockCache: make(map[string]*ChainNode),
	}
}

func (a *ChainTree) AddBlock(block interfaces.ReadOnlySignedBeaconBlock) *ChainNode {
	if a.rootNode == nil {
		a.mux.Lock()
		defer a.mux.Unlock()
		root, _ := block.Block().HashTreeRoot()
		node := &ChainNode{
			root:   root[:],
			block:  block,
			parent: nil,
		}
		a.rootNode = node
		a.blockCache[hex.EncodeToString(root[:])] = node
		a.blockSlotCache[int64(block.Block().Slot())] = node
		return node
	} else {
		a.mux.Lock()
		defer a.mux.Unlock()
		root, _ := block.Block().HashTreeRoot()
		parentRoot := block.Block().ParentRoot()
		parentStr := hex.EncodeToString(parentRoot[:])
		if node, exist := a.blockCache[parentStr]; exist {
			newNode := &ChainNode{
				root:   root[:],
				block:  block,
				parent: node,
			}
			node.forked[hex.EncodeToString(root[:])] = newNode
			a.blockCache[hex.EncodeToString(root[:])] = newNode
			a.blockSlotCache[int64(block.Block().Slot())] = newNode
			return newNode
		} else {
			logrus.WithFields(logrus.Fields{
				"root":   hex.EncodeToString(root[:]),
				"slot":   block.Block().Slot(),
				"parent": block.Block().ParentRoot(),
			}).Error("not find parent node when add block")
			return nil
		}
	}
}

func (a *ChainTree) AddAttestation(attest *ethpb.Attestation) {
	root := attest.Data.BeaconBlockRoot
	rootStr := hex.EncodeToString(root)
	if _, ok := a.blockCache[rootStr]; ok {
		if _, ok := a.blockCache[rootStr].attest[hex.EncodeToString(attest.Signature)]; !ok {
			a.blockCache[rootStr].attest[hex.EncodeToString(attest.Signature)] = attest
		}
	} else {
		logrus.WithFields(logrus.Fields{
			"root": hex.EncodeToString(root),
			"slot": attest.Data.Slot,
		}).Error("not find block node when add attestation")
	}
}

func (a *ChainTree) GetBlockByRoot(root string) *ChainNode {
	if node, ok := a.blockCache[root]; ok {
		return node
	}
	return nil
}

func (a *ChainTree) GetBlockBySlot(slot int64) *ChainNode {
	if node, ok := a.blockSlotCache[slot]; ok {
		return node
	}
	return nil
}
