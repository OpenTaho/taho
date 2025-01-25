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
	main      []*hclwrite.Block
	num       int
	outputs   []*hclwrite.Block
	providers []*hclwrite.Block
	tempDir   string
	terraform []*hclwrite.Block
	variables []*hclwrite.Block
	version   string
}

func main() {
	ctx1 := context{version: "0.0.23"}
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

				blocks = rewriteBlocks(ctx, blocks, true)

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

func rewriteBlocks(ctx *context, blocks []*hclwrite.Block,
	metaArgMode bool) []*hclwrite.Block {

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
		block = rewriteBlock(ctx, block, metaArgMode)
		newBlocks = append(newBlocks, block)
	}

	return newBlocks
}

func rewriteBlock(
	ctx *context, block *hclwrite.Block,
	metaArgMode bool) *hclwrite.Block {

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

	block = cleanBlock(ctx, block)
	block = sortAttributes(ctx, block, metaArgMode)

	keys := getAttributeKeys(block)
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
			nestedBlock = rewriteBlock(ctx, nestedBlock, metaArgMode)
			blockBody.AppendBlock(nestedBlock)
			newLineNeeded = true
		}
	}

	return block
}

func sortAttributes(
	ctx *context, block *hclwrite.Block, metaArgMode bool) *hclwrite.Block {

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

	if !metaArgMode {
		metaArguments = map[string]bool{}
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

	for k1 := range keys {
		key := keys[k1]
		tempBlock, err := readBlock(tempFilename1)
		if err != nil {
			panic(err)
		}
		for k2 := range keys {
			if keys[k1] != keys[k2] {
				tempBlock.Body().RemoveAttribute(keys[k2])
			}
		}

		tempBlock = cleanBlock(ctx, tempBlock)

		tempFilename2 := fmt.Sprintf(
			"%s/%d-%s-%s.hcl", ctx.tempDir, num(ctx), block.Type(), key)
		writeBlock(tempFilename2, tempBlock)
		lines2 := readLines(tempFilename2)
		isMultiLine := checkIfMultiline(lines2)
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
	metaArgumentSection := false
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

		tempBlock, err := readBlock(tempFilename1)
		if err != nil {
			panic(err)
		}
		for k2 := range keys {
			if keys[k1] != keys[k2] {
				tempBlock.Body().RemoveAttribute(keys[k2])
			}
		}
		tempBlock = cleanBlock(ctx, tempBlock)

		tempFilename2 := fmt.Sprintf(
			"%s/%d-%s-%s.hcl", ctx.tempDir, num(ctx), block.Type(), key)

		writeBlock(tempFilename2, tempBlock)
		lines2 := readLines(tempFilename2)
		lenLines2 := len(lines2)
		isMultiLine := checkIfMultiline(lines2)

		if isMultiLine {
			for n := range lines2 {
				line := lines2[n]
				if strings.HasPrefix(strings.TrimLeft(line, " "), fmt.Sprintf("%s = ", key)) {
					if strings.HasSuffix(line, "= {") {
						body := []string{"map {"}
						body = append(body, lines2[n+1:]...)
						body = body[:len(body)-2]
						body = append(body, "}")
						tempFilename3 := fmt.Sprintf(
							"%s/%d-%s-%s.hcl", ctx.tempDir, num(ctx), block.Type(), key)
						writeLines(tempFilename3, body)
						mapBlock, err := readBlock(tempFilename3)
						if err == nil {
							mapBlock = rewriteBlock(ctx, mapBlock, false)
							tempFilename4 := fmt.Sprintf(
								"%s/%d-%s-%s.hcl", ctx.tempDir, num(ctx), block.Type(), key)
							writeBlock(tempFilename4, mapBlock)
							body = readLines(tempFilename4)
							body = append(lines2[:n+1], body[1:]...)
							body = body[:len(body)-1]
							tempFilename5 := fmt.Sprintf(
								"%s/%d-%s-%s.hcl", ctx.tempDir, num(ctx), block.Type(), key)
							body = append(body, lines2[lenLines2-2])
							body = append(body, "}")
							writeLines(tempFilename5, body)
							mapBlock, err = readBlock(tempFilename5)
							if err == nil {
								tempFilename6 := fmt.Sprintf(
									"%s/%d-%s-%s.hcl", ctx.tempDir, num(ctx), block.Type(), key)
								writeBlock(tempFilename6, mapBlock)
								lines2 = readLines(tempFilename6)
								lenLines2 = len(lines2)
							}
						}
					}
				}
			}
		}

		if hasProcessedOneKey {
			if isMultiLine {
				lines = append(lines, "")
			}
		}

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
	block, err = readBlock(tempFilename3)
	if err != nil {
		panic(err)
	}
	if strings.HasSuffix(ctx.version, "-0") {
		writeDebugBlock(ctx, "get", block)
	}
	return block
}

func readLines(filename string) []string {
	open, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	s := bufio.NewScanner(open)
	s.Split(bufio.ScanLines)
	lines := []string{}
	for s.Scan() {
		text := s.Text()
		lines = append(lines, text)
	}
	return lines
}

func checkIfMultiline(lines2 []string) bool {
	lenLines2 := len(lines2)
	testLine1 := lines2[lenLines2-2]
	testLine2 := strings.TrimLeft(lines2[lenLines2-3], " ")
	test1 := strings.Fields(testLine1)
	isMultiLine := false
	if len(test1) > 1 {
		if test1[1] == "=" {
			isMultiLine = strings.HasPrefix(testLine2, "#")
		} else {
			isMultiLine = true
		}
	} else {
		isMultiLine = true
	}
	return isMultiLine
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

// Cleans a block by writing to a file, processing, and then rereading.
//
// We remove comment lines that are at the end of the block because these cause
// problems with other areas of parsing.
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
	strippingComments := true
	for line := range lines {
		text := lines[lenLines-1-line]
		if strippingComments {
			trimmedText := strings.TrimLeft(text, " ")
			if strings.HasPrefix(trimmedText, "#") {
				text = ""
			} else {
				if trimmedText != "}" {
					strippingComments = false
				}
			}
		}
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
	block, err = readBlock(tempFilename2)
	if err != nil {
		panic(err)
	}
	return block
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

func readBlock(filename string) (*hclwrite.Block, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	file, diag := hclwrite.ParseConfig(fileBytes, filename, hcl.InitialPos)
	if diag.HasErrors() {
		return nil, errors.New("unparseable")
	}
	return file.Body().Blocks()[0], err
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

		blocks = rewriteBlocks(ctx, blocks, true)

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
