package pkg

import (
	"fmt"
	"strings"

	"github.com/ahmetb/go-linq/v3"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/spf13/afero"
)

var Fs = afero.NewOsFs()

// MergeBlocks merges the contents of srcBlock into destBlock.
func MergeBlocks(destBlock, srcBlock *hclwrite.Block) {
	// Merge attributes
	for name, attr := range srcBlock.Body().Attributes() {
		destBlock.Body().SetAttributeRaw(name, attr.Expr().BuildTokens(nil))
	}

	// Merge blocks
	srcBlocks := srcBlock.Body().Blocks()
	srcNestedBlockTypes := make(map[string]struct{})
	linq.From(srcBlocks).ForEach(func(i interface{}) {
		t := i.(*hclwrite.Block).Type()
		srcNestedBlockTypes[t] = struct{}{}
	})
	destBlocks := destBlock.Body().Blocks()
	for _, nb := range destBlocks {
		if _, conflict := srcNestedBlockTypes[nb.Type()]; conflict {
			destBlock.Body().RemoveBlock(nb)
		}
	}
	for _, nb := range srcBlocks {
		destBlock.Body().AppendBlock(nb)
	}
}

func MergeFiles(destFile, srcFile *hclwrite.File) {
	srcBlocks := srcFile.Body().Blocks()
	destBlocks := destFile.Body().Blocks()

	destBlockAddresses := make(map[string]*hclwrite.Block)
	for _, b := range destBlocks {
		destBlockAddresses[address(b)] = b
	}

	for _, srcBlock := range srcBlocks {
		destBlock, found := destBlockAddresses[address(srcBlock)]
		if !found {
			tokens := destFile.BuildTokens(nil)
			if tokens[len(tokens)-1].Type != hclsyntax.TokenNewline {
				destFile.Body().AppendNewline()
			}
			destFile.Body().AppendBlock(srcBlock)
		} else {
			MergeBlocks(destBlock, srcBlock)
		}
	}
}

func MergeFile(file1, file2, destFile string) error {
	content1, err := afero.ReadFile(Fs, file1)
	if err != nil {
		return fmt.Errorf("error reading source file: %+v", err)
	}

	content2, err := afero.ReadFile(Fs, file2)
	if err != nil {
		return fmt.Errorf("error reading destination file: %+v", err)
	}

	hclFile1, _ := hclwrite.ParseConfig(content1, file1, hcl.InitialPos)
	hclFile2, _ := hclwrite.ParseConfig(content2, file2, hcl.InitialPos)

	MergeFiles(hclFile2, hclFile1)

	err = afero.WriteFile(Fs, destFile, hclFile2.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("error writing destination file %s: %+v", destFile, err)
	}
	return nil
}

func address(block *hclwrite.Block) string {
	parts := append([]string{block.Type()}, block.Labels()...)
	return strings.Join(parts, ".")
}
