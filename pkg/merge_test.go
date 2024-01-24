package pkg_test

import (
	"fmt"
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/lonegunmanb/hclmerge/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeBlocksAttributes(t *testing.T) {
	destHCL := `resource "aws_instance" "example" {
		ami = "ami-abc123"
	}`
	srcHCL := `resource "aws_instance" "example" {
		ami = "ami-def456"
	}`

	destFile, _ := hclwrite.ParseConfig([]byte(destHCL), "", hcl.InitialPos)
	srcFile, _ := hclwrite.ParseConfig([]byte(srcHCL), "", hcl.InitialPos)

	destBlock := destFile.Body().Blocks()[0]
	srcBlock := srcFile.Body().Blocks()[0]

	pkg.MergeBlocks(destBlock, srcBlock)
	cfg := destBlock.BuildTokens(nil).Bytes()
	syntaxCfg, diag := hclsyntax.ParseConfig(cfg, "", hcl.InitialPos)
	require.False(t, diag.HasErrors())
	value, diag := syntaxCfg.Body.(*hclsyntax.Body).Blocks[0].Body.Attributes["ami"].Expr.Value(&hcl.EvalContext{})
	require.False(t, diag.HasErrors())
	assert.Equal(t, "ami-def456", value.AsString())
}

func TestMergeBlocksNestedBlocks(t *testing.T) {
	destHCL := `resource "aws_instance" "example" {
		ebs_block_device {
			volume_size = 10
		}
	}`
	srcHCL := `resource "aws_instance" "example" {
		ebs_block_device {
			volume_size = 20
		}
	}`
	destFile, _ := hclwrite.ParseConfig([]byte(destHCL), "", hcl.InitialPos)
	srcFile, _ := hclwrite.ParseConfig([]byte(srcHCL), "", hcl.InitialPos)

	destBlock := destFile.Body().Blocks()[0]
	srcBlock := srcFile.Body().Blocks()[0]

	pkg.MergeBlocks(destBlock, srcBlock)
	newSize := strings.TrimSpace(string(destBlock.Body().Blocks()[0].Body().Attributes()["volume_size"].Expr().BuildTokens(nil).Bytes()))
	assert.Equal(t, "20", newSize)
}

func TestMergeBlocksNewAttribute(t *testing.T) {
	destHCL := `resource "aws_instance" "example" {
        ami = "ami-abc123"
    }`
	srcHCL := `resource "aws_instance" "example" {
        instance_type = "t2.micro"
    }`

	destFile, _ := hclwrite.ParseConfig([]byte(destHCL), "", hcl.InitialPos)
	srcFile, _ := hclwrite.ParseConfig([]byte(srcHCL), "", hcl.InitialPos)

	destBlock := destFile.Body().Blocks()[0]
	srcBlock := srcFile.Body().Blocks()[0]

	pkg.MergeBlocks(destBlock, srcBlock)

	cfg := destBlock.BuildTokens(nil).Bytes()
	syntaxCfg, diag := hclsyntax.ParseConfig(cfg, "", hcl.InitialPos)
	require.False(t, diag.HasErrors())

	// Check if the new attribute "instance_type" exists in the destBlock
	body := syntaxCfg.Body.(*hclsyntax.Body).Blocks[0].Body
	instanceType, diag := body.Attributes["instance_type"].Expr.Value(&hcl.EvalContext{})
	require.False(t, diag.HasErrors())
	assert.Equal(t, "t2.micro", instanceType.AsString())
	ami, diag := body.Attributes["ami"].Expr.Value(&hcl.EvalContext{})
	require.False(t, diag.HasErrors())
	assert.Equal(t, "ami-abc123", ami.AsString())
}

func TestMergeBlocksNewNestedBlock(t *testing.T) {
	destHCL := `resource "aws_instance" "example" {
        ami = "ami-abc123"
    }`
	srcHCL := `resource "aws_instance" "example" {
        ami = "ami-def456"
        ebs_block_device {
            volume_size = 20
        }
    }`

	destFile, _ := hclwrite.ParseConfig([]byte(destHCL), "", hcl.InitialPos)
	srcFile, _ := hclwrite.ParseConfig([]byte(srcHCL), "", hcl.InitialPos)

	destBlock := destFile.Body().Blocks()[0]
	srcBlock := srcFile.Body().Blocks()[0]

	pkg.MergeBlocks(destBlock, srcBlock)

	cfg := destBlock.BuildTokens(nil).Bytes()
	syntaxCfg, diag := hclsyntax.ParseConfig(cfg, "", hcl.InitialPos)
	require.False(t, diag.HasErrors())

	// Check if the new nested block "ebs_block_device" exists in the destBlock
	nestedBlock := syntaxCfg.Body.(*hclsyntax.Body).Blocks[0].Body.Blocks[0]
	require.Equal(t, "ebs_block_device", nestedBlock.Type)
	volumeSize, diag := nestedBlock.Body.Attributes["volume_size"].Expr.Value(&hcl.EvalContext{})
	require.False(t, diag.HasErrors())
	size, _ := volumeSize.AsBigFloat().Int64()
	assert.Equal(t, int64(20), size)
}

func TestMergeFiles_NewAttribute(t *testing.T) {
	destHCL := `resource "aws_instance" "example" {
        ami = "ami-abc123"
    }`
	srcHCL := `resource "aws_instance" "example" {
        instance_type = "t2.micro"
    }`

	destFile, _ := hclwrite.ParseConfig([]byte(destHCL), "", hcl.InitialPos)
	srcFile, _ := hclwrite.ParseConfig([]byte(srcHCL), "", hcl.InitialPos)

	pkg.MergeFiles(destFile, srcFile)
	destBlock := destFile.Body().Blocks()[0]

	cfg := destBlock.BuildTokens(nil).Bytes()
	syntaxCfg, diag := hclsyntax.ParseConfig(cfg, "", hcl.InitialPos)
	require.False(t, diag.HasErrors())

	// Check if the new attribute "instance_type" exists in the destBlock
	body := syntaxCfg.Body.(*hclsyntax.Body).Blocks[0].Body
	instanceType, diag := body.Attributes["instance_type"].Expr.Value(&hcl.EvalContext{})
	require.False(t, diag.HasErrors())
	assert.Equal(t, "t2.micro", instanceType.AsString())
	ami, diag := body.Attributes["ami"].Expr.Value(&hcl.EvalContext{})
	require.False(t, diag.HasErrors())
	assert.Equal(t, "ami-abc123", ami.AsString())
}

func TestMergeFiles_NewBlock(t *testing.T) {
	destHCL := `resource "aws_instance" "example" {
        ami = "ami-abc123"
    }`
	srcHCL := `resource "aws_instance" "example2" {
        instance_type = "t2.micro"
    }`

	destFile, _ := hclwrite.ParseConfig([]byte(destHCL), "", hcl.InitialPos)
	srcFile, _ := hclwrite.ParseConfig([]byte(srcHCL), "", hcl.InitialPos)

	pkg.MergeFiles(destFile, srcFile)
	cfg := destFile.Bytes()
	syntaxCfg, diag := hclsyntax.ParseConfig(cfg, "", hcl.InitialPos)
	require.False(t, diag.HasErrors())

	body := syntaxCfg.Body.(*hclsyntax.Body)
	assert.Equal(t, 2, len(body.Blocks))
	// Check if the new attribute "instance_type" exists in the destBlock
	exampleBody := body.Blocks[0].Body
	assert.Equal(t, 1, len(exampleBody.Attributes))
	ami, diag := exampleBody.Attributes["ami"].Expr.Value(&hcl.EvalContext{})
	require.False(t, diag.HasErrors())
	assert.Equal(t, "ami-abc123", ami.AsString())
	assert.Empty(t, exampleBody.Blocks)

	exampleBody2 := body.Blocks[1].Body
	assert.Equal(t, 1, len(exampleBody2.Attributes))
	assert.Empty(t, exampleBody2.Blocks)
	instanceType, diag := exampleBody2.Attributes["instance_type"].Expr.Value(new(hcl.EvalContext))
	require.False(t, diag.HasErrors())
	assert.Equal(t, "t2.micro", instanceType.AsString())
}

func TestMergeFile(t *testing.T) {
	overwrite := []bool{
		false,
		true,
	}
	for _, w := range overwrite {
		ow := w
		t.Run(fmt.Sprintf("overwrite %t", ow), func(t *testing.T) {
			// Create a new in-memory file system
			fs := afero.NewMemMapFs()

			// Create and write to source and destination files
			_ = afero.WriteFile(fs, "/src.hcl", []byte(`resource "aws_instance" "example" {
		ami = "ami-abc123"
	}`), 0644)
			_ = afero.WriteFile(fs, "/dest.hcl", []byte(`resource "aws_instance" "example" {
		instance_type = "t2.micro"
	}`), 0644)
			if ow {
				_ = afero.WriteFile(fs, "/merged.hcl", []byte(""), 0644)
			}
			stub := gostub.Stub(&pkg.Fs, fs)
			defer stub.Reset()

			// Call the MergeFile function
			err := pkg.MergeFile("/src.hcl", "/dest.hcl", "/merged.hcl")
			require.NoError(t, err)

			// Read the merged file
			mergedContent, _ := afero.ReadFile(fs, "/merged.hcl")

			// Check the content of the merged file
			assert.Contains(t, string(mergedContent), `ami           = "ami-abc123"`)
			assert.Contains(t, string(mergedContent), `instance_type = "t2.micro"`)
		})
	}
}

func TestMergeFile_EmptyDestFileShouldPrintMergedFile(t *testing.T) {
	// Create a new in-memory file system
	fs := afero.NewMemMapFs()

	// Create and write to source and destination files
	_ = afero.WriteFile(fs, "/src.hcl", []byte(`resource "aws_instance" "example" {
        ami = "ami-abc123"
    }`), 0644)
	_ = afero.WriteFile(fs, "/dest.hcl", []byte(`resource "aws_instance" "example" {
        instance_type = "t2.micro"
    }`), 0644)
	stub := gostub.Stub(&pkg.Fs, fs)
	defer stub.Reset()

	r, w, _ := os.Pipe()
	stub.Stub(&os.Stdout, w)

	// Call the MergeFile function
	err := pkg.MergeFile("/src.hcl", "/dest.hcl", "")
	require.NoError(t, err)

	// Close the writer and restore stdout
	w.Close()

	// Read the buffer which has stdout
	out, _ := io.ReadAll(r)

	// Check the content of the stdout
	assert.Contains(t, string(out), `ami           = "ami-abc123"`)
	assert.Contains(t, string(out), `instance_type = "t2.micro"`)
}
