package sema

import (
	"fmt"
	"sync"

	"shinya.click/cvm/common"
	"shinya.click/cvm/entity"
)

// Analyze 是公开的候选森林入口：先执行 PreFilter，再并发分析每棵保留下来的语法树。
// 返回唯一通过语义检查的 Program；若没有干净结果，则返回最有用的错误。
func Analyze(candidates []*entity.AstNode) (*Program, error) {
	return AnalyzeWithOptions(candidates, SemaOptions{})
}

func AnalyzeWithOptions(candidates []*entity.AstNode, opts SemaOptions) (*Program, error) {
	survivors, prefilterErrs := PreFilter(candidates)
	if len(survivors) == 0 {
		if len(prefilterErrs) > 0 {
			return nil, prefilterErrs[0]
		}
		return nil, fmt.Errorf("no candidates remain after PreFilter")
	}

	results := make([]*SemaResult, len(survivors))
	var wg sync.WaitGroup
	for i, tree := range survivors {
		wg.Add(1)
		go func(i int, tree *entity.AstNode) {
			defer wg.Done()
			results[i] = NewSemaWithOptions(opts).analyzeOne(tree)
		}(i, tree)
	}
	wg.Wait()

	var clean []*SemaResult
	for _, r := range results {
		if r != nil && len(r.Errors) == 0 {
			clean = append(clean, r)
		}
	}

	switch len(clean) {
	case 1:
		return clean[0].Program, nil
	case 0:
		best := pickBestErrorResult(results)
		if best != nil && len(best.Errors) > 0 {
			return nil, best.Errors[0]
		}
		if len(prefilterErrs) > 0 {
			return nil, prefilterErrs[0]
		}
		return nil, fmt.Errorf("no result and no errors recorded")
	default:
		return nil, ambiguousParse(clean)
	}
}

func pickBestErrorResult(results []*SemaResult) *SemaResult {
	var best *SemaResult
	for _, r := range results {
		if r == nil {
			continue
		}
		if best == nil {
			best = r
			continue
		}
		if len(r.Errors) < len(best.Errors) {
			best = r
			continue
		}
		if len(r.Errors) > len(best.Errors) {
			continue
		}
		if len(r.Errors) > 0 && len(best.Errors) > 0 && compareErrPos(r.Errors[0], best.Errors[0]) > 0 {
			best = r
		}
	}
	return best
}

func compareErrPos(a, b *common.CvmError) int {
	if len(a.Messages) == 0 || len(b.Messages) == 0 {
		return 0
	}
	pa, pb := a.Messages[0].SourcePos, b.Messages[0].SourcePos
	if pa.Line != pb.Line {
		return pa.Line - pb.Line
	}
	return pa.Column - pb.Column
}

func ambiguousParse(_ []*SemaResult) error {
	return fmt.Errorf("ambiguous parse: multiple candidates type-check cleanly")
}
