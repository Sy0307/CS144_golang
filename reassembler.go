package reassembler

import (
	"sort"
	"sync"
)

type TCPSegment struct {
	Seq    uint32
	Data   []byte
	Length int
}

type Reassembler struct {
	mutex         sync.Mutex
	segments      []TCPSegment
	expectedSeq   uint32
	byteStream    []byte
	isInitialized bool
}

func NewReassembler(initialSeq uint32) *Reassembler {
	return &Reassembler{
		expectedSeq:   initialSeq,
		segments:      make([]TCPSegment, 0),
		byteStream:    make([]byte, 0),
		isInitialized: true,
	}
}

func (r *Reassembler) AddSegment(seg TCPSegment) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if !r.isInitialized {
		r.expectedSeq = seg.Seq
		r.isInitialized = true
	}
	if seg.Seq+uint32(seg.Length) <= r.expectedSeq {
		return
	}
	r.insertSegment(seg)
}

func (r *Reassembler) insertSegment(seg TCPSegment) {
	idx := sort.Search(len(r.segments), func(i int) bool {
		return r.segments[i].Seq >= seg.Seq
	})
	r.segments = append(r.segments, TCPSegment{})
	copy(r.segments[idx+1:], r.segments[idx:])
	r.segments[idx] = seg

	if idx > 0 {
		m, ok := mergeTwoSegments(r.segments[idx-1], r.segments[idx])
		if ok {
			r.segments[idx-1] = m
			copy(r.segments[idx:], r.segments[idx+1:])
			r.segments = r.segments[:len(r.segments)-1]
			idx--
		}
	}

	for idx < len(r.segments)-1 {
		m, ok := mergeTwoSegments(r.segments[idx], r.segments[idx+1])
		if !ok {
			break
		}
		r.segments[idx] = m
		copy(r.segments[idx+1:], r.segments[idx+2:])
		r.segments = r.segments[:len(r.segments)-1]
	}
}

func mergeTwoSegments(a, b TCPSegment) (TCPSegment, bool) {
	aStart, aEnd := a.Seq, a.Seq+uint32(a.Length)
	bStart, bEnd := b.Seq, b.Seq+uint32(b.Length)
	if bStart > aEnd {
		return TCPSegment{}, false
	}
	start := aStart
	if bStart < start {
		start = bStart
	}
	end := aEnd
	if bEnd > end {
		end = bEnd
	}
	newLen := end - start
	newData := make([]byte, newLen)
	copy(newData[aStart-start:], a.Data)
	copy(newData[bStart-start:], b.Data)
	return TCPSegment{
		Seq:    start,
		Data:   newData,
		Length: int(newLen),
	}, true
}

func (r *Reassembler) MergeSegments() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for len(r.segments) > 0 {
		seg := r.segments[0]
		if seg.Seq > r.expectedSeq {
			break
		}
		start := r.expectedSeq - seg.Seq
		if start >= uint32(len(seg.Data)) {
			r.segments = r.segments[1:]
			continue
		}
		dataToAdd := seg.Data[start:]
		r.byteStream = append(r.byteStream, dataToAdd...)
		r.expectedSeq += uint32(len(dataToAdd))
		r.segments = r.segments[1:]
	}
}

func (r *Reassembler) GetByteStream() []byte {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	return r.byteStream
}
