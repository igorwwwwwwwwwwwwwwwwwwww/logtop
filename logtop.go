package logtop

import (
	"time"

	"github.com/timtadh/data-structures/tree"
	"github.com/timtadh/data-structures/tree/avl"
	"github.com/timtadh/data-structures/types"
)

type Tuple struct {
	ID        string
	Count     uint64
	UpdatedAt time.Time

	indexFieldCache map[string]Field
}

func (tup *Tuple) FlushIndexFieldCache(key string) {
	delete(tup.indexFieldCache, key)
}

func (tup *Tuple) IndexedByCount() Field {
	if f, ok := tup.indexFieldCache["Count"]; ok {
		return f
	}

	f := &CompoundField{
		Fields: []Field{
			&Uint64Field{tup.Count},
			&StringField{tup.ID},
		},
	}
	tup.indexFieldCache["Count"] = f
	return f
}

func (tup *Tuple) IndexedByUpdatedAt() Field {
	if f, ok := tup.indexFieldCache["UpdatedAt"]; ok {
		return f
	}

	f := &CompoundField{
		Fields: []Field{
			&Uint64Field{uint64(tup.UpdatedAt.UnixNano())},
			&StringField{tup.ID},
		},
	}
	tup.indexFieldCache["UpdatedAt"] = f
	return f
}

func NewTuple(id string) *Tuple {
	return &Tuple{
		ID:              id,
		Count:           0,
		indexFieldCache: make(map[string]Field),
	}
}

type TopNTree struct {
	countIndex     *avl.AvlTree
	updatedAtIndex *avl.AvlTree
	table          map[string]*Tuple
}

func NewTopNTree() *TopNTree {
	return &TopNTree{
		countIndex:     &avl.AvlTree{},
		updatedAtIndex: &avl.AvlTree{},
		table:          make(map[string]*Tuple),
	}
}

func (top *TopNTree) Increment(id string, updatedAt time.Time) {
	tup, ok := top.table[id]
	if !ok {
		tup = NewTuple(id)
		top.table[id] = tup
	}

	if ok {
		top.countIndex.Remove(tup.IndexedByCount())
		top.updatedAtIndex.Remove(tup.IndexedByUpdatedAt())

		tup.FlushIndexFieldCache("Count")
		tup.FlushIndexFieldCache("UpdatedAt")
	}

	tup.Count++
	tup.UpdatedAt = updatedAt

	top.countIndex.Put(tup.IndexedByCount(), tup)
	top.updatedAtIndex.Put(tup.IndexedByUpdatedAt(), tup)
}

func (top *TopNTree) TopN(n uint64) []Tuple {
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
	for _, tup, next := top.iterateByUpdatedAtAsc()(); next != nil; _, tup, next = next() {
		tup := tup.(*Tuple)
		if before.Before(tup.UpdatedAt) {
			break
		}
		top.countIndex.Remove(tup.IndexedByCount())
		top.updatedAtIndex.Remove(tup.IndexedByUpdatedAt())
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
