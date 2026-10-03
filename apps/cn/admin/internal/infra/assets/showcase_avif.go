package assets

import (
	"encoding/binary"
	"errors"
)

// Inspect the primary item's associated spatial property, never the first ispe
// byte sequence (which may belong to a thumbnail or an alpha plane). See
// https://aomediacodec.github.io/av1-avif/v1.2.0.html#image-spatial-extents-property
// This bounded metadata reader preserves bytes and does not decode or transform pixels.
type avifBox struct {
	kind string
	data []byte
}

var errShowcaseAVIF = errors.New("invalid or unsupported Showcase AVIF metadata")

func avifBoxes(data []byte) ([]avifBox, error) {
	boxes := make([]avifBox, 0)
	for len(data) > 0 {
		if len(data) < 8 || len(boxes) >= 4096 {
			return nil, errShowcaseAVIF
		}
		n, header := uint64(binary.BigEndian.Uint32(data)), uint64(8)
		if n == 1 {
			if len(data) < 16 {
				return nil, errShowcaseAVIF
			}
			n = binary.BigEndian.Uint64(data[8:])
			header = 16
		}
		if n == 0 {
			n = uint64(len(data))
		}
		if n < header || n > uint64(len(data)) {
			return nil, errShowcaseAVIF
		}
		boxes = append(boxes, avifBox{string(data[4:8]), data[header:n]})
		data = data[n:]
	}
	return boxes, nil
}

func oneAVIFBox(boxes []avifBox, kind string) ([]byte, error) {
	var data []byte
	for _, box := range boxes {
		if box.kind == kind {
			if data != nil {
				return nil, errShowcaseAVIF
			}
			data = box.data
		}
	}
	if data == nil {
		return nil, errShowcaseAVIF
	}
	return data, nil
}

func showcaseAVIFDimensions(data []byte) (uint32, uint32, error) {
	if len(data) == 0 || len(data) > MaxSize {
		return 0, 0, errShowcaseAVIF
	}
	top, e := avifBoxes(data)
	if e != nil {
		return 0, 0, e
	}
	f, e := oneAVIFBox(top, "ftyp")
	if e != nil || len(f) < 8 || len(f)%4 != 0 {
		return 0, 0, errShowcaseAVIF
	}
	brand := string(f[:4]) == "avif"
	for i := 8; i < len(f); i += 4 {
		brand = brand || string(f[i:i+4]) == "avif"
	}
	if !brand {
		return 0, 0, errShowcaseAVIF
	}
	meta, e := oneAVIFBox(top, "meta")
	if e != nil || len(meta) < 4 || binary.BigEndian.Uint32(meta) != 0 {
		return 0, 0, errShowcaseAVIF
	}
	boxes, e := avifBoxes(meta[4:])
	if e != nil {
		return 0, 0, e
	}
	pitm, e := oneAVIFBox(boxes, "pitm")
	if e != nil || len(pitm) < 6 || pitm[1] != 0 || pitm[2] != 0 || pitm[3] != 0 {
		return 0, 0, errShowcaseAVIF
	}
	var id uint32
	switch pitm[0] {
	case 0:
		if len(pitm) != 6 {
			return 0, 0, errShowcaseAVIF
		}
		id = uint32(binary.BigEndian.Uint16(pitm[4:]))
	case 1:
		if len(pitm) != 8 {
			return 0, 0, errShowcaseAVIF
		}
		id = binary.BigEndian.Uint32(pitm[4:])
	default:
		return 0, 0, errShowcaseAVIF
	}
	iprp, e := oneAVIFBox(boxes, "iprp")
	if e != nil {
		return 0, 0, e
	}
	properties, e := avifBoxes(iprp)
	if e != nil {
		return 0, 0, e
	}
	ipco, e := oneAVIFBox(properties, "ipco")
	if e != nil {
		return 0, 0, e
	}
	props, e := avifBoxes(ipco)
	if e != nil {
		return 0, 0, e
	}
	ipma, e := oneAVIFBox(properties, "ipma")
	if e != nil || len(ipma) < 8 || ipma[0] > 1 || ipma[1] != 0 || ipma[2] != 0 || ipma[3] > 1 {
		return 0, 0, errShowcaseAVIF
	}
	wide, version := ipma[3] == 1, ipma[0]
	count := binary.BigEndian.Uint32(ipma[4:])
	ipma = ipma[8:]
	if count > 4096 {
		return 0, 0, errShowcaseAVIF
	}
	var width, height uint32
	spatialSeen := false
	seen := false
	for entry := uint32(0); entry < count; entry++ {
		n := 2
		if version == 1 {
			n = 4
		}
		if len(ipma) < n+1 {
			return 0, 0, errShowcaseAVIF
		}
		item := uint32(binary.BigEndian.Uint16(ipma))
		if version == 1 {
			item = binary.BigEndian.Uint32(ipma)
		}
		associations := int(ipma[n])
		ipma = ipma[n+1:]
		if item == id {
			if seen {
				return 0, 0, errShowcaseAVIF
			}
			seen = true
		}
		for j := 0; j < associations; j++ {
			size := 1
			if wide {
				size = 2
			}
			if len(ipma) < size {
				return 0, 0, errShowcaseAVIF
			}
			index := int(ipma[0] & 127)
			if wide {
				index = int(binary.BigEndian.Uint16(ipma) & 32767)
			}
			ipma = ipma[size:]
			if index > len(props) {
				return 0, 0, errShowcaseAVIF
			}
			if item != id || index == 0 {
				continue
			}
			p := props[index-1]
			switch p.kind {
			case "ispe":
				if spatialSeen || len(p.data) != 12 || binary.BigEndian.Uint32(p.data) != 0 {
					return 0, 0, errShowcaseAVIF
				}
				spatialSeen = true
				width = binary.BigEndian.Uint32(p.data[4:])
				height = binary.BigEndian.Uint32(p.data[8:])
			case "clap", "irot":
				// Transformative geometry would make encoded and displayed dimensions differ.
				return 0, 0, errShowcaseAVIF
			}
		}
	}
	if len(ipma) != 0 || !seen || width == 0 || height == 0 {
		return 0, 0, errShowcaseAVIF
	}
	return width, height, nil
}
