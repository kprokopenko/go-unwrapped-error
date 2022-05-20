package analyzer_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/kprokopenko/go-unwrapped-error/pkg/analyzer"
	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAll(t *testing.T) {
	//	t.Skip("TODO")
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get wd: %s", err)
	}

	testdata := filepath.Join(filepath.Dir(filepath.Dir(wd)), "testdata")
	analysistest.Run(t, testdata, analyzer.Analyzer, "p")
}

func TestCountUnwrappedReturns(t *testing.T) {
	src :=
		`func () error { 
return err
return nil
return someMethod("x")
return fmt.Errorf("x: %w", err)
return fmt.Errorf("x: %w", err)
return nil, fmt.Errorf("xxx", err)
return os.Open("test2")
return errors.New("abc")
 }`

	expr, err := parser.ParseExpr(src)
	if err != nil {
		t.Fatalf("parse expr: %s", err)
	}
	stmt := expr.(*ast.FuncLit)

	positions := analyzer.CountUnwrappedReturns(stmt.Body)
	got := make([]int, len(positions))
	for i := range positions {
		got[i] = lineByPos(src, positions[i])
	}

	want := []int{2, 4, 8}

	assert.Equal(t, want, got)
}

func lineByPos(src string, pos token.Pos) int {
	lines := 1
	for i := range src {
		if i > int(pos) {
			return lines
		}

		if src[i] == '\n' {
			lines++
		}
	}

	return lines
}
