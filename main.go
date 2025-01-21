package main

import (
	"bufio"
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"math/rand"
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
	providers []*hclwrite.Block
	run       int
	tempDir   string
	terraform []*hclwrite.Block
	variables []*hclwrite.Block
	version   string
}

func main() {
	ctx1 := context{version: "0.0.6"}
	handleOptions(ctx1.version)

	run(&ctx1)
	if ctx1.exit == 1 {
		ctx2 := context{run: 1}
		run(&ctx2)
		if ctx2.exit != 0 {
			ctx1.exit = 2
		}
	}

	os.Exit(ctx1.exit)
}

func run(ctx *context) {
	entries, err := os.ReadDir("./")
	if err != nil {
		panic(err)
	}

	for {
		ctx.tempDir = fmt.Sprintf(".terraform/taho/%d-%d", ctx.run, rand.Int())
		_, err := os.Stat(ctx.tempDir)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				errM := os.MkdirAll(ctx.tempDir, 0777)
				if errM != nil {
					panic(errM)
				}
			} else {
				panic(err)
			}
			break
		}
	}

	specialNames := map[string]bool{
		"main.tf":      true,
		"outputs.tf":   true,
		"providers.tf": true,
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
					} else if block.Type() == "provider" {
						ctx.providers = append(ctx.providers, block)
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

				blocks = rewriteBlocks(ctx, "run", blocks)

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
						os.Stderr.WriteString(
							fmt.Sprintf(
								"Main file state mismatch for \"%s\"\n", filename))
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

	writeTfFile(ctx, "main.tf", ctx.main)
	writeTfFile(ctx, "outputs.tf", ctx.outputs)
	writeTfFile(ctx, "terraform.tf", ctx.terraform)
	writeTfFile(ctx, "variables.tf", ctx.variables)

	if len(ctx.providers) > 0 {
		writeTfFile(ctx, "providers.tf", ctx.providers)
	} else {
		_, errS := os.Stat("providers.tf")
		if errS == nil {
			errR := os.Remove("providers.tf")
			if errR != nil {
				panic(errR)
			}
		}
	}

	errR := os.RemoveAll(ctx.tempDir)
	if errR != nil {
		panic(errR)
	}
}

func rewriteBlocks(ctx *context, stage string,
	blocks []*hclwrite.Block) []*hclwrite.Block {

	blockCmp := func(a *hclwrite.Block, b *hclwrite.Block) int {
		typeOrder := cmp.Compare(a.Type(), b.Type())
		if typeOrder != 0 {
			return typeOrder
		}

		aLabel0 := ""
		aLabels := a.Labels()
		if len(aLabels) > 0 {
			aLabel0 = aLabels[0]
		}

		bLabels := b.Labels()
		bLabel0 := ""
		if len(bLabels) > 0 {
			bLabel0 = bLabels[0]
		}
		label0Order := cmp.Compare(aLabel0, bLabel0)
		if label0Order != 0 {
			return label0Order
		}

		aLabel1 := ""
		bLabel1 := ""
		if len(aLabels) > 1 {
			aLabel1 = aLabels[1]
		}

		if len(bLabels) > 1 {
			bLabel1 = bLabels[1]
		}
		return cmp.Compare(aLabel1, bLabel1)
	}

	slices.SortFunc(blocks, blockCmp)

	newBlocks := make([]*hclwrite.Block, 0)
	for b := range blocks {
		block := blocks[b]
		newBlocks = rewriteBlock(ctx, stage, newBlocks, block)
	}

	return newBlocks
}

func rewriteBlock(
	ctx *context, stage string, newBlocks []*hclwrite.Block,
	block *hclwrite.Block) []*hclwrite.Block {

	tempBlocks := make([]*hclwrite.Block, 0)
	tempBlocks = append(tempBlocks, block)

	bodyBlocks := block.Body().Blocks()
	lenBodyBlocks := len(bodyBlocks)
	for b := range bodyBlocks {
		newBlock := hclwrite.NewBlock(block.Type(), block.Labels())
		movedBlock := bodyBlocks[lenBodyBlocks-1-b]
		newBlock.Body().AppendBlock(movedBlock)
		block.Body().RemoveBlock(movedBlock)
		tempBlocks = append(tempBlocks, newBlock)
	}

	// The block now only has attributes
	block = cleanBlock(ctx, stage+"-m", block)
	block = sortBlock(ctx, stage+"-s", block)

	blockBody := block.Body()
	for b := range tempBlocks {
		tempBlock := cleanBlock(ctx, stage, tempBlocks[b])
		tempBlocks := tempBlock.Body().Blocks()
		for tbb := range tempBlocks {
			blockBody.AppendNewline()
			blockBody.AppendBlock(cleanBlock(ctx, stage, tempBlocks[tbb]))
		}
	}

	return append(newBlocks, block)
}

func sortBlock(
	ctx *context, stage string, block *hclwrite.Block) *hclwrite.Block {

	tempFilename1 := fmt.Sprintf(
		"%s/%s-%s-1-%d.hcl", ctx.tempDir, block.Type(), stage, rand.Int())
	writeBlock(tempFilename1, block)

	keys := []string{}
	for key := range maps.Keys(block.Body().Attributes()) {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	for k1 := range keys {
		tempBlock := readBlock(tempFilename1)
		for k2 := range keys {
			if keys[k1] != keys[k2] {
				tempBlock.Body().RemoveAttribute(keys[k2])
			}
		}
		tempBlock = cleanBlock(ctx, stage, tempBlock)
		writeBlock(fmt.Sprintf("%s-%s.hcl", tempFilename1, keys[k1]), tempBlock)
	}

	return block
}

func writeBlock(filename string, block *hclwrite.Block) {
	tempName := block.Type()
	tempFile := hclwrite.NewFile()
	tempFile.Body().AppendBlock(block)
	tempBytes := tempFile.Bytes()
	labels := block.Labels()
	for i := range labels {
		tempName += fmt.Sprintf("-%s", labels[i])
	}
	errW := os.WriteFile(filename, tempBytes, 0644)
	if errW != nil {
		panic(errW)
	}
}

func cleanBlock(ctx *context, stage string,
	block *hclwrite.Block) *hclwrite.Block {

	tempFile := hclwrite.NewFile()
	tempFile.Body().AppendBlock(block)
	tempBytes := tempFile.Bytes()
	tempName := block.Type()
	labels := block.Labels()
	for i := range labels {
		tempName += fmt.Sprintf("-%s", labels[i])
	}
	tempFilename1 := fmt.Sprintf(
		"%s/%s-%s-1-%d.hcl", ctx.tempDir, tempName, stage, rand.Int())

	errW := os.WriteFile(tempFilename1, tempBytes, 0644)
	if errW != nil {
		panic(errW)
	}
	open, err := os.Open(tempFilename1)
	if err != nil {
		panic(err)
	}
	s := bufio.NewScanner(open)
	s.Split(bufio.ScanLines)
	lines := []string{}
	mode := 0
	for s.Scan() {
		text := s.Text()
		if text != "" || mode > 1 {
			lines = append(lines, text)
			mode++
		}
	}
	lenLines := len(lines)
	newLines := []string{}
	mode = 0
	for line := range lines {
		text := lines[lenLines-1-line]
		if text != "" || mode > 1 {
			newLines = append(newLines, text)
		}
	}

	lines = []string{}
	lenNewLines := len(newLines)
	for line := range newLines {
		text := newLines[lenNewLines-1-line]
		lines = append(lines, text)
	}

	tempFilename2 := fmt.Sprintf(
		"%s/%s-rewrite-2-%d.hcl", ctx.tempDir, tempName, rand.Int())

	openFile, err := os.OpenFile(
		tempFilename2, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

	if err != nil {
		panic(err)
	}
	writer := bufio.NewWriter(openFile)
	for line := range lines {
		writer.WriteString(lines[line])
		writer.WriteString("\n")
	}
	writer.Flush()
	return readBlock(tempFilename2)
}

func readBlock(filename string) *hclwrite.Block {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	file, diag := hclwrite.ParseConfig(fileBytes, filename, hcl.InitialPos)
	if diag.HasErrors() {
		panic(diag.Error())
	}
	return file.Body().Blocks()[0]
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
		file, err := os.Create(filename)
		if err != nil {
			panic(err)
		}
		defer file.Close()
	} else {
		file := hclwrite.NewFile()

		blocks = rewriteBlocks(ctx, "tf", blocks)

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
			os.Stderr.WriteString(
				fmt.Sprintf("File state mismatch for \"%s\"\n", filename))
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
