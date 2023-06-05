package store

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/streamingfast/substreams/block"
)

var stateFileRegex = regexp.MustCompile(`([\d]+)-([\d]+)(?:\.([^\.]+))?\.(kv|partial)`)

type FileInfos []*FileInfo

func (f FileInfos) Ranges() (out block.Ranges) {
	if len(f) == 0 {
		return nil
	}

	out = make(block.Ranges, len(f))
	for i, file := range f {
		out[i] = file.Range
	}
	return
}

func (f FileInfos) String() string {
	ranges := make([]string, len(f))
	for i, file := range f {
		ranges[i] = file.Range.String()
	}

	return strings.Join(ranges, ",")
}

type FileInfo struct {
	Filename string
	Range    *block.Range
	TraceID  string
	Partial  bool
}

func NewCompleteFileInfo(moduleInitialBlock uint64, exlusiveEnd uint64) *FileInfo {
	bRange := block.NewRange(moduleInitialBlock, exlusiveEnd)

	return &FileInfo{
		Filename: FullStateFileName(bRange),
		Range:    block.NewRange(moduleInitialBlock, exlusiveEnd),
		Partial:  false,
	}
}

func NewPartialFileInfo(start uint64, exlusiveEnd uint64, traceID string) *FileInfo {
	bRange := block.NewRange(start, exlusiveEnd)

	return &FileInfo{
		Filename: PartialFileName(bRange, traceID),
		Range:    bRange,
		TraceID:  traceID,
		Partial:  true,
	}
}

func parseFileName(filename string) (*FileInfo, bool) {
	res := stateFileRegex.FindAllStringSubmatch(filename, 1)
	if len(res) != 1 {
		return nil, false
	}

	return &FileInfo{
		Filename: filename,
		Range:    block.NewRange(uint64(mustAtoi(res[0][2])), uint64(mustAtoi(res[0][1]))),
		TraceID:  res[0][3],
		Partial:  res[0][4] == "partial",
	}, true
}

func PartialFileName(r *block.Range, traceID string) string {
	if traceID == "" {
		// Generate legacy partial filename
		return fmt.Sprintf("%010d-%010d.partial", r.ExclusiveEndBlock, r.StartBlock)
	}

	return fmt.Sprintf("%010d-%010d.%s.partial", r.ExclusiveEndBlock, r.StartBlock, traceID)
}

func FullStateFileName(r *block.Range) string {
	return fmt.Sprintf("%010d-%010d.kv", r.ExclusiveEndBlock, r.StartBlock)
}

func mustAtoi(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}
