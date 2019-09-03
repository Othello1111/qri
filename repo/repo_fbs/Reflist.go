// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package repo_fbs

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type Reflist struct {
	_tab flatbuffers.Table
}

func GetRootAsReflist(buf []byte, offset flatbuffers.UOffsetT) *Reflist {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Reflist{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Reflist) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Reflist) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Reflist) Refs(obj *DatasetRef, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *Reflist) RefsLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func ReflistStart(builder *flatbuffers.Builder) {
	builder.StartObject(1)
}
func ReflistAddRefs(builder *flatbuffers.Builder, refs flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(refs), 0)
}
func ReflistStartRefsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func ReflistEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}