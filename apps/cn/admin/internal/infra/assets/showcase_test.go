package assets

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"
)

// Small ISOBMFF metadata fixtures exercise primary-item association and bounds;
// the parser intentionally does not decode/re-encode AV1 pixel payloads.
func showcaseFixture(w, h uint32) []byte {
	box := func(kind string, data []byte) []byte {
		out := make([]byte, len(data)+8)
		binary.BigEndian.PutUint32(out, uint32(len(out)))
		copy(out[4:8], kind)
		copy(out[8:], data)
		return out
	}
	ispe := make([]byte, 12)
	binary.BigEndian.PutUint32(ispe[4:], w)
	binary.BigEndian.PutUint32(ispe[8:], h)
	pitm := box("pitm", []byte{0, 0, 0, 0, 0, 1})
	ipma := box("ipma", []byte{0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 1, 1})
	iprp := box("iprp", append(box("ipco", box("ispe", ispe)), ipma...))
	meta := box("meta", append(append([]byte{0, 0, 0, 0}, pitm...), iprp...))
	return append(box("ftyp", []byte("avif\x00\x00\x00\x00avifmif1")), meta...)
}
func TestShowcaseExactAVIF(t *testing.T) {
	for _, tc := range []struct {
		kind string
		w, h uint32
		ok   bool
	}{{"showcase-desktop", 1600, 800, true}, {"showcase-desktop", 1599, 800, false}, {"showcase-desktop", 1600, 801, false}, {"showcase-mobile", 1200, 675, true}, {"showcase-mobile", 1200, 674, false}} {
		data := showcaseFixture(tc.w, tc.h)
		o, e := NewObject(tc.kind, 119, "input.avif", data)
		if (e == nil) != tc.ok {
			t.Fatal(tc, e)
		}
		if tc.ok && (!bytes.Equal(data, o.Data) || !ValidKey(o.Key)) {
			t.Fatal("bytes or path changed")
		}
	}
	for _, data := range [][]byte{[]byte("\x89PNG\r\n\x1a\n"), []byte("\xff\xd8\xffJPEG"), make([]byte, MaxSize+1), showcaseFixture(1600, 800)[:40]} {
		if _, _, e := Validate("showcase-desktop", "bad.avif", data); e == nil {
			t.Fatal("invalid upload")
		}
	}
	good := showcaseFixture(1600, 800)
	for i := 0; i < len(good); i++ {
		_, _, _ = showcaseAVIFDimensions(good[:i])
	}
	bad := append([]byte{}, good...)
	binary.BigEndian.PutUint32(bad, 0xffffffff)
	if _, _, e := showcaseAVIFDimensions(bad); e == nil {
		t.Fatal("out-of-bounds box accepted")
	}
	bad = append([]byte{}, good...)
	bad[len(bad)-1] = 2
	if _, _, e := showcaseAVIFDimensions(bad); e == nil {
		t.Fatal("unassociated dimensions accepted")
	}
}
func TestShowcasePublicationAndRepair(t *testing.T) {
	p, m := &memoryStore{objects: map[string]Object{}}, &memoryStore{objects: map[string]Object{}}
	s := &Service{p, m}
	object, e := NewObject("showcase-desktop", 119, "image.avif", showcaseFixture(1600, 800))
	if e != nil {
		t.Fatal(e)
	}
	p.fail = true
	if _, e := s.Publish(context.Background(), object); e == nil || m.puts != 0 {
		t.Fatal("Primary failure accepted")
	}
	p.fail = false
	m.fail = true
	result, e := s.Publish(context.Background(), object)
	if e != nil || len(result.Warnings) != 1 || result.Primary != "ready" {
		t.Fatal(result, e)
	}
	m.fail = false
	if _, e := s.Publish(context.Background(), object); e != nil || p.puts != 1 {
		t.Fatal("non-idempotent publication", e)
	}
	delete(m.objects, object.Key)
	if e := s.RepairMirror(context.Background(), object.Key); e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(m.objects[object.Key].Data, object.Data) {
		t.Fatal("repair altered image")
	}
}
