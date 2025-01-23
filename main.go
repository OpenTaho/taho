package main

import (
	"bufio"
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type context struct {
	exit      int
	level     int
	num       int
	main      []*hclwrite.Block
	outputs   []*hclwrite.Block
	providers []*hclwrite.Block
	tempDir   string
	terraform []*hclwrite.Block
	variables []*hclwrite.Block
	version   string
}

func main() {
	ctx1 := context{version: "0.0.19"}
	handleOptions(ctx1.version)

	run(&ctx1)
	if ctx1.exit == 1 {
		ctx2 := context{}
		ctx2.level = ctx1.level + 1
		run(&ctx2)
		if ctx2.exit != 0 {
			ctx1.exit = 2

			ctx3 := context{}
			ctx3.level = ctx2.level + 1
			run(&ctx3)
			if ctx3.exit != 0 {
				ctx1.exit = 3
			}
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
		ctx.tempDir = fmt.Sprintf(".terraform/taho/%d-%d", ctx.level, num(ctx))
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

				blocks = rewriteBlocks(ctx, blocks)

				len := len(blocks)
				newFileBody := newFile.Body()
				for i, block := range blocks {
					newFileBody.AppendBlock(block)
					if i < len-1 {
						newFileBody.AppendNewline()
					}
				}

				newBytes := newFile.Bytes()

				if !specialName {
					if !bytes.Equal(fileBytes, newBytes) {
						os.Stderr.WriteString(
							fmt.Sprintf(
								"File mismatch: \"%s\"\n", filename))
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

	if !strings.HasSuffix(ctx.version, "-0") {
		errR := os.RemoveAll(ctx.tempDir)
		if errR != nil {
			panic(errR)
		}
	}
}

func rewriteBlocks(ctx *context,
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
		block = rewriteBlock(ctx, block)
		newBlocks = append(newBlocks, block)
	}

	return newBlocks
}

func rewriteBlock(
	ctx *context, block *hclwrite.Block) *hclwrite.Block {

	detatchedNestedBlocks := make([]*hclwrite.Block, 0)
	bodyBlocks := block.Body().Blocks()
	lenBodyBlocks := len(bodyBlocks)
	for b := range bodyBlocks {
		newBlock := hclwrite.NewBlock(block.Type(), block.Labels())
		movedBlock := bodyBlocks[lenBodyBlocks-1-b]
		newBlock.Body().AppendBlock(movedBlock)
		block.Body().RemoveBlock(movedBlock)
		detatchedNestedBlocks = append(detatchedNestedBlocks, newBlock)
	}

	// The block now only has attributes and in this state we can sort the
	// attributes.
	block = cleanBlock(ctx, block)
	block = sortBlockAttributes(ctx, block)

	keys := getAttributeKeys(block)

	// Now we can reattach the nested blocks
	newLineNeeded := len(keys) > 0
	lenTempBlocks := len(detatchedNestedBlocks)
	blockBody := block.Body()
	for b := range detatchedNestedBlocks {
		tempBlock := cleanBlock(ctx, detatchedNestedBlocks[lenTempBlocks-1-b])
		tempBlocks := tempBlock.Body().Blocks()
		for tbb := range tempBlocks {
			if newLineNeeded {
				blockBody.AppendNewline()
			}
			nestedBlock := tempBlocks[tbb]
			nestedBlock = rewriteBlock(ctx, nestedBlock)
			blockBody.AppendBlock(nestedBlock)
			newLineNeeded = true
		}
	}

	return block
}

func sortBlockAttributes(
	ctx *context, block *hclwrite.Block) *hclwrite.Block {

	tempFilename1 := fmt.Sprintf(
		"%s/%d-%s.hcl", ctx.tempDir, num(ctx), block.Type())
	writeBlock(tempFilename1, block)

	keys := getAttributeKeys(block)

	slices.Sort(keys)

	lines := []string{}
	open, err := os.Open(tempFilename1)
	if err != nil {
		panic(err)
	}
	s := bufio.NewScanner(open)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		text := s.Text()
		lines = append(lines, text)
		if !strings.HasPrefix(text, "#") {
			break
		}
	}
	start := len(lines) - 1

	metaArguments := map[string]bool{
		"count":      true,
		"depends_on": true,
		"for_each":   true,
		"source":     true,
	}

	metaKeys := []string{}
	nonMetaKeys := []string{}
	for k := range keys {
		key := keys[k]

		_, metaArgument := metaArguments[key]
		if metaArgument {
			metaKeys = append(metaKeys, key)
		} else {
			nonMetaKeys = append(nonMetaKeys, key)
		}
	}

	keys = append(metaKeys, nonMetaKeys...)

	multiLineKeys := []string{}
	multiLineMetaKeys := []string{}
	singleLineKeys := []string{}
	singleLineMetaKeys := []string{}

	metaArgumentSection := false
	if len(keys) > 0 {
		metaArgumentSection = metaArguments[keys[0]]
	}
	for k1 := range keys {
		key := keys[k1]
		tempBlock := readBlock(tempFilename1)
		for k2 := range keys {
			if keys[k1] != keys[k2] {
				tempBlock.Body().RemoveAttribute(keys[k2])
			}
		}
		tempBlock = cleanBlock(ctx, tempBlock)
		tempFilename2 := fmt.Sprintf(
			"%s/%d-%s-%s.hcl", ctx.tempDir, num(ctx), block.Type(), key)
		writeBlock(tempFilename2, tempBlock)
		open, err := os.Open(tempFilename2)
		if err != nil {
			panic(err)
		}
		s := bufio.NewScanner(open)
		s.Split(bufio.ScanLines)
		lines2 := []string{}
		for s.Scan() {
			text := s.Text()
			lines2 = append(lines2, text)
		}
		lenLines2 := len(lines2)
		testLine := lines2[lenLines2-2]
		test := strings.Fields(testLine)
		isMultiLine := false
		if len(test) > 1 {
			if test[1] != "=" {
				isMultiLine = true
			}
		} else {
			isMultiLine = true
		}

		_, metaKey := metaArguments[key]
		if isMultiLine {
			if metaKey {
				multiLineMetaKeys = append(multiLineMetaKeys, key)
			} else {
				multiLineKeys = append(multiLineKeys, key)
			}
		} else {
			if metaKey {
				singleLineMetaKeys = append(singleLineMetaKeys, key)
			} else {
				singleLineKeys = append(singleLineKeys, key)
			}
		}
	}

	keys = append(singleLineMetaKeys, multiLineMetaKeys...)
	keys = append(keys, singleLineKeys...)
	keys = append(keys, multiLineKeys...)

	hasProcessedOneKey := false
	priorWasMultiLine := false
	metaArgumentSection = false
	if len(keys) > 0 {
		metaArgumentSection = metaArguments[keys[0]]
	}
	for k1 := range keys {
		key := keys[k1]
		_, metaArgument := metaArguments[key]

		if !metaArgument {
			if metaArgumentSection {
				lines = append(lines, "")
				metaArgumentSection = false
			}
		}

		tempBlock := readBlock(tempFilename1)
		for k2 := range keys {
			if keys[k1] != keys[k2] {
				tempBlock.Body().RemoveAttribute(keys[k2])
			}
		}
		tempBlock = cleanBlock(ctx, tempBlock)

		tempFilename2 := fmt.Sprintf(
			"%s/%d-%s-%s.hcl", ctx.tempDir, num(ctx), block.Type(), key)

		writeBlock(tempFilename2, tempBlock)

		open, err := os.Open(tempFilename2)
		if err != nil {
			panic(err)
		}
		s := bufio.NewScanner(open)
		s.Split(bufio.ScanLines)
		lines2 := []string{}
		for s.Scan() {
			text := s.Text()
			lines2 = append(lines2, text)
		}

		lenLines2 := len(lines2)
		testLine := lines2[lenLines2-2]
		test := strings.Fields(testLine)

		isMultiLine := false
		if len(test) > 1 {
			if test[1] != "=" {
				isMultiLine = true
			}
		} else {
			isMultiLine = true
		}

		if hasProcessedOneKey {
			if isMultiLine {
				lines = append(lines, "")
			} else {
				if priorWasMultiLine {
					lines = append(lines, "")
				}
			}
		}

		priorWasMultiLine = isMultiLine

		for n := range lines2 {
			if n > start && n < lenLines2-1 {
				lines = append(lines, lines2[n])
			}
		}

		hasProcessedOneKey = true
	}

	if !strings.HasSuffix(lines[start], "{}") {
		lines = append(lines, "}")
	}

	tempFilename3 := fmt.Sprintf(
		"%s/%d-put-%s.hcl", ctx.tempDir, num(ctx), block.Type())
	writeLines(tempFilename3, lines)
	block = readBlock(tempFilename3)
	writeDebugBlock(ctx, "get", block)
	return block
}

func writeDebugBlock(ctx *context, desc string, block *hclwrite.Block) {
	writeBlock(fmt.Sprintf("%s/%d-debug-%s.hcl", ctx.tempDir, num(ctx), desc), block)
}

func getAttributeKeys(block *hclwrite.Block) []string {
	keys := []string{}
	for key := range maps.Keys(block.Body().Attributes()) {
		keys = append(keys, key)
	}
	return keys
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

func cleanBlock(ctx *context,
	block *hclwrite.Block) *hclwrite.Block {

	tempName := block.Type()
	labels := block.Labels()
	for i := range labels {
		tempName += fmt.Sprintf("-%s", labels[i])
	}

	tempFilename1 := fmt.Sprintf(
		"%s/%d-%s.hcl", ctx.tempDir, num(ctx), tempName)
	writeBlock(tempFilename1, block)

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
		"%s/%d-%s.hcl", ctx.tempDir, num(ctx), tempName)

	writeLines(tempFilename2, lines)
	return readBlock(tempFilename2)
}

func writeLines(filename string, lines []string) {
	openFile, err := os.OpenFile(
		filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	writer := bufio.NewWriter(openFile)
	for line := range lines {
		lineText := lines[line]
		writer.WriteString(lineText)
		writer.WriteString("\n")
	}
	writer.Flush()
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
	block := file.Body().Blocks()[0]
	return block
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

		blocks = rewriteBlocks(ctx, blocks)

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
				fmt.Sprintf("Managed file mismatch: \"%s\"\n", filename))
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

func num(ctx *context) int {
	ctx.num++
	return ctx.num
}
