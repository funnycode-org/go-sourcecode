// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runtime

import (
	"runtime/internal/sys"
	"unsafe"
)

func mapaccess1_faststr(t *maptype, h *hmap, ky string) unsafe.Pointer {
	if raceenabled && h != nil {
		callerpc := getcallerpc()
		racereadpc(unsafe.Pointer(h), callerpc, funcPC(mapaccess1_faststr))
	}
	if h == nil || h.count == 0 {
		return unsafe.Pointer(&zeroVal[0])
	}
	if h.flags&hashWriting != 0 {
		throw("concurrent map read and map write")
	}
	key := stringStructOf(&ky)
	if h.B == 0 {
		// One-bucket table.
		b := (*bmap)(h.buckets)
		if key.len < 32 {
			// short key, doing lots of comparisons is ok
			for i, kptr := uintptr(0), b.keys(); i < bucketCnt; i, kptr = i+1, add(kptr, 2*sys.PtrSize) {
				k := (*stringStruct)(kptr)
				if k.len != key.len || isEmpty(b.tophash[i]) {
					if b.tophash[i] == emptyRest {
						break
					}
					continue
				}
				if k.str == key.str || memequal(k.str, key.str, uintptr(key.len)) {
					return add(unsafe.Pointer(b), dataOffset+bucketCnt*2*sys.PtrSize+i*uintptr(t.elemsize))
				}
			}
			return unsafe.Pointer(&zeroVal[0])
		}
		// long key, try not to do more comparisons than necessary
		keymaybe := uintptr(bucketCnt)
		for i, kptr := uintptr(0), b.keys(); i < bucketCnt; i, kptr = i+1, add(kptr, 2*sys.PtrSize) {
			k := (*stringStruct)(kptr)
			if k.len != key.len || isEmpty(b.tophash[i]) {
				if b.tophash[i] == emptyRest {
					break
				}
				continue
			}
			if k.str == key.str {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*2*sys.PtrSize+i*uintptr(t.elemsize))
			}
			// check first 4 bytes
			if *((*[4]byte)(key.str)) != *((*[4]byte)(k.str)) {
				continue
			}
			// check last 4 bytes
			if *((*[4]byte)(add(key.str, uintptr(key.len)-4))) != *((*[4]byte)(add(k.str, uintptr(key.len)-4))) {
				continue
			}
			if keymaybe != bucketCnt {
				// Two keys are potential matches. Use hash to distinguish them.
				goto dohash
			}
			keymaybe = i
		}
		if keymaybe != bucketCnt {
			k := (*stringStruct)(add(unsafe.Pointer(b), dataOffset+keymaybe*2*sys.PtrSize))
			if memequal(k.str, key.str, uintptr(key.len)) {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*2*sys.PtrSize+keymaybe*uintptr(t.elemsize))
			}
		}
		return unsafe.Pointer(&zeroVal[0])
	}
dohash:
	hash := t.hasher(noescape(unsafe.Pointer(&ky)), uintptr(h.hash0))
	m := bucketMask(h.B)
	b := (*bmap)(add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
	if c := h.oldbuckets; c != nil {
		if !h.sameSizeGrow() {
			// There used to be half as many buckets; mask down one more power of two.
			m >>= 1
		}
		oldb := (*bmap)(add(c, (hash&m)*uintptr(t.bucketsize)))
		if !evacuated(oldb) {
			b = oldb
		}
	}
	top := tophash(hash)
	for ; b != nil; b = b.overflow(t) {
		for i, kptr := uintptr(0), b.keys(); i < bucketCnt; i, kptr = i+1, add(kptr, 2*sys.PtrSize) {
			k := (*stringStruct)(kptr)
			if k.len != key.len || b.tophash[i] != top {
				continue
			}
			if k.str == key.str || memequal(k.str, key.str, uintptr(key.len)) {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*2*sys.PtrSize+i*uintptr(t.elemsize))
			}
		}
	}
	return unsafe.Pointer(&zeroVal[0])
}

func mapaccess2_faststr(t *maptype, h *hmap, ky string) (unsafe.Pointer, bool) {
	if raceenabled && h != nil {
		callerpc := getcallerpc()
		racereadpc(unsafe.Pointer(h), callerpc, funcPC(mapaccess2_faststr))
	}
	if h == nil || h.count == 0 {
		return unsafe.Pointer(&zeroVal[0]), false
	}
	if h.flags&hashWriting != 0 {
		throw("concurrent map read and map write")
	}
	key := stringStructOf(&ky)
	if h.B == 0 {
		// One-bucket table.
		b := (*bmap)(h.buckets)
		if key.len < 32 {
			// short key, doing lots of comparisons is ok
			for i, kptr := uintptr(0), b.keys(); i < bucketCnt; i, kptr = i+1, add(kptr, 2*sys.PtrSize) {
				k := (*stringStruct)(kptr)
				if k.len != key.len || isEmpty(b.tophash[i]) {
					if b.tophash[i] == emptyRest {
						break
					}
					continue
				}
				if k.str == key.str || memequal(k.str, key.str, uintptr(key.len)) {
					return add(unsafe.Pointer(b), dataOffset+bucketCnt*2*sys.PtrSize+i*uintptr(t.elemsize)), true
				}
			}
			return unsafe.Pointer(&zeroVal[0]), false
		}
		// long key, try not to do more comparisons than necessary
		keymaybe := uintptr(bucketCnt)
		for i, kptr := uintptr(0), b.keys(); i < bucketCnt; i, kptr = i+1, add(kptr, 2*sys.PtrSize) {
			k := (*stringStruct)(kptr)
			if k.len != key.len || isEmpty(b.tophash[i]) {
				if b.tophash[i] == emptyRest {
					break
				}
				continue
			}
			if k.str == key.str {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*2*sys.PtrSize+i*uintptr(t.elemsize)), true
			}
			// check first 4 bytes
			if *((*[4]byte)(key.str)) != *((*[4]byte)(k.str)) {
				continue
			}
			// check last 4 bytes
			if *((*[4]byte)(add(key.str, uintptr(key.len)-4))) != *((*[4]byte)(add(k.str, uintptr(key.len)-4))) {
				continue
			}
			if keymaybe != bucketCnt {
				// Two keys are potential matches. Use hash to distinguish them.
				goto dohash
			}
			keymaybe = i
		}
		if keymaybe != bucketCnt {
			k := (*stringStruct)(add(unsafe.Pointer(b), dataOffset+keymaybe*2*sys.PtrSize))
			if memequal(k.str, key.str, uintptr(key.len)) {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*2*sys.PtrSize+keymaybe*uintptr(t.elemsize)), true
			}
		}
		return unsafe.Pointer(&zeroVal[0]), false
	}
dohash:
	hash := t.hasher(noescape(unsafe.Pointer(&ky)), uintptr(h.hash0))
	m := bucketMask(h.B)
	b := (*bmap)(add(h.buckets, (hash&m)*uintptr(t.bucketsize)))
	if c := h.oldbuckets; c != nil {
		if !h.sameSizeGrow() {
			// There used to be half as many buckets; mask down one more power of two.
			m >>= 1
		}
		oldb := (*bmap)(add(c, (hash&m)*uintptr(t.bucketsize)))
		if !evacuated(oldb) {
			b = oldb
		}
	}
	top := tophash(hash)
	for ; b != nil; b = b.overflow(t) {
		for i, kptr := uintptr(0), b.keys(); i < bucketCnt; i, kptr = i+1, add(kptr, 2*sys.PtrSize) {
			k := (*stringStruct)(kptr)
			if k.len != key.len || b.tophash[i] != top {
				continue
			}
			if k.str == key.str || memequal(k.str, key.str, uintptr(key.len)) {
				return add(unsafe.Pointer(b), dataOffset+bucketCnt*2*sys.PtrSize+i*uintptr(t.elemsize)), true
			}
		}
	}
	return unsafe.Pointer(&zeroVal[0]), false
}

func mapassign_faststr(t *maptype, h *hmap, s string) unsafe.Pointer {
	if h == nil {
		panic(plainError("assignment to entry in nil map"))
	}
	if raceenabled {
		callerpc := getcallerpc()
		racewritepc(unsafe.Pointer(h), callerpc, funcPC(mapassign_faststr))
	}
	//d := (*dmap)(unsafe.Pointer(uintptr(h.buckets)))
	//bucketD := uintptr(0)
	//for bucketD < bucketShift(h.B)+3 {
	//	flag := false
	//	for _, debugKey := range d.debugKeys {
	//		if debugKey == "" {
	//			continue
	//		}
	//		if flag == false {
	//			print("bucket:")
	//			println(bucketD)
	//		}
	//		print("key:")
	//		println(debugKey)
	//		flag = true
	//	}
	//	bucketD++
	//	d = (*dmap)(unsafe.Pointer(uintptr(h.buckets) + bucketD*uintptr(t.bucketsize)))
	//}
	//println()
	//取出第三位是否是1，如果是1则表示正有另外一个协程在往map里面写数据
	if h.flags&hashWriting != 0 {
		throw("concurrent map writes")
	}
	key := stringStructOf(&s)
	//获取key的hash值
	hash := t.hasher(noescape(unsafe.Pointer(&s)), uintptr(h.hash0))

	// Set hashWriting after calling t.hasher for consistency with mapassign.
	// 将标志位设置为正在写
	h.flags ^= hashWriting

	if h.buckets == nil {
		h.buckets = newobject(t.bucket) // newarray(t.bucket, 1)
	}

again:
	// 获取该key落到第几个bucket,每个bucket指的是类似链并的bmap结构
	mask := bucketMask(h.B)
	bucket := hash & mask
	// 如果存在扩容情况
	if h.growing() {
		// 从oldbuckets里面复制到新申请的buckets里面
		growWork_faststr(t, h, bucket)
	}
	// 寻址到第几个bmap
	b := (*bmap)(unsafe.Pointer(uintptr(h.buckets) + bucket*uintptr(t.bucketsize)))
	// 得到bmap的tophash值
	top := tophash(hash)

	var insertb *bmap          // 插入到哪个bmap里面
	var inserti uintptr        // 插入到bmap哪个位置
	var insertk unsafe.Pointer // 插入key到bmap哪个位置

	//找到一个空的地方插入该key
bucketloop:
	for {
		for i := uintptr(0); i < bucketCnt; i++ {
			if b.tophash[i] != top {
				if isEmpty(b.tophash[i]) && insertb == nil {
					//println("我确实进来了啊")
					insertb = b
					inserti = i
				}
				// 一开始都是0，也就是emptyRest
				// 当是emptyRest状态的时候就表示该桶后面没有数据了，或者说是后面的溢出桶也没有数据了，所以就不用接着找了，不用担心后面会找到一个相同key的数据
				if b.tophash[i] == emptyRest {
					break bucketloop
				}
				continue
			}
			// 到这里已经找到tophash了，2个不同的key也有可能相等，继续判断是否key相等
			// 在bucket中的key位置
			k := (*stringStruct)(add(unsafe.Pointer(b), dataOffset+i*2*sys.PtrSize))
			// 字符串key的长度都不等的话肯定不是一个key
			if k.len != key.len {
				continue
			}
			// 要么2个字符串直接相等，要么直接内存地址相等
			if k.str != key.str && !memequal(k.str, key.str, uintptr(key.len)) {
				continue
			}
			// already have a mapping for key. Update it.
			// 找到了相同的key，则要去更新value
			inserti = i
			insertb = b
			goto done
		}
		// 插入第9个的时候会走向这里，但是溢出的hmap是没有的
		ovf := b.overflow(t)
		if ovf == nil {
			break
		}
		b = ovf
	}

	// Did not find mapping for key. Allocate new cell & add entry.

	// If we hit the max load factor or we have too many overflow buckets,
	// and we're not already in the middle of growing, start growing.
	// 如果次数个数超出了增长因子，或者没有超出增长因子，但是有太多的逸出桶了，这个和java的hashmap一样，当太多红黑树了，还是会影响查找效率，因为理想情况下，map的
	// 查找效率应该是o(1)
	if !h.growing() && (overLoadFactor(h.count+1, h.B) || tooManyOverflowBuckets(h.noverflow, h.B)) {
		//d := (*dmap)(unsafe.Pointer(uintptr(h.buckets)))
		//bucketD := uintptr(0)
		//for bucketD < bucketShift(h.B)+3 {
		//	flag := false
		//	for i, debugKey := range d.debugKeys {
		//		if debugKey == "" {
		//			continue
		//		}
		//		println(d.tophash[i])
		//		if flag == false {
		//			print("bucket:")
		//			println(bucketD)
		//		}
		//		print("key:")
		//		println(debugKey)
		//		flag = true
		//	}
		//	bucketD++
		//	d = (*dmap)(unsafe.Pointer(uintptr(h.buckets) + bucketD*uintptr(t.bucketsize)))
		//}
		//println()
		hashGrow(t, h)
		goto again // Growing the table invalidates everything, so try again
	}

	if insertb == nil {
		// all current buckets are full, allocate a new one.
		//println(bucket)
		insertb = h.newoverflow(t, b)
		inserti = 0 // not necessary, but avoids needlessly spilling inserti
	}
	// 把tophash值放到topsh槽里面去
	insertb.tophash[inserti&(bucketCnt-1)] = top // mask inserti to avoid bounds checks

	// 把key放到bmap里面
	// dataOffset是为了得到内存对齐后的key的位置
	// 为什么插入的是2*sys.PtrSize呢，因为string其实占了16字节
	insertk = add(unsafe.Pointer(insertb), dataOffset+inserti*2*sys.PtrSize)
	// store new key at insert position
	// 这块内存就放key的值
	*((*stringStruct)(insertk)) = *key
	// key个数加1
	h.count++

done:
	// done不关心是否是更新还是新增，拿到相应的位置即可
	// 找到value存的内存位置
	elem := add(unsafe.Pointer(insertb), dataOffset+bucketCnt*2*sys.PtrSize+inserti*uintptr(t.elemsize))
	if h.flags&hashWriting == 0 {
		throw("concurrent map writes")
	}
	// 将标志位恢复
	h.flags &^= hashWriting
	return elem
}

func mapdelete_faststr(t *maptype, h *hmap, ky string) {
	if raceenabled && h != nil {
		callerpc := getcallerpc()
		racewritepc(unsafe.Pointer(h), callerpc, funcPC(mapdelete_faststr))
	}
	if h == nil || h.count == 0 {
		return
	}
	if h.flags&hashWriting != 0 {
		throw("concurrent map writes")
	}

	key := stringStructOf(&ky)
	hash := t.hasher(noescape(unsafe.Pointer(&ky)), uintptr(h.hash0))

	// Set hashWriting after calling t.hasher for consistency with mapdelete
	h.flags ^= hashWriting

	bucket := hash & bucketMask(h.B)
	// 顺便迁移
	if h.growing() {
		growWork_faststr(t, h, bucket)
	}
	// 找到key所在的桶
	b := (*bmap)(add(h.buckets, bucket*uintptr(t.bucketsize)))
	// 记录下刚开始key的所在的桶
	bOrig := b
	top := tophash(hash)
search:
	// 外层循环该key所在的桶以及桶后面的逸出桶
	for ; b != nil; b = b.overflow(t) {
		// 遍历桶的数据
		for i, kptr := uintptr(0), b.keys(); i < bucketCnt; i, kptr = i+1, add(kptr, 2*sys.PtrSize) {
			k := (*stringStruct)(kptr)
			if k.len != key.len || b.tophash[i] != top {
				continue
			}
			if k.str != key.str && !memequal(k.str, key.str, uintptr(key.len)) {
				continue
			}
			// Clear key's pointer.
			// 清除key的字符，把长度留下
			k.str = nil
			e := add(unsafe.Pointer(b), dataOffset+bucketCnt*2*sys.PtrSize+i*uintptr(t.elemsize))
			// 清除value的内存
			if t.elem.ptrdata != 0 {
				memclrHasPointers(e, t.elem.size)
			} else {
				memclrNoHeapPointers(e, t.elem.size)
			}
			// 标记为值已经被清除
			b.tophash[i] = emptyOne
			// If the bucket now ends in a bunch of emptyOne states,
			// change those to emptyRest states.
			// 判断该所在桶的key后面的key是否都已经被清空
			// 或者如果该key已经是桶内的第8个key，那么就判断该桶的所有逸出桶是否已经被清空
			if i == bucketCnt-1 {
				if b.overflow(t) != nil && b.overflow(t).tophash[0] != emptyRest {
					goto notLast
				}
			} else {
				if b.tophash[i+1] != emptyRest {
					goto notLast
				}
			}
			// 往前面一直去试图标记桶的hash槽为emptyRest状态
			// 当该槽的状态为emptyRest之后，那么就是该key之后所有的槽，以及该桶后面的逸出桶都是emptyRest
			for {
				b.tophash[i] = emptyRest
				if i == 0 {
					if b == bOrig {
						break // beginning of initial bucket, we're done.
					}
					// Find previous bucket, continue at its last entry.
					c := b
					// 去找该桶的前一个桶
					for b = bOrig; b.overflow(t) != c; b = b.overflow(t) {
					}
					i = bucketCnt - 1
				} else {
					i--
				}
				// 如果是emptyOne，也就是被删除了，那么就标记为emptyRest
				if b.tophash[i] != emptyOne {
					break
				}
			}
		notLast:
			h.count--
			break search
		}
	}

	if h.flags&hashWriting == 0 {
		throw("concurrent map writes")
	}
	h.flags &^= hashWriting
}

func growWork_faststr(t *maptype, h *hmap, bucket uintptr) {
	// make sure we evacuate the oldbucket corresponding
	// to the bucket we're about to use
	// bucket&h.oldbucketmask() 会得到一个oldbuckets中的一个bucket，然后把该bucket里面的数据移到新的bucket里面去，即新申请的。
	mask := h.oldbucketmask()
	oldbucket := bucket & mask
	// 迁移逻辑
	evacuate_faststr(t, h, oldbucket)

	// evacuate one more oldbucket to make progress on growing
	// 如果还有oldbuckets需要迁移
	if h.growing() {
		// 继续移除oldbuckets第nevacuate+1个，也是继续迁移上次迁移后面的那一个桶
		evacuate_faststr(t, h, h.nevacuate)
	}
}

func evacuate_faststr(t *maptype, h *hmap, oldbucket uintptr) {
	// 得到即将移动到新的bucket的老bucket的地址
	b := (*bmap)(add(h.oldbuckets, oldbucket*uintptr(t.bucketsize)))
	// 多少个老的bucket
	newbit := h.noldbuckets()
	// 判断是否已经被移动到新的bucket
	if !evacuated(b) {
		// TODO: reuse overflow buckets instead of using new ones, if there
		// is no iterator using the old buckets.  (If !oldIterator.)

		// xy contains the x and y (low and high) evacuation destinations.
		// xy是把老的bucket的数据分配到2个新的bucket里面去,hash&newbit != 0 决定放到x还是y的bucket里面去
		var xy [2]evacDst
		// x和老的bucket同一个位置
		x := &xy[0]
		// bucket的起始地址
		x.b = (*bmap)(add(h.buckets, oldbucket*uintptr(t.bucketsize)))
		// bucket内的key的起始地址
		x.k = add(unsafe.Pointer(x.b), dataOffset)
		// bucket内的value的起始地址
		x.e = add(x.k, bucketCnt*2*sys.PtrSize)

		if !h.sameSizeGrow() {
			// Only calculate y pointers if we're growing bigger.
			// Otherwise GC can see bad pointers.
			// 偏移newbit位，得到y的bucket位置，也就是说x和y隔了老buckets个数的地址
			y := &xy[1]
			y.b = (*bmap)(add(h.buckets, (oldbucket+newbit)*uintptr(t.bucketsize)))
			y.k = add(unsafe.Pointer(y.b), dataOffset)
			y.e = add(y.k, bucketCnt*2*sys.PtrSize)
		}

		// 开始复制到新的bucket，注意也会遍历溢出桶
		for ; b != nil; b = b.overflow(t) {
			// bucket的key的起始地址
			k := add(unsafe.Pointer(b), dataOffset)
			// bucket的value的起始地址
			e := add(k, bucketCnt*2*sys.PtrSize)
			for i := 0; i < bucketCnt; i, k, e = i+1, add(k, 2*sys.PtrSize), add(e, uintptr(t.elemsize)) {
				top := b.tophash[i]
				if isEmpty(top) {
					b.tophash[i] = evacuatedEmpty
					continue
				}
				if top < minTopHash {
					throw("bad map state")
				}
				var useY uint8
				if !h.sameSizeGrow() {
					// Compute hash to make our evacuation decision (whether we need
					// to send this key/elem to bucket x or bucket y).
					// 求得key的hash值决定放到x还是y
					hash := t.hasher(k, uintptr(h.hash0))
					if hash&newbit != 0 {
						useY = 1
					}
				}

				// 挡在移动key的过程当中，会给tophsh设置一些标志位，evacuatedX表示会移动到x，evacuatedY表示会移动y，
				// 当useY是1的时候就移动到y
				b.tophash[i] = evacuatedX + useY // evacuatedX + 1 == evacuatedY, enforced in makemap
				// 得到移动的目标
				dst := &xy[useY] // evacuation destination

				// 一个bmap只存8个key/value，超过了就要用dest的溢出桶
				if dst.i == bucketCnt {
					// 申请一个溢出桶
					dst.b = h.newoverflow(t, dst.b)
					// 新桶要重置位0
					dst.i = 0
					// 得到新桶的key的位置
					dst.k = add(unsafe.Pointer(dst.b), dataOffset)
					// 得到新桶的value的位置
					dst.e = add(dst.k, bucketCnt*2*sys.PtrSize)
				}
				// 复制tophash
				dst.b.tophash[dst.i&(bucketCnt-1)] = top // mask dst.i as an optimization, to avoid a bounds check

				// Copy key.
				// 把老桶的key复制到dst
				*(*string)(dst.k) = *(*string)(k)
				// 把老桶的value复制到dst
				typedmemmove(t.elem, dst.e, e)
				// 桶个数+1
				dst.i++
				// These updates might push these pointers past the end of the
				// key or elem arrays.  That's ok, as we have the overflow pointer
				// at the end of the bucket to protect against pointing past the
				// end of the bucket.
				// dst的key向前偏移一个key的大小
				dst.k = add(dst.k, 2*sys.PtrSize)
				// dst的value向前偏移一个key的大小
				dst.e = add(dst.e, uintptr(t.elemsize))
			}
		}
		// Unlink the overflow buckets & clear key/elem to help GC.
		// 把老桶释放掉
		if h.flags&oldIterator == 0 && t.bucket.ptrdata != 0 {
			b := add(h.oldbuckets, oldbucket*uintptr(t.bucketsize))
			// Preserve b.tophash because the evacuation
			// state is maintained there.
			// 保留了8个tophash槽不释放，只释放key和value的内存
			// 得到key的内存地址
			ptr := add(b, dataOffset)
			// 得到一个bucket除了8个tophash槽之后的偏移
			n := uintptr(t.bucketsize) - dataOffset
			// 释放地址
			memclrHasPointers(ptr, n)
		}
	}

	// 这段代码比较魔幻
	// 据我分析，第一次进入应该当时oldbucket为0的时候。后面也有可能进入，取决于nevacuate的值
	// 【0，nevacuate+1】之前的oldbuckets已经被迁移了.
	// 上面只是把bucket里面的数据清除掉了，但是tophash值还在。
	if oldbucket == h.nevacuate {
		advanceEvacuationMark(h, t, newbit)
	}
}
