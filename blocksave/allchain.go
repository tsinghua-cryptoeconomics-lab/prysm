package blocksave

import (
	"encoding/hex"
	"github.com/prysmaticlabs/prysm/v5/consensus-types/interfaces"
	ethpb "github.com/prysmaticlabs/prysm/v5/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v5/time/slots"
	"github.com/sirupsen/logrus"
	"sync"
)

const (
	TotalValidatorCount = 256
	ValidatorPerSlot    = TotalValidatorCount / 32
)

type ChainNode struct {
	forked  map[string]*ChainNode
	root    []byte
	block   interfaces.ReadOnlySignedBeaconBlock
	parent  *ChainNode
	attest  map[string]*ethpb.Attestation
	stabled bool
	mux     sync.Mutex
}

func (n *ChainNode) Block() interfaces.ReadOnlySignedBeaconBlock {
	return n.block
}

func (n *ChainNode) Stabled() bool {
	return n.stabled
}

func (n *ChainNode) UpdateStable() {
	n.mux.Lock()
	defer n.mux.Unlock()
	if len(n.attest) >= ValidatorPerSlot/3 {
		n.stabled = true
	}
}

func changeUnstableToStable(unstable int64) int64 {
	/*
			(16, 3), (17, 3), (18, 4), (19, 5), (20, 6),
		    (21, 7), (22, 8), (23, 9), (24, 10), (25, 10),
		    (26, 11), (27, 12), (28, 13), (29, 14), (30, 16),
		    (31, 17), (32, 18), (33, 19), (34, 20), (35, 21),
		    (36, 22), (37, 23), (38, 24), (39, 26), (40, 27),
		    (41, 28), (42, 29), (43, 30), (44, 32), (45, 33),
		    (46, 34), (47, 35), (48, 37), (49, 38), (50, 39),
		    (51, 40), (52, 42), (53, 43), (54, 44), (55, 45),
		    (56, 47), (57, 48), (58, 49), (59, 51), (60, 52),
		    (61, 53), (62, 55), (63, 56), (64, 57)
	*/
	if unstable < 16 {
		return 0
	}
	switch unstable {
	case 16, 17:
		return 3
	case 18:
		return 4
	case 19:
		return 5
	case 20:
		return 6
	case 21:
		return 7
	case 22:
		return 8
	case 23:
		return 9
	case 24:
		return 10
	case 25:
		return 10
	case 26:
		return 11
	case 27:
		return 12
	case 28:
		return 13
	case 29:
		return 14
	case 30:
		return 16
	case 31:
		return 17
	case 32:
		return 18
	case 33:
		return 19
	case 34:
		return 20
	case 35:
		return 21
	case 36:
		return 22
	case 37:
		return 23
	case 38:
		return 24
	case 39:
		return 26
	case 40:
		return 27
	case 41:
		return 28
	case 42:
		return 29
	case 43:
		return 30
	case 44:
		return 32
	case 45:
		return 33
	case 46:
		return 34
	case 47:
		return 35
	case 48:
		return 37
	case 49:
		return 38
	case 50:
		return 39
	case 51:
		return 40
	case 52:
		return 42
	case 53:
		return 43
	case 54:
		return 44
	case 55:
		return 45
	case 56:
		return 47
	case 57:
		return 48
	case 58:
		return 49
	case 59:
		return 51
	case 60:
		return 52
	case 61:
		return 53
	case 62:
		return 55
	case 63:
		return 56
	case 64:
		return 57
	default:
		return 0
	}
}

func (n *ChainNode) CalcLengthWithStableTransport(finalized int64) int64 {
	unstabledCount := int64(0)
	stabledCount := int64(0)
	n.mux.Lock()
	defer n.mux.Unlock()
	start := n
	for start != nil && int64(start.block.Block().Slot()) > finalized {
		if start.Stabled() {
			stabledCount++
		} else {
			unstabledCount++
		}
		start = start.parent
	}
	return stabledCount + changeUnstableToStable(unstabledCount)
}

func (n *ChainNode) CalcLengthWithoutStableTransport(finalized int64) int64 {
	unstabledCount := int64(0)
	stabledCount := int64(0)
	n.mux.Lock()
	defer n.mux.Unlock()
	start := n
	for start != nil && int64(start.block.Block().Slot()) > finalized {
		if start.Stabled() {
			stabledCount++
		} else {
			unstabledCount++
		}
		start = start.parent
	}
	return stabledCount + unstabledCount
}

type ChainTree struct {
	rootNode       *ChainNode
	blockCache     map[string]*ChainNode // key is block root, and value is ChainNode
	blockSlotCache map[int64]*ChainNode  // key is slot, and value ChainNode
	mux            sync.Mutex
}

func NewChainTree() *ChainTree {
	return &ChainTree{
		rootNode:       nil,
		blockCache:     make(map[string]*ChainNode),
		blockSlotCache: make(map[int64]*ChainNode),
	}
}

func (a *ChainTree) AddBlock(block interfaces.ReadOnlySignedBeaconBlock) *ChainNode {
	if a.rootNode == nil {
		a.mux.Lock()
		defer a.mux.Unlock()
		root, _ := block.Block().HashTreeRoot()
		node := &ChainNode{
			forked: make(map[string]*ChainNode),
			attest: make(map[string]*ethpb.Attestation),
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
				forked: make(map[string]*ChainNode),
				attest: make(map[string]*ethpb.Attestation),
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

func (a *ChainTree) UpdateBlockStatus(attest *ethpb.Attestation) {
	// set block status to stabled if attestation count >= 1/3 validator count
	root := attest.Data.BeaconBlockRoot
	rootStr := hex.EncodeToString(root)
	if _, ok := a.blockCache[rootStr]; ok {
		a.blockCache[rootStr].UpdateStable()
	} else {
		logrus.WithFields(logrus.Fields{
			"root": hex.EncodeToString(root),
			"slot": attest.Data.Slot,
		}).Error("not find block node when update block status")
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

func (a *ChainTree) IteratorAllNode(todo func(node *ChainNode)) {
	a.mux.Lock()
	defer a.mux.Unlock()
	// implement iterator method for all node from root node
	if a.rootNode == nil {
		return
	}
	queue := []*ChainNode{a.rootNode}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		todo(node)
		for _, n := range node.forked {
			queue = append(queue, n)
		}
	}
}

func (a *ChainTree) GetAllLatestBlock() []*ChainNode {
	var res []*ChainNode
	a.IteratorAllNode(func(node *ChainNode) {
		if node.forked == nil || len(node.forked) == 0 {
			res = append(res, node)
		}
	})
	return res
}

func (a *ChainTree) GetAllBlockOnSlot(slot int64) []*ChainNode {
	var res []*ChainNode
	a.IteratorAllNode(func(node *ChainNode) {
		if int64(node.block.Block().Slot()) == slot {
			res = append(res, node)
		}
	})
	return res
}

func (a *ChainTree) GetLongestChainWithStableTransport(checkpoint *ethpb.Checkpoint) *ChainNode {
	allLeaf := a.GetAllLatestBlock()
	finalized, _ := slots.EpochStart(checkpoint.Epoch)

	// slot 周期，从上一个finalized的checkoutpoint开始查找。
	var longest *ChainNode
	for _, node := range allLeaf {
		if longest == nil {
			longest = node
		} else {
			if node.CalcLengthWithStableTransport(int64(finalized)) > longest.CalcLengthWithStableTransport(int64(finalized)) {
				longest = node
			}
		}
	}
	return longest
}

func (a *ChainTree) GetLongestChainWithoutStableTransport(checkpoint *ethpb.Checkpoint) *ChainNode {
	// unstable block 不进行转换
	allLeaf := a.GetAllLatestBlock()
	finalized, _ := slots.EpochStart(checkpoint.Epoch)

	// slot 周期，从上一个finalized的checkoutpoint开始查找。
	var longest *ChainNode
	for _, node := range allLeaf {
		if longest == nil {
			longest = node
		} else {
			if node.CalcLengthWithoutStableTransport(int64(finalized)) > longest.CalcLengthWithoutStableTransport(int64(finalized)) {
				longest = node
			}
		}
	}
	return longest
}

func (a *ChainTree) FilterLatestBlock(slot int64, checkpoint *ethpb.Checkpoint) *ChainNode {
	// 查找 slot 内的所有投票，达到stable的block，
	allBlock := a.GetAllBlockOnSlot(slot)
	stabled := make([]*ChainNode, 0)
	unstabled := make([]*ChainNode, 0)
	for _, node := range allBlock {
		if node.Stabled() {
			stabled = append(stabled, node)
		} else {
			unstabled = append(unstabled, node)
		}
	}

	// exist stabled block
	if len(stabled) > 0 {
		if len(stabled) == 1 {
			return stabled[0]
		} else {
			// choose the longest chain.
			return a.getLongestChainWithStableTransport(stabled, checkpoint)
		}
	}
	// 如果，没有，则选择一个最长链的block.
	// otherwise, choose longest chain head.
	if len(unstabled) > 0 {
		if len(unstabled) == 1 {
			return unstabled[0]
		} else {
			// choose the longest chain.
			return a.getLongestChainWithoutStableTransport(unstabled, checkpoint)
		}
	}
	return nil
}

func (a *ChainTree) getLongestChainWithStableTransport(nodes []*ChainNode, checkpoint *ethpb.Checkpoint) *ChainNode {
	finalized, _ := slots.EpochStart(checkpoint.Epoch)
	var longest *ChainNode
	for _, node := range nodes {
		if longest == nil {
			longest = node
		} else {
			if node.CalcLengthWithStableTransport(int64(finalized)) > longest.CalcLengthWithStableTransport(int64(finalized)) {
				longest = node
			}
		}
	}
	return longest
}

func (a *ChainTree) getLongestChainWithoutStableTransport(nodes []*ChainNode, checkpoint *ethpb.Checkpoint) *ChainNode {
	finalized, _ := slots.EpochStart(checkpoint.Epoch)
	var longest *ChainNode
	for _, node := range nodes {
		if longest == nil {
			longest = node
		} else {
			if node.CalcLengthWithoutStableTransport(int64(finalized)) > longest.CalcLengthWithoutStableTransport(int64(finalized)) {
				longest = node
			}
		}
	}
	return longest
}
