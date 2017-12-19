package logtop

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"

	"github.com/timtadh/data-structures/types"
)

type Field interface {
	types.Hashable
}

type Uint64Field struct {
	Val uint64
}

func (k *Uint64Field) Equals(b types.Equatable) bool {
	if b, ok := b.(*Uint64Field); ok {
		return k.Val == b.Val
	}
	panic(fmt.Sprintf("expected *Uint64Field, got: %v", b))
	return false
}
func (k *Uint64Field) Less(b types.Sortable) bool {
	if b, ok := b.(*Uint64Field); ok {
		return k.Val < b.Val
	}
	panic(fmt.Sprintf("expected *Uint64Field, got: %v", b))
	return false
}
func (k *Uint64Field) Hash() int {
	h := fnv.New32a()

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, k.Val)
	h.Write(b)

	return int(h.Sum32())
}

type StringField struct {
	Val string
}

func (k *StringField) Equals(b types.Equatable) bool {
	if b, ok := b.(*StringField); ok {
		return k.Val == b.Val
	}
	panic(fmt.Sprintf("expected *StringField, got: %v", b))
	return false
}
func (k *StringField) Less(b types.Sortable) bool {
	if b, ok := b.(*StringField); ok {
		return k.Val < b.Val
	}
	panic(fmt.Sprintf("expected *StringField, got: %v", b))
	return false
}
func (k *StringField) Hash() int {
	h := fnv.New32a()
	h.Write([]byte(k.Val))
	return int(h.Sum32())
}

type CompoundField struct {
	Fields []Field
}

func (k *CompoundField) Equals(b types.Equatable) bool {
	if b, ok := b.(*CompoundField); ok {
		if len(k.Fields) != len(b.Fields) {
			panic(fmt.Sprintf("expected field of length %v, got %v", len(k.Fields), len(b.Fields)))
		}
		for i, ki := range k.Fields {
			bi := b.Fields[i]
			if !ki.Equals(bi) {
				return false
			}
		}
		return true
	}
	panic(fmt.Sprintf("expected *CompoundField, got: %v", b))
	return false
}
func (k *CompoundField) Less(b types.Sortable) bool {
	if b, ok := b.(*CompoundField); ok {
		if len(k.Fields) != len(b.Fields) {
			panic(fmt.Sprintf("expected field of length %v, got %v", len(k.Fields), len(b.Fields)))
		}
		for i, ki := range k.Fields {
			bi := b.Fields[i]
			if ki.Equals(bi) {
				continue
			}
			return ki.Less(bi)
		}
		return false
	}
	panic(fmt.Sprintf("expected *CompoundField, got: %v", b))
	return false
}
func (k *CompoundField) Hash() int {
	h := 0
	for _, ki := range k.Fields {
		h = h ^ ki.Hash()
	}
	return h
}
