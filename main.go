package main

import (
	"bytes"
	"cmp"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type context struct {
	exit      int
	main      []*hclwrite.Block
	outputs   []*hclwrite.Block
	terraform []*hclwrite.Block
	variables []*hclwrite.Block
}

func main() {
	handleOptions("0.0.1")

	ctx := context{exit: 0}

	entries, err := os.ReadDir("./")
	if err != nil {
		panic(err)
	}

	specialNames := map[string]bool{
		"main.tf":      true,
		"outputs.tf":   true,
		"terraform.tf": true,
		"variables.tf": true,
	}

	for _, entry := range entries {
		filename := entry.Name()
		if strings.HasSuffix(filename, ".tf") {
			if !strings.HasPrefix(filename, "_") {
				_, specialName := specialNames[filename]

				fileBytes, err := os.ReadFile(filename)
				if err != nil {
					panic(err)
				}

				pos := hcl.InitialPos
				file, diag := hclwrite.ParseConfig(fileBytes, filename, pos)
				if diag.HasErrors() {
					panic("HasErrors")
				}

				blocks := []*hclwrite.Block{}
				for _, block := range file.Body().Blocks() {
					if block.Type() == "terraform" {
						ctx.terraform = append(ctx.terraform, block)
					} else if block.Type() == "variable" {
						ctx.variables = append(ctx.variables, block)
					} else if block.Type() == "output" {
						ctx.outputs = append(ctx.outputs, block)
					} else {
						if specialName {
							ctx.main = append(ctx.main, block)
						} else {
							blocks = append(blocks, block)
						}
					}
				}

				newFile := hclwrite.NewFile()

				sortBlocks(blocks)

				len := len(blocks)
				for i, block := range blocks {
					newFile.Body().AppendBlock(block)
					if i < len-1 {
						newFile.Body().AppendNewline()
					}
				}

				newBytes := newFile.Bytes()

				if !specialName {
					if !bytes.Equal(fileBytes, newBytes) {
						os.Stderr.WriteString(fmt.Sprintf("Main file state mismatch for \"%s\"\n", filename))
						ctx.exit = 1
					}
					if len == 0 {
						errR := os.Remove(filename)
						if errR != nil {
							panic(errR)
						}
					} else {
						errW := os.WriteFile(filename, newBytes, 0644)
						if errW != nil {
							panic(errW)
						}
					}
				}
			}
		}
	}

	if len(ctx.terraform) == 0 {
		blockLabels := []string{}
		newBlock := hclwrite.NewBlock("terraform", blockLabels)
		newValue := cty.StringVal(">=0.0.1")
		newBlock.Body().SetAttributeValue("required_version", newValue)
		ctx.terraform = append(ctx.terraform, newBlock)
	}

	writeTfFile(&ctx, "main.tf", ctx.main)
	writeTfFile(&ctx, "outputs.tf", ctx.outputs)
	writeTfFile(&ctx, "terraform.tf", ctx.terraform)
	writeTfFile(&ctx, "variables.tf", ctx.variables)
	os.Exit(ctx.exit)
}

func sortBlocks(blocks []*hclwrite.Block) {
	blockCmp := func(a *hclwrite.Block, b *hclwrite.Block) int {
		typeOrder := cmp.Compare(a.Type(), b.Type())
		if typeOrder != 0 {
			return typeOrder
		}

		aLabel0 := ""
		aLabels := a.Labels()
		if len(aLabels) >= 0 {
			aLabel0 = aLabels[0]
		}

		bLabels := b.Labels()
		bLabel0 := ""
		if len(bLabels) >= 0 {
			bLabel0 = bLabels[0]
		}
		label0Order := cmp.Compare(aLabel0, bLabel0)
		if label0Order != 0 {
			return label0Order
		}

		aLabel1 := ""
		bLabel1 := ""
		if len(aLabels) >= 1 {
			aLabel1 = aLabels[1]
		}

		if len(bLabels) >= 1 {
			bLabel1 = bLabels[1]
		}
		return cmp.Compare(aLabel1, bLabel1)
	}

	slices.SortFunc(blocks, blockCmp)
}

func writeTfFile(ctx *context, filename string, blocks []*hclwrite.Block) {
	fileBytes := []byte{}

	_, err := os.Stat(filename)
	if err == nil {
		fileBytes, err = os.ReadFile(filename)
		if err != nil {
			panic(err)
		}
	}

	if len(blocks) == 0 {
		// Create a enw file
		file, err := os.Create(filename)
		if err != nil {
			panic(err)
		}
		defer file.Close()
	} else {
		file := hclwrite.NewFile()

		sortBlocks(blocks)

		len := len(blocks)
		body := file.Body()
		for i, block := range blocks {
			body.AppendBlock(block)
			if i < len-1 {
				body.AppendNewline()
			}
		}

		newBytes := file.Bytes()

		if !bytes.Equal(fileBytes, newBytes) {
			os.Stderr.WriteString(fmt.Sprintf("File state mismatch for \"%s\"\n", filename))
			ctx.exit = 1
		}

		err := os.WriteFile(filename, newBytes, 0644)
		if err != nil {
			panic(err)
		}
	}
}

func handleOptions(version string) {
	if len(os.Args[1:]) > 0 {
		handleVersionOption(version)
	}
}

func handleVersionOption(version string) {
	if os.Args[1:][0] == "-v" || os.Args[1:][0] == "--version" {
		fmt.Println(version)
		os.Exit(0)
	} else {
		panic("Invalid parameter")
	}
}
