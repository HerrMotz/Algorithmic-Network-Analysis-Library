package graphlib

import "errors"

// ---------------------------
//  Name → index lookup table
// ---------------------------

type LookupTable map[string]int

// InitializeLookupTable builds a mapping from external string identifiers to
// consecutive internal indices [0..n). If externalNames is nil you get an empty
// table of the requested capacity.
func InitializeLookupTable(size int, externalNames []string) (LookupTable, error) {
	if size <= 0 {
		return nil, errors.New("size must be > 0")
	}
	if externalNames != nil && len(externalNames) != size {
		return nil, errors.New("len(externalNames) must equal size (or be nil)")
	}
	tbl := make(LookupTable, size)
	for i := 0; i < size; i++ {
		name := ""
		if externalNames != nil {
			name = externalNames[i]
		} else {
			name = "v" + itoa(i) // fallback name e.g. v0, v1, …
		}
		tbl[name] = i
	}
	return tbl, nil
}

// Cheap int→string without fmt overhead (works for small non‑negative ints).
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for i > 0 {
		buf = append([]byte{byte('0' + i%10)}, buf...)
		i /= 10
	}
	return string(buf)
}
