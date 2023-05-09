/*
   Copyright 2021 Erigon contributors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package state

import (
	"bytes"
	"container/heap"
	"context"
	"encoding/binary"
	"fmt"
	"hash"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/btree"
	"github.com/ledgerwatch/log/v3"
	"golang.org/x/crypto/sha3"

	"github.com/ledgerwatch/erigon-lib/common/background"

	"github.com/ledgerwatch/erigon-lib/commitment"
	"github.com/ledgerwatch/erigon-lib/common"
	"github.com/ledgerwatch/erigon-lib/common/length"
	"github.com/ledgerwatch/erigon-lib/compress"
)

// Defines how to evaluate commitments
type CommitmentMode uint

const (
	CommitmentModeDisabled CommitmentMode = 0
	CommitmentModeDirect   CommitmentMode = 1
	CommitmentModeUpdate   CommitmentMode = 2
)

func (m CommitmentMode) String() string {
	switch m {
	case CommitmentModeDisabled:
		return "disabled"
	case CommitmentModeDirect:
		return "direct"
	case CommitmentModeUpdate:
		return "update"
	default:
		return "unknown"
	}
}

func ParseCommitmentMode(s string) CommitmentMode {
	var mode CommitmentMode
	switch s {
	case "off":
		mode = CommitmentModeDisabled
	case "update":
		mode = CommitmentModeUpdate
	default:
		mode = CommitmentModeDirect
	}
	return mode
}

type ValueMerger func(prev, current []byte) (merged []byte, err error)

type UpdateTree struct {
	tree   *btree.BTreeG[*CommitmentItem]
	mode   CommitmentMode
	keccak hash.Hash
}

func NewUpdateTree(mode CommitmentMode) *UpdateTree {
	return &UpdateTree{
		tree:   btree.NewG[*CommitmentItem](32, commitmentItemLess),
		mode:   mode,
		keccak: sha3.NewLegacyKeccak256(),
	}
}

// TouchPlainKey marks plainKey as updated and applies different fn for different key types
// (different behaviour for Code, Account and Storage key modifications).
func (t *UpdateTree) TouchPlainKey(key, val []byte, fn func(c *CommitmentItem, val []byte)) {
	if t.mode == CommitmentModeDisabled {
		return
	}
	c := &CommitmentItem{plainKey: common.Copy(key), hashedKey: t.hashAndNibblizeKey(key)}
	if t.mode > CommitmentModeDirect {
		fn(c, val)
	}
	t.tree.ReplaceOrInsert(c)
}

func (t *UpdateTree) TouchAccountKey(c *CommitmentItem, val []byte) {
	if len(val) == 0 {
		c.update.Flags = commitment.DeleteUpdate
		return
	}
	c.update.DecodeForStorage(val)
	item, found := t.tree.Get(&CommitmentItem{hashedKey: c.hashedKey})
	//fmt.Printf("TouchAccountKey: found=%t, %x\n", found, c.hashedKey)
	//if item != nil {
	//	fmt.Printf("TouchAccountKey2: %t, %x, %x\n", item.update.Flags&commitment.CodeUpdate != 0, item.update.CodeHashOrStorage[:], c.update.CodeHashOrStorage[:])
	//}
	if found && item.update.Flags&commitment.CodeUpdate != 0 {
		c.update.Flags |= commitment.CodeUpdate
		copy(c.update.CodeHashOrStorage[:], item.update.CodeHashOrStorage[:])
	}
}

func (t *UpdateTree) UpdatePrefix(prefix, val []byte, fn func(c *CommitmentItem, val []byte)) {
	t.tree.AscendGreaterOrEqual(&CommitmentItem{}, func(item *CommitmentItem) bool {
		if !bytes.HasPrefix(item.plainKey, prefix) {
			return false
		}
		fn(item, val)
		return true
	})
}

func (t *UpdateTree) TouchStorageKey(c *CommitmentItem, val []byte) {
	c.update.ValLength = len(val)
	if len(val) == 0 {
		c.update.Flags = commitment.DeleteUpdate
	} else {
		c.update.Flags = commitment.StorageUpdate
		copy(c.update.CodeHashOrStorage[:], val)
	}
}

func (t *UpdateTree) TouchCodeKey(c *CommitmentItem, val []byte) {
	c.update.Flags = commitment.CodeUpdate
	item, found := t.tree.Get(c)
	if !found {
		t.keccak.Reset()
		t.keccak.Write(val)
		copy(c.update.CodeHashOrStorage[:], t.keccak.Sum(nil))
		return
	}
	if item.update.Flags&commitment.BalanceUpdate != 0 {
		c.update.Flags |= commitment.BalanceUpdate
		c.update.Balance.Set(&item.update.Balance)
	}
	if item.update.Flags&commitment.NonceUpdate != 0 {
		c.update.Flags |= commitment.NonceUpdate
		c.update.Nonce = item.update.Nonce
	}
	if item.update.Flags == commitment.DeleteUpdate && len(val) == 0 {
		c.update.Flags = commitment.DeleteUpdate
	} else {
		t.keccak.Reset()
		t.keccak.Write(val)
		copy(c.update.CodeHashOrStorage[:], t.keccak.Sum(nil))
	}
}

// Returns list of both plain and hashed keys. If .mode is CommitmentModeUpdate, updates also returned.
func (t *UpdateTree) List() ([][]byte, [][]byte, []commitment.Update) {
	plainKeys := make([][]byte, t.tree.Len())
	hashedKeys := make([][]byte, t.tree.Len())
	updates := make([]commitment.Update, t.tree.Len())

	j := 0
	t.tree.Ascend(func(item *CommitmentItem) bool {
		fmt.Printf("List(): hk=%x,  %x, balance=%s, nonce=%d, %x\n", item.hashedKey, item.plainKey, item.update.Balance.String(), item.update.Nonce, item.update.CodeHashOrStorage)
		plainKeys[j] = item.plainKey
		hashedKeys[j] = item.hashedKey
		updates[j] = item.update
		j++
		return true
	})

	t.tree.Clear(true)
	return plainKeys, hashedKeys, updates
}

// TODO(awskii): let trie define hashing function
func (t *UpdateTree) hashAndNibblizeKey(key []byte) []byte {
	hashedKey := make([]byte, length.Hash)

	t.keccak.Reset()
	t.keccak.Write(key[:length.Addr])
	copy(hashedKey[:length.Hash], t.keccak.Sum(nil))

	if len(key[length.Addr:]) > 0 {
		hashedKey = append(hashedKey, make([]byte, length.Hash)...)
		t.keccak.Reset()
		t.keccak.Write(key[length.Addr:])
		copy(hashedKey[length.Hash:], t.keccak.Sum(nil))
	}

	nibblized := make([]byte, len(hashedKey)*2)
	for i, b := range hashedKey {
		nibblized[i*2] = (b >> 4) & 0xf
		nibblized[i*2+1] = b & 0xf
	}
	return nibblized
}

type DomainCommitted struct {
	*Domain
	trace        bool
	updates      *UpdateTree
	patriciaTrie commitment.Trie
	branchMerger *commitment.BranchMerger
	prevState    []byte

	comKeys uint64
	comTook time.Duration
}

func (d *DomainCommitted) ResetFns(
	branchFn func(prefix []byte) ([]byte, error),
	accountFn func(plainKey []byte, cell *commitment.Cell) error,
	storageFn func(plainKey []byte, cell *commitment.Cell) error,
) {
	d.patriciaTrie.ResetFns(branchFn, accountFn, storageFn)
}

func (d *DomainCommitted) Hasher() hash.Hash {
	return d.updates.keccak
}

func NewCommittedDomain(d *Domain, mode CommitmentMode, trieVariant commitment.TrieVariant) *DomainCommitted {
	return &DomainCommitted{
		Domain:       d,
		updates:      NewUpdateTree(mode),
		patriciaTrie: commitment.InitializeTrie(trieVariant),
		branchMerger: commitment.NewHexBranchMerger(8192),
	}
}

func (d *DomainCommitted) SetCommitmentMode(m CommitmentMode) { d.updates.mode = m }

// TouchPlainKey marks plainKey as updated and applies different fn for different key types
// (different behaviour for Code, Account and Storage key modifications).
func (d *DomainCommitted) TouchPlainKey(key, val []byte, fn func(c *CommitmentItem, val []byte)) {
	d.updates.TouchPlainKey(key, val, fn)
}

func (d *DomainCommitted) TouchAccount(c *CommitmentItem, val []byte) {
	d.updates.TouchAccountKey(c, val)
}

func (d *DomainCommitted) TouchStorage(c *CommitmentItem, val []byte) {
	d.updates.TouchStorageKey(c, val)
}

func (d *DomainCommitted) TouchCode(c *CommitmentItem, val []byte) {
	d.updates.TouchCodeKey(c, val)
}

type CommitmentItem struct {
	plainKey  []byte
	hashedKey []byte
	update    commitment.Update
}

func commitmentItemLess(i, j *CommitmentItem) bool {
	return bytes.Compare(i.hashedKey, j.hashedKey) < 0
}

func (d *DomainCommitted) storeCommitmentState(blockNum uint64) error {
	var state []byte
	var err error

	switch trie := (d.patriciaTrie).(type) {
	case *commitment.HexPatriciaHashed:
		state, err = trie.EncodeCurrentState(nil)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported state storing for patricia trie type: %T", d.patriciaTrie)
	}
	cs := &commitmentState{txNum: d.txNum, trieState: state, blockNum: blockNum}
	encoded, err := cs.Encode()
	if err != nil {
		return err
	}

	var stepbuf [2]byte
	step := uint16(d.txNum / d.aggregationStep)
	binary.BigEndian.PutUint16(stepbuf[:], step)
	switch d.Domain.wal {
	case nil:
		if err = d.Domain.Put(keyCommitmentState, stepbuf[:], encoded); err != nil {
			return err
		}
	default:
		if err := d.Domain.PutWithPrev(keyCommitmentState, stepbuf[:], encoded, d.prevState); err != nil {
			return err
		}
		d.prevState = encoded
	}
	return nil
}

// nolint
func (d *DomainCommitted) replaceKeyWithReference(fullKey, shortKey []byte, typeAS string, list ...*filesItem) bool {
	numBuf := [2]byte{}
	var found bool
	for _, item := range list {
		//g := item.decompressor.MakeGetter()
		//index := recsplit.NewIndexReader(item.index)

		cur, err := item.bindex.Seek(fullKey)
		if err != nil {
			continue
		}
		step := uint16(item.endTxNum / d.aggregationStep)
		binary.BigEndian.PutUint16(numBuf[:], step)

		shortKey = encodeU64(cur.Ordinal(), numBuf[:])

		if d.trace {
			fmt.Printf("replacing %s [%x] => {%x} [step=%d, offset=%d, file=%s.%d-%d]\n", typeAS, fullKey, shortKey, step, cur.Ordinal(), typeAS, item.startTxNum, item.endTxNum)
		}
		found = true
		break
	}
	//if !found {
	//	log.Warn("bt index key replacement seek failed", "key", fmt.Sprintf("%x", fullKey))
	//}
	return found
}

// nolint
func (d *DomainCommitted) lookupShortenedKey(shortKey, fullKey []byte, typAS string, list []*filesItem) bool {
	fileStep, offset := shortenedKey(shortKey)
	expected := uint64(fileStep) * d.aggregationStep

	var found bool
	for _, item := range list {
		if item.startTxNum > expected || item.endTxNum < expected {
			continue
		}

		cur := item.bindex.OrdinalLookup(offset)
		//nolint
		fullKey = cur.Key()
		if d.trace {
			fmt.Printf("offsetToKey %s [%x]=>{%x} step=%d offset=%d, file=%s.%d-%d.kv\n", typAS, fullKey, shortKey, fileStep, offset, typAS, item.startTxNum, item.endTxNum)
		}
		found = true
		break
	}
	return found
}

// commitmentValTransform parses the value of the commitment record to extract references
// to accounts and storage items, then looks them up in the new, merged files, and replaces them with
// the updated references
func (d *DomainCommitted) commitmentValTransform(files *SelectedStaticFiles, merged *MergedFiles, val commitment.BranchData) ([]byte, error) {
	if len(val) == 0 {
		return nil, nil
	}
	accountPlainKeys, storagePlainKeys, err := val.ExtractPlainKeys()
	if err != nil {
		return nil, err
	}

	transAccountPks := make([][]byte, 0, len(accountPlainKeys))
	var apkBuf, spkBuf []byte
	for _, accountPlainKey := range accountPlainKeys {
		if len(accountPlainKey) == length.Addr {
			// Non-optimised key originating from a database record
			apkBuf = append(apkBuf[:0], accountPlainKey...)
		} else {
			f := d.lookupShortenedKey(accountPlainKey, apkBuf, "account", files.accounts)
			if !f {
				fmt.Printf("lost key %x\n", accountPlainKeys)
			}
		}
		d.replaceKeyWithReference(apkBuf, accountPlainKey, "account", merged.accounts)
		transAccountPks = append(transAccountPks, accountPlainKey)
	}

	transStoragePks := make([][]byte, 0, len(storagePlainKeys))
	for _, storagePlainKey := range storagePlainKeys {
		if len(storagePlainKey) == length.Addr+length.Hash {
			// Non-optimised key originating from a database record
			spkBuf = append(spkBuf[:0], storagePlainKey...)
		} else {
			// Optimised key referencing a state file record (file number and offset within the file)
			f := d.lookupShortenedKey(storagePlainKey, spkBuf, "storage", files.storage)
			if !f {
				fmt.Printf("lost skey %x\n", storagePlainKey)
			}
		}

		d.replaceKeyWithReference(spkBuf, storagePlainKey, "storage", merged.storage)
		transStoragePks = append(transStoragePks, storagePlainKey)
	}

	transValBuf, err := val.ReplacePlainKeys(transAccountPks, transStoragePks, nil)
	if err != nil {
		return nil, err
	}
	return transValBuf, nil
}

func (d *DomainCommitted) mergeFiles(ctx context.Context, oldFiles SelectedStaticFiles, mergedFiles MergedFiles, r DomainRanges, workers int, ps *background.ProgressSet) (valuesIn, indexIn, historyIn *filesItem, err error) {
	if !r.any() {
		return
	}

	domainFiles := oldFiles.commitment
	indexFiles := oldFiles.commitmentIdx
	historyFiles := oldFiles.commitmentHist

	var comp *compress.Compressor
	var closeItem bool = true
	defer func() {
		if closeItem {
			if comp != nil {
				comp.Close()
			}
			if indexIn != nil {
				if indexIn.decompressor != nil {
					indexIn.decompressor.Close()
				}
				if indexIn.index != nil {
					indexIn.index.Close()
				}
				if indexIn.bindex != nil {
					indexIn.bindex.Close()
				}
			}
			if historyIn != nil {
				if historyIn.decompressor != nil {
					historyIn.decompressor.Close()
				}
				if historyIn.index != nil {
					historyIn.index.Close()
				}
				if historyIn.bindex != nil {
					historyIn.bindex.Close()
				}
			}
			if valuesIn != nil {
				if valuesIn.decompressor != nil {
					valuesIn.decompressor.Close()
				}
				if valuesIn.index != nil {
					valuesIn.index.Close()
				}
				if valuesIn.bindex != nil {
					valuesIn.bindex.Close()
				}
			}
		}
	}()
	if indexIn, historyIn, err = d.History.mergeFiles(ctx, indexFiles, historyFiles,
		HistoryRanges{
			historyStartTxNum: r.historyStartTxNum,
			historyEndTxNum:   r.historyEndTxNum,
			history:           r.history,
			indexStartTxNum:   r.indexStartTxNum,
			indexEndTxNum:     r.indexEndTxNum,
			index:             r.index}, workers, ps); err != nil {
		return nil, nil, nil, err
	}

	if r.values {
		datFileName := fmt.Sprintf("%s.%d-%d.kv", d.filenameBase, r.valuesStartTxNum/d.aggregationStep, r.valuesEndTxNum/d.aggregationStep)
		datPath := filepath.Join(d.dir, datFileName)
		p := ps.AddNew(datFileName, 1)
		defer ps.Delete(p)

		if comp, err = compress.NewCompressor(ctx, "merge", datPath, d.dir, compress.MinPatternScore, workers, log.LvlTrace); err != nil {
			return nil, nil, nil, fmt.Errorf("merge %s compressor: %w", d.filenameBase, err)
		}
		var cp CursorHeap
		heap.Init(&cp)
		for _, item := range domainFiles {
			g := item.decompressor.MakeGetter()
			g.Reset(0)
			if g.HasNext() {
				key, _ := g.NextUncompressed()
				var val []byte
				if d.compressVals {
					val, _ = g.Next(nil)
				} else {
					val, _ = g.NextUncompressed()
				}
				if d.trace {
					fmt.Printf("merge: read value '%x'\n", key)
				}
				heap.Push(&cp, &CursorItem{
					t:        FILE_CURSOR,
					dg:       g,
					key:      key,
					val:      val,
					endTxNum: item.endTxNum,
					reverse:  true,
				})
			}
		}
		keyCount := 0
		// In the loop below, the pair `keyBuf=>valBuf` is always 1 item behind `lastKey=>lastVal`.
		// `lastKey` and `lastVal` are taken from the top of the multi-way merge (assisted by the CursorHeap cp), but not processed right away
		// instead, the pair from the previous iteration is processed first - `keyBuf=>valBuf`. After that, `keyBuf` and `valBuf` are assigned
		// to `lastKey` and `lastVal` correspondingly, and the next step of multi-way merge happens. Therefore, after the multi-way merge loop
		// (when CursorHeap cp is empty), there is a need to process the last pair `keyBuf=>valBuf`, because it was one step behind
		var keyBuf, valBuf []byte
		for cp.Len() > 0 {
			lastKey := common.Copy(cp[0].key)
			lastVal := common.Copy(cp[0].val)
			// Advance all the items that have this key (including the top)
			for cp.Len() > 0 && bytes.Equal(cp[0].key, lastKey) {
				ci1 := cp[0]
				if ci1.dg.HasNext() {
					ci1.key, _ = ci1.dg.NextUncompressed()
					if d.compressVals {
						ci1.val, _ = ci1.dg.Next(ci1.val[:0])
					} else {
						ci1.val, _ = ci1.dg.NextUncompressed()
					}
					heap.Fix(&cp, 0)
				} else {
					heap.Pop(&cp)
				}
			}
			// For the rest of types, empty value means deletion
			skip := r.valuesStartTxNum == 0 && len(lastVal) == 0
			if !skip {
				if keyBuf != nil {
					if err = comp.AddUncompressedWord(keyBuf); err != nil {
						return nil, nil, nil, err
					}
					keyCount++ // Only counting keys, not values
					switch d.compressVals {
					case true:
						if err = comp.AddWord(valBuf); err != nil {
							return nil, nil, nil, err
						}
					default:
						if err = comp.AddUncompressedWord(valBuf); err != nil {
							return nil, nil, nil, err
						}
					}
				}
				keyBuf = append(keyBuf[:0], lastKey...)
				valBuf = append(valBuf[:0], lastVal...)
			}
		}
		if keyBuf != nil {
			if err = comp.AddUncompressedWord(keyBuf); err != nil {
				return nil, nil, nil, err
			}
			keyCount++ // Only counting keys, not values
			//fmt.Printf("last heap key %x\n", keyBuf)
			valBuf, err = d.commitmentValTransform(&oldFiles, &mergedFiles, valBuf)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("merge: 2valTransform [%x] %w", valBuf, err)
			}
			if d.compressVals {
				if err = comp.AddWord(valBuf); err != nil {
					return nil, nil, nil, err
				}
			} else {
				if err = comp.AddUncompressedWord(valBuf); err != nil {
					return nil, nil, nil, err
				}
			}
		}
		if err = comp.Compress(); err != nil {
			return nil, nil, nil, err
		}
		comp.Close()
		comp = nil
		valuesIn = newFilesItem(r.valuesStartTxNum, r.valuesEndTxNum, d.aggregationStep)
		if valuesIn.decompressor, err = compress.NewDecompressor(datPath); err != nil {
			return nil, nil, nil, fmt.Errorf("merge %s decompressor [%d-%d]: %w", d.filenameBase, r.valuesStartTxNum, r.valuesEndTxNum, err)
		}
		ps.Delete(p)

		idxFileName := fmt.Sprintf("%s.%d-%d.kvi", d.filenameBase, r.valuesStartTxNum/d.aggregationStep, r.valuesEndTxNum/d.aggregationStep)
		idxPath := filepath.Join(d.dir, idxFileName)

		p = ps.AddNew(datFileName, uint64(keyCount))
		defer ps.Delete(p)
		if valuesIn.index, err = buildIndexThenOpen(ctx, valuesIn.decompressor, idxPath, d.dir, keyCount, false /* values */, p); err != nil {
			return nil, nil, nil, fmt.Errorf("merge %s buildIndex [%d-%d]: %w", d.filenameBase, r.valuesStartTxNum, r.valuesEndTxNum, err)
		}

		btPath := strings.TrimSuffix(idxPath, "kvi") + "bt"
		valuesIn.bindex, err = CreateBtreeIndexWithDecompressor(btPath, 2048, valuesIn.decompressor, p)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("create btindex %s [%d-%d]: %w", d.filenameBase, r.valuesStartTxNum, r.valuesEndTxNum, err)
		}
	}
	closeItem = false
	d.stats.MergesCount++
	return
}

// Deprecated?
func (d *DomainCommitted) CommitmentOver(touchedKeys, hashedKeys [][]byte, updates []commitment.Update, trace bool) (rootHash []byte, branchNodeUpdates map[string]commitment.BranchData, err error) {
	defer func(s time.Time) { d.comTook = time.Since(s) }(time.Now())

	d.comKeys = uint64(len(touchedKeys))
	if len(touchedKeys) == 0 {
		rootHash, err = d.patriciaTrie.RootHash()
		return rootHash, nil, err
	}

	// data accessing functions should be set once before
	d.patriciaTrie.Reset()
	d.patriciaTrie.SetTrace(trace)

	switch d.updates.mode {
	case CommitmentModeDirect:
		rootHash, branchNodeUpdates, err = d.patriciaTrie.ReviewKeys(touchedKeys, hashedKeys)
		if err != nil {
			return nil, nil, err
		}
	case CommitmentModeUpdate:
		rootHash, branchNodeUpdates, err = d.patriciaTrie.ProcessUpdates(touchedKeys, hashedKeys, updates)
		if err != nil {
			return nil, nil, err
		}
	case CommitmentModeDisabled:
		return nil, nil, nil
	default:
		return nil, nil, fmt.Errorf("invalid commitment mode: %d", d.updates.mode)
	}
	return rootHash, branchNodeUpdates, err
}

// Evaluates commitment for processed state.
func (d *DomainCommitted) ComputeCommitment(trace bool) (rootHash []byte, branchNodeUpdates map[string]commitment.BranchData, err error) {
	defer func(s time.Time) { d.comTook = time.Since(s) }(time.Now())

	touchedKeys, hashedKeys, updates := d.updates.List()
	d.comKeys = uint64(len(touchedKeys))

	if len(touchedKeys) == 0 {
		rootHash, err = d.patriciaTrie.RootHash()
		return rootHash, nil, err
	}

	// data accessing functions should be set once before
	d.patriciaTrie.Reset()
	d.patriciaTrie.SetTrace(trace)
	switch d.updates.mode {
	case CommitmentModeDirect:
		rootHash, branchNodeUpdates, err = d.patriciaTrie.ReviewKeys(touchedKeys, hashedKeys)
		if err != nil {
			return nil, nil, err
		}
	case CommitmentModeUpdate:
		rootHash, branchNodeUpdates, err = d.patriciaTrie.ProcessUpdates(touchedKeys, hashedKeys, updates)
		if err != nil {
			return nil, nil, err
		}
	case CommitmentModeDisabled:
		return nil, nil, nil
	default:
		return nil, nil, fmt.Errorf("invalid commitment mode: %d", d.updates.mode)
	}
	return rootHash, branchNodeUpdates, err
}

func (d *DomainCommitted) Close() {
	d.Domain.Close()
	d.updates.tree.Clear(true)
}

var keyCommitmentState = []byte("state")

// SeekCommitment searches for last encoded state from DomainCommitted
// and if state found, sets it up to current domain
func (d *DomainCommitted) SeekCommitment(aggStep, sinceTx uint64) (blockNum, txNum uint64, err error) {
	if d.patriciaTrie.Variant() != commitment.VariantHexPatriciaTrie {
		return 0, 0, fmt.Errorf("state storing is only supported hex patricia trie")
	}
	// todo add support of bin state dumping

	var (
		latestState []byte
		stepbuf     [2]byte
		step               = uint16(sinceTx/aggStep) - 1
		latestTxNum uint64 = sinceTx - 1
	)
	if sinceTx == 0 {
		step = 0
		latestTxNum = 0
	}

	d.SetTxNum(latestTxNum)
	ctx := d.MakeContext()
	defer ctx.Close()

	for {
		binary.BigEndian.PutUint16(stepbuf[:], step)

		s, err := ctx.Get(keyCommitmentState, stepbuf[:], d.tx)
		if err != nil {
			return 0, 0, err
		}
		if len(s) < 8 {
			break
		}
		v := binary.BigEndian.Uint64(s)
		if v == latestTxNum && len(latestState) != 0 {
			break
		}
		latestTxNum, latestState = v, s
		lookupTxN := latestTxNum + aggStep
		step = uint16(latestTxNum/aggStep) + 1
		d.SetTxNum(lookupTxN)
	}

	var latest commitmentState
	if err := latest.Decode(latestState); err != nil {
		return 0, 0, nil
	}

	if hext, ok := d.patriciaTrie.(*commitment.HexPatriciaHashed); ok {
		if err := hext.SetState(latest.trieState); err != nil {
			return 0, 0, err
		}
	} else {
		return 0, 0, fmt.Errorf("state storing is only supported hex patricia trie")
	}

	return latest.blockNum, latest.txNum, nil
}

type commitmentState struct {
	txNum     uint64
	blockNum  uint64
	trieState []byte
}

func (cs *commitmentState) Decode(buf []byte) error {
	if len(buf) < 10 {
		return fmt.Errorf("ivalid commitment state buffer size")
	}
	pos := 0
	cs.txNum = binary.BigEndian.Uint64(buf[pos : pos+8])
	pos += 8
	cs.blockNum = binary.BigEndian.Uint64(buf[pos : pos+8])
	pos += 8
	cs.trieState = make([]byte, binary.BigEndian.Uint16(buf[pos:pos+2]))
	pos += 2
	if len(cs.trieState) == 0 && len(buf) == 10 {
		return nil
	}
	copy(cs.trieState, buf[pos:pos+len(cs.trieState)])
	return nil
}

func (cs *commitmentState) Encode() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	var v [18]byte
	binary.BigEndian.PutUint64(v[:], cs.txNum)
	binary.BigEndian.PutUint64(v[8:16], cs.blockNum)
	binary.BigEndian.PutUint16(v[16:18], uint16(len(cs.trieState)))
	if _, err := buf.Write(v[:]); err != nil {
		return nil, err
	}
	if _, err := buf.Write(cs.trieState); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeU64(from []byte) uint64 {
	var i uint64
	for _, b := range from {
		i = (i << 8) | uint64(b)
	}
	return i
}

func encodeU64(i uint64, to []byte) []byte {
	// writes i to b in big endian byte order, using the least number of bytes needed to represent i.
	switch {
	case i < (1 << 8):
		return append(to, byte(i))
	case i < (1 << 16):
		return append(to, byte(i>>8), byte(i))
	case i < (1 << 24):
		return append(to, byte(i>>16), byte(i>>8), byte(i))
	case i < (1 << 32):
		return append(to, byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
	case i < (1 << 40):
		return append(to, byte(i>>32), byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
	case i < (1 << 48):
		return append(to, byte(i>>40), byte(i>>32), byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
	case i < (1 << 56):
		return append(to, byte(i>>48), byte(i>>40), byte(i>>32), byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
	default:
		return append(to, byte(i>>56), byte(i>>48), byte(i>>40), byte(i>>32), byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
	}
}

// Optimised key referencing a state file record (file number and offset within the file)
func shortenedKey(apk []byte) (step uint16, offset uint64) {
	step = binary.BigEndian.Uint16(apk[:2])
	return step, decodeU64(apk[1:])
}
