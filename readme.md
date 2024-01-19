# HCLMerge Tool

The `hclmerge` tool is a command-line utility written in Go that is used to merge two HashiCorp Configuration Language (HCL) configuration files. It uses the `hclwrite` package from Hashicorp to parse and manipulate the HCL files.

It should work like [Terraform's override file](https://developer.hashicorp.com/terraform/language/files/override) by following these rules:

* A top-level block in an override file merges with a block in a normal configuration file that has the same block header. The block header is the block type and any quoted labels that follow it.

* Within a top-level block, an attribute argument within an override block replaces any argument of the same name in the original block.

* Within a top-level block, any nested blocks within an override block replace all blocks of the same type in the original block. Any block types that do not appear in the override block remain from the original block.

* The contents of nested configuration blocks are not merged.

## Usage

The tool takes two input files and a destination file as arguments. It reads the contents of the two input files, merges them, and writes the merged content into the destination file.

```bash
hclmerge --file1 <source_file> --file2 <destination_file> --dest <merged_file>
```

Or:

```bash
hclmerge -1 <source_file> -2 <destination_file> -d <merged_file>
```

## Installation

To install the `hclmerge` tool, you need to have Go installed on your machine. Once Go is installed, you can use the `go get` command to install `hclmerge`.

```bash
go install github.com/lonegunmanb/hclmerge@latest
```

## Testing

The `hclmerge` tool comes with a suite of tests that verify its functionality. You can run these tests using the `go test` command.

```bash
go test github.com/lonegunmanb/hclmerge/...
```

## Contributing

Contributions to the `hclmerge` tool are welcome. Please make sure to read the contributing guide before making a pull request.

## License

The `hclmerge` tool is licensed under the MIT License. See the LICENSE file for more details.
