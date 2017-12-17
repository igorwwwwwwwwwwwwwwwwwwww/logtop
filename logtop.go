package logtop

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"sync"

	"github.com/timtadh/data-structures/tree/avl"
	"github.com/timtadh/data-structures/types"
)

type Entry struct {
	Line  string
	Count uint64
}

func (e *Entry) Equals(b types.Equatable) bool {
	if b, ok := b.(*Entry); ok {
		return e.Line == b.Line && e.Count == b.Count
	}
	panic(fmt.Sprintf("expected *Entry, got: %v", b))
	return false
}
func (e *Entry) Less(b types.Sortable) bool {
	if b, ok := b.(*Entry); ok {
		if e.Count == b.Count {
			return types.String(e.Line).Less(types.String(b.Line))
		}
		return e.Count < b.Count
	}
	// what does this even mean?
	// i never thought i was going to say it
	// but this seems like a pretty ... Generic problem
	panic(fmt.Sprintf("expected *Entry, got: %v", b))
	return false
}
func (e *Entry) Hash() int {
	h := fnv.New32a()
	h.Write([]byte(string(e.Line)))

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, e.Count)
	h.Write(b)

	return int(h.Sum32())
}

type TopNTree struct {
	t *avl.AvlTree
	h map[string]*Entry
	m *sync.Mutex
}

func NewTopNTree() *TopNTree {
	return &TopNTree{
		t: &avl.AvlTree{},
		h: make(map[string]*Entry),
		m: &sync.Mutex{},
	}
}

func (top *TopNTree) Increment(line string) error {
	top.m.Lock()
	defer top.m.Unlock()

	e, ok := top.h[line]
	if !ok {
		e = &Entry{
			Line:  line,
			Count: 0,
		}
		top.h[line] = e
		top.t.Put(e, e)
	}

	_, err := top.t.Remove(e)
	if err != nil {
		return err
	}
	e.Count++
	top.t.Put(e, e)

	return nil
}

func (top *TopNTree) TopN(n uint64) []Entry {
	top.m.Lock()
	defer top.m.Unlock()

	es := []Entry{}

	for e, _, next := top.iterate()(); next != nil; e, _, next = next() {
		if n == 0 {
			break
		}
		n--
		if e, ok := e.(*Entry); ok {
			es = append(es, *e)
		}
	}

	return es
}

func (top *TopNTree) iterate() types.KVIterator {
	root := top.t.Root().(*avl.AvlNode)
	tni := TraverseBinaryTreeInReverseOrder(root)
	return types.MakeKVIteratorFromTreeNodeIterator(tni)
}
