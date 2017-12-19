package logtop

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/timtadh/data-structures/tree"
	"github.com/timtadh/data-structures/tree/avl"
	"github.com/timtadh/data-structures/types"
)

type Tuple struct {
	ID        string
	Count     uint64
	UpdatedAt time.Time
}

type IndexedByCount struct {
	Tup *Tuple
}

func (k *IndexedByCount) Equals(b types.Equatable) bool {
	if b, ok := b.(*IndexedByCount); ok {
		return k.Tup.ID == b.Tup.ID && k.Tup.Count == b.Tup.Count
	}
	panic(fmt.Sprintf("expected *IndexedByCount, got: %v", b))
	return false
}
func (k *IndexedByCount) Less(b types.Sortable) bool {
	if b, ok := b.(*IndexedByCount); ok {
		if k.Tup.Count == b.Tup.Count {
			return types.String(k.Tup.ID).Less(types.String(b.Tup.ID))
		}
		return k.Tup.Count < b.Tup.Count
	}
	panic(fmt.Sprintf("expected *IndexedByCount, got: %v", b))
	return false
}
func (k *IndexedByCount) Hash() int {
	h := fnv.New32a()
	h.Write([]byte(string(k.Tup.ID)))

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, k.Tup.Count)
	h.Write(b)

	return int(h.Sum32())
}

type IndexedByUpdatedAt struct {
	Tup *Tuple
}

func (k *IndexedByUpdatedAt) Equals(b types.Equatable) bool {
	if b, ok := b.(*IndexedByUpdatedAt); ok {
		return k.Tup.ID == b.Tup.ID && k.Tup.UpdatedAt == b.Tup.UpdatedAt
	}
	panic(fmt.Sprintf("expected *IndexedByUpdatedAt, got: %v", b))
	return false
}
func (k *IndexedByUpdatedAt) Less(b types.Sortable) bool {
	if b, ok := b.(*IndexedByUpdatedAt); ok {
		if k.Tup.UpdatedAt == b.Tup.UpdatedAt {
			return types.String(k.Tup.ID).Less(types.String(b.Tup.ID))
		}
		return k.Tup.UpdatedAt.Before(b.Tup.UpdatedAt)
	}
	panic(fmt.Sprintf("expected *IndexedByUpdatedAt, got: %v", b))
	return false
}
func (k *IndexedByUpdatedAt) Hash() int {
	h := fnv.New32a()
	h.Write([]byte(string(k.Tup.ID)))

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(k.Tup.UpdatedAt.UnixNano()))
	h.Write(b)

	return int(h.Sum32())
}

type TopNTree struct {
	countIndex     *avl.AvlTree
	updatedAtIndex *avl.AvlTree
	table          map[string]*Tuple
	m              *sync.Mutex
}

func NewTopNTree() *TopNTree {
	return &TopNTree{
		countIndex:     &avl.AvlTree{},
		updatedAtIndex: &avl.AvlTree{},
		table:          make(map[string]*Tuple),
		m:              &sync.Mutex{},
	}
}

func (top *TopNTree) Increment(id string) error {
	top.m.Lock()
	defer top.m.Unlock()

	tup, ok := top.table[id]
	if !ok {
		tup = &Tuple{
			ID:    id,
			Count: 0,
		}
		top.table[id] = tup
	}

	k1 := &IndexedByCount{Tup: tup}
	k2 := &IndexedByUpdatedAt{Tup: tup}

	if ok {
		_, err := top.countIndex.Remove(k1)
		if err != nil {
			return err
		}
		_, err = top.updatedAtIndex.Remove(k2)
		if err != nil {
			return err
		}
	}

	tup.Count++
	tup.UpdatedAt = time.Now()

	top.countIndex.Put(k1, tup)
	top.updatedAtIndex.Put(k2, tup)

	return nil
}

func (top *TopNTree) TopN(n uint64) []Tuple {
	top.m.Lock()
	defer top.m.Unlock()

	tups := []Tuple{}

	for _, tup, next := top.iterateByCountDesc()(); next != nil; _, tup, next = next() {
		if n == 0 {
			break
		}
		n--
		if tup, ok := tup.(*Tuple); ok {
			tups = append(tups, *tup)
		}
	}

	return tups
}

func (top *TopNTree) PruneBefore(before time.Time) {
	top.m.Lock()
	defer top.m.Unlock()

	for _, tup, next := top.iterateByUpdatedAtAsc()(); next != nil; _, tup, next = next() {
		tup := tup.(*Tuple)
		if before.Before(tup.UpdatedAt) {
			break
		}
		top.countIndex.Remove(&IndexedByCount{Tup: tup})
		top.updatedAtIndex.Remove(&IndexedByUpdatedAt{Tup: tup})
		delete(top.table, tup.ID)
	}
}

func (top *TopNTree) iterateByCountDesc() types.KVIterator {
	root := top.countIndex.Root().(*avl.AvlNode)
	tni := TraverseBinaryTreeInReverseOrder(root)
	return types.MakeKVIteratorFromTreeNodeIterator(tni)
}

func (top *TopNTree) iterateByUpdatedAtAsc() types.KVIterator {
	root := top.countIndex.Root().(*avl.AvlNode)
	tni := tree.TraverseBinaryTreeInOrder(root)
	return types.MakeKVIteratorFromTreeNodeIterator(tni)
}
