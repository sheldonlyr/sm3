// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sm3 implements the SM3 hash algorithm as defined in
// https://tools.ietf.org/id/draft-oscca-cfrg-sm3-01.html
package sm3

import "hash"

const (
	Size      = 32
	BlockSize = 64

	chunk = BlockSize
)

var (
	iv = [8]uint32{
		0x7380166f,
		0x4914b2b9,
		0x172442d7,
		0xda8a0600,
		0xa96f30bc,
		0x163138aa,
		0xe38dee4d,
		0xb0fb0e4e,
	}
)

func New() hash.Hash {
	d := new(digest)
	d.Reset()
	return d
}

type digest struct {
	h   [8]uint32
	x   [chunk]byte
	nx  int
	len uint64
}

func (d *digest) Write(p []byte) (nn int, err error) {
	nn = len(p)
	d.len += uint64(nn)
	if d.nx > 0 {
		n := copy(d.x[d.nx:], p)
		d.nx += n
		if d.nx == chunk {
			block(d, d.x[:])
			d.nx = 0
		}
		p = p[n:]
	}

	if len(p) >= chunk {
		n := len(p) &^ (chunk - 1)
		block(d, p[:n])
		p = p[n:]
	}
	if len(p) > 0 {
		d.nx = copy(d.x[:], p)
	}
	return
}

func (d0 *digest) Sum(b []byte) []byte {
	d := *d0
	hash := d.checkSum()
	return append(b, hash[:]...)
}

func (d *digest) Reset() {
	copy(d.h[:], iv[:])
	d.nx = 0
	d.len = 0
}

func (d *digest) Size() int {
	return Size
}

func (d *digest) BlockSize() int {
	return BlockSize
}

func (d *digest) checkSum() [Size]byte {
	len := d.len
	// padding method like crypto/sha1
	var tmp [64]byte
	tmp[0] = 0x80
	if len%64 < 56 {
		d.Write(tmp[0 : 56-len%64])
	} else {
		d.Write(tmp[0 : 64+56-len%64])
	}

	// Length in bits.
	len <<= 3
	putUint64(tmp[:], len)
	d.Write(tmp[0:8])

	if d.nx != 0 {
		panic("d.nx != 0")
	}

	var digest [Size]byte

	putUint32(digest[0:], d.h[0])
	putUint32(digest[4:], d.h[1])
	putUint32(digest[8:], d.h[2])
	putUint32(digest[12:], d.h[3])
	putUint32(digest[16:], d.h[4])
	putUint32(digest[20:], d.h[5])
	putUint32(digest[24:], d.h[6])
	putUint32(digest[28:], d.h[7])

	return digest
}

func putUint64(x []byte, s uint64) {
	_ = x[7]
	x[0] = byte(s >> 56)
	x[1] = byte(s >> 48)
	x[2] = byte(s >> 40)
	x[3] = byte(s >> 32)
	x[4] = byte(s >> 24)
	x[5] = byte(s >> 16)
	x[6] = byte(s >> 8)
	x[7] = byte(s)
}

func putUint32(x []byte, s uint32) {
	_ = x[3]
	x[0] = byte(s >> 24)
	x[1] = byte(s >> 16)
	x[2] = byte(s >> 8)
	x[3] = byte(s)
}

// Sum returns the SM3 checksum of the data.
func Sum(data []byte) [Size]byte {
	var d digest
	d.Reset()
	d.Write(data)
	return d.checkSum()
}