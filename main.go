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
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

const version = "0.0.34"

type context struct {
	exit      int
	imports   []*hclwrite.Block
	level     int
	main      []*hclwrite.Block
	num       int
	out       *os.File
	checks    []*hclwrite.Block
	outputs   []*hclwrite.Block
	providers []*hclwrite.Block
	tempDir   string
	terraform []*hclwrite.Block
	variables []*hclwrite.Block
	version   string
}

func doMain(ctx *context) {
	run(ctx)
	if ctx.exit == 1 {
		ctx1 := context{}
		ctx1.level = ctx.level + 1
		run(&ctx1)
		if ctx1.exit != 0 {
			ctx.exit = 2
			ctx2 := context{}
			ctx2.level = ctx1.level + 1
			run(&ctx2)
			if ctx2.exit != 0 {
				ctx.exit = 3
			}
		}
	}
}

// Get the keys for attributes in the block
func getKeys(block *hclwrite.Block) []string {
	keys := []string{}
	for key := range maps.Keys(block.Body().Attributes()) {
		keys = append(keys, key)
	}
	return keys
}

func getTypeAndLabels(newBlock *hclwrite.Block) string {
	typeAndLabels := newBlock.Type()
	labels := newBlock.Labels()
	for n := range labels {
		typeAndLabels += fmt.Sprintf(".%s", labels[n])
	}
	return typeAndLabels
}

// Handle command line options
func handleOptions(ctx *context) {
	if len(os.Args[1:]) > 0 {
		if os.Args[1:][0] == "-v" || os.Args[1:][0] == "--version" {
			fmt.Println(version)
			os.Exit(0)
		} else if os.Args[1:][0] == "-r" || os.Args[1:][0] == "--recursive" {
			wd, err := os.Getwd()
			if err != nil {
				panic(err)
			}

			wd += "/"

			list := []string{}
			err = filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					panic(err)
				}
				path = wd + path
				if !strings.Contains(path, "/.") &&
					!strings.Contains(path, "/_") &&
					!strings.Contains(path, "/taho/") {
					if d.IsDir() {
						list = append(list, path)
					}
				}
				return nil
			})

			if err != nil {
				panic(err)
			}

			for n := range list {
				path := list[n]
				err := os.Chdir(path)
				if isTestable() {
					if n > 0 {
						out(ctx, "")
					}
					out(ctx, path)
					if path == "." {
						path = ""
					}
				}
				if err != nil {
					panic(err)
				}
				newCtx := context{version: version}
				newCtx.out = ctx.out
				doMain(&newCtx)
			}
		} else if os.Args[1:][0] == "-h" || os.Args[1:][0] == "--help" {
			fmt.Println(
				"Usage: taho [options]\n" +
					"\n" +
					"Options: are as follows\n" +
					"\n" +
					"-v, --version\n" +
					"-r, --recursive\n" +
					"-h, --help")
			os.Exit(0)
		} else {
			panic("Invalid parameter")
		}
	}
}

// Check if an array of lines fro a block is a multiline attribute.
//
// If the array of lines has a non-comment line size that equals 3 it is a
// a single line attribute.
func ifMultiline(lines []string) bool {
	size := 0
	for n := range lines {
		if !strings.HasPrefix(lines[n], "#") {
			size++
		}
	}
	return size != 3
}

// Check if the working directory is testable.
func isTestable() bool {
	entries, err := os.ReadDir("./")
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if entry.Type().IsRegular() {
			name := entry.Name()
			if strings.HasSuffix(name, ".tf") ||
				strings.HasSuffix(name, ".tfvars") {
				return true
			}
		}
	}
	return false
}

// Run this program up to three times.
//
// We consider the result successful if running the program does not create a
// change in the terraform files.
//
// In some cases, the program needs to run multiple times due to changes that
// define a state that needs subsequent changes.
func main() {
	ctx := context{version: version}
	ctx.out = os.Stdout
	handleOptions(&ctx)
	doMain(&ctx)
	os.Exit(ctx.exit)
}

func num(ctx *context) int {
	ctx.num++
	return ctx.num
}

func out(ctx *context, text string) {
	ctx.out.WriteString(text + "\n")
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

func readBlockX(temp string) (block *hclwrite.Block, err error) {
	tempBlock, err := readBlock(temp)
	if err != nil {
		return nil, err
	}
	return tempBlock, nil
}

func readLines(filename string, stopPrefix string) ([]string, int) {
	lines := []string{}
	open, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	s := bufio.NewScanner(open)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		text := s.Text()
		lines = append(lines, text)
		if !strings.HasPrefix(text, stopPrefix) {
			break
		}
	}
	start := len(lines) - 1
	return lines, start
}

// We remove comment lines that are at the end of the block because these cause
// problems with other areas of parsing.
func removeTrailingComments(ctx *context,
	block *hclwrite.Block, fixOtherComments bool, key string) *hclwrite.Block {

	temp := writeBlock(ctx, block)
	open, err := os.Open(temp)
	if err != nil {
		panic(err)
	}
	s := bufio.NewScanner(open)
	s.Split(bufio.ScanLines)
	lines := []string{}
	mode := 0

	// Remove blank lines because they cause problems with parsing
	blockPrefix := block.Type() + " "
	keyPrefix := "  " + key + " = "
	inspan := false
	inblock := false
	for s.Scan() {
		line := s.Text()
		if fixOtherComments {
			if strings.HasPrefix(line, blockPrefix) {
				fixOtherComments = key != ""
				inblock = true
			} else if strings.HasPrefix(line, keyPrefix) {
				fixOtherComments = false
			} else {
				trim := strings.TrimLeft(line, " ")

				if strings.HasPrefix(trim, "//") {
					line = "# " + strings.TrimPrefix(strings.TrimPrefix(trim, "//"), " ")
				} else if strings.HasPrefix(trim, "/*") {
					line = strings.TrimPrefix(trim, "/*")
					if line != "" {
						line = "# " + line
						if inblock {
							line = "  " + line
						}
					}
					inspan = true
				}

				if inspan {
					pos := strings.Index(line, "*/")
					if pos >= 0 {
						line = line[pos+2:]
						inspan = false
					} else {
						line = strings.TrimLeft(line, " ")
						line = strings.TrimPrefix(line, "* ")
						if line != "" {
							line = "# " + line
							if inblock {
								line = "  " + line
							}
						}
					}
				}
			}
		}
		if line != "" || mode > 1 {
			lines = append(lines, line)
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

	temp = fmt.Sprintf("%s/%d.hcl", ctx.tempDir, num(ctx))
	writeLines(temp, lines)
	block, err = readBlock(temp)
	if err != nil {
		panic(err)
	}
	return block
}

func rewriteBlock(
	ctx *context, block *hclwrite.Block, metaMode bool) *hclwrite.Block {

	detatchedNestedBlocks := make([]*hclwrite.Block, 0)
	bodyBlocks := block.Body().Blocks()
	lenBodyBlocks := len(bodyBlocks)
	blockTypes := []string{}
	for b := range bodyBlocks {
		newBlock := hclwrite.NewBlock(block.Type(), block.Labels())
		movedBlock := bodyBlocks[lenBodyBlocks-1-b]
		newBlock.Body().AppendBlock(movedBlock)
		block.Body().RemoveBlock(movedBlock)
		typeAndLabels := getTypeAndLabels(movedBlock)
		blockTypes = append(blockTypes, typeAndLabels)
		detatchedNestedBlocks = append(detatchedNestedBlocks, newBlock)
	}

	slices.Sort(blockTypes)
	blockTypes = slices.Compact(blockTypes)

	if metaMode {
		metaBlocks := map[string]bool{
			"connection":  true,
			"lifecycle":   true,
			"provisioner": true,
		}
		lastBlockTypes := len(blockTypes) - 1
		for b := range blockTypes {
			if b < lastBlockTypes {
				if metaBlocks[blockTypes[b]] {
					blockTypes[b] = blockTypes[b+1]
					blockTypes[b+1] = "lifecycle"
				}
			}
		}
	}

	block = sortAttributes(ctx, block, metaMode)

	keys := getKeys(block)
	newLineNeeded := len(keys) > 0

	lenNestedBlocks := len(detatchedNestedBlocks)
	blockBody := block.Body()

	for n := range blockTypes {
		orderedBlockType := blockTypes[n]
		for b := range detatchedNestedBlocks {
			detachedBlock := detatchedNestedBlocks[lenNestedBlocks-1-b]
			tempBlock := removeTrailingComments(ctx, detachedBlock, false, "")
			tempBlocks := tempBlock.Body().Blocks()
			for tbb := range tempBlocks {
				nestedBlock := tempBlocks[tbb]
				typeAndLabels := getTypeAndLabels(nestedBlock)
				if typeAndLabels == orderedBlockType {
					if newLineNeeded {
						blockBody.AppendNewline()
					}
					nestedBlock = rewriteBlock(ctx, nestedBlock, metaMode)
					blockBody.AppendBlock(nestedBlock)
					newLineNeeded = true
				}
			}
		}
	}

	return block
}

func rewriteBlocks(ctx *context, blocks []*hclwrite.Block,
	metaMode bool) []*hclwrite.Block {

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
		block = rewriteBlock(ctx, block, metaMode)
		newBlocks = append(newBlocks, block)
	}

	return newBlocks
}

func rewriteTfVars(ctx *context, filename string) {
	lines, _ := readLines(filename, "")

	mapLines := []string{"map {"}
	mapLines = append(mapLines, lines...)
	mapLines = append(mapLines, "}")

	temp1 := fmt.Sprintf("%s/%d.hcl", ctx.tempDir, num(ctx))
	writeLines(temp1, mapLines)
	mapBlock, err := readBlockX(temp1)
	if err != nil {
		return
	}

	mapBlock = rewriteBlock(ctx, mapBlock, false)

	temp2 := writeBlock(ctx, mapBlock)
	mapLines, _ = readLines(temp2, "")
	newLines := mapLines[1 : len(mapLines)-1]

	temp3 := fmt.Sprintf("%s/%d.tf", ctx.tempDir, num(ctx))
	writeLines(temp3, newLines)

	cmd := exec.Command("tofu", "fmt", temp3)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Run()

	fmtLines, _ := readLines(temp3, "")

	mismatch := false
	if len(fmtLines) == len(lines) {
		for n := range fmtLines {
			if fmtLines[n] != lines[n] {
				mismatch = true
				break
			}
		}
	} else {
		mismatch = true
	}

	if mismatch {
		text := fmt.Sprintf("File mismatch: \"%s\"", filename)
		out(ctx, text)
		writeLines(filename, fmtLines)
		setErrExit(ctx)
	}
}

// Run the program once.
func run(ctx *context) {
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
	}

	specialNames := map[string]bool{
		"checks.tf":    true,
		"imports.tf":   true,
		"main.tf":      true,
		"outputs.tf":   true,
		"providers.tf": true,
		"terraform.tf": true,
		"variables.tf": true,
	}

	entries, err := os.ReadDir("./")
	if err != nil {
		panic(err)
	}

	tf := false
	for _, entry := range entries {
		filename := entry.Name()
		if !strings.HasPrefix(filename, "_") {
			if strings.HasSuffix(filename, ".tfvars") {
				rewriteTfVars(ctx, filename)
			} else if strings.HasSuffix(filename, ".tf") {
				tf = true
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
						if block.Type() == "check" {
							ctx.checks = append(ctx.imports, block)
						} else if block.Type() == "import" {
							ctx.imports = append(ctx.imports, block)
						} else if block.Type() == "terraform" {
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
							text := fmt.Sprintf("File mismatch: \"%s\"", filename)
							out(ctx, text)
							setErrExit(ctx)
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
	}

	if tf {
		if len(ctx.terraform) == 0 {
			blockLabels := []string{}
			newBlock := hclwrite.NewBlock("terraform", blockLabels)
			newValue := cty.StringVal(">=0.0.1")
			newBlock.Body().SetAttributeValue("required_version", newValue)
			ctx.terraform = append(ctx.terraform, newBlock)
		}

		writeTfFile(ctx, false, "checks.tf", ctx.checks)
		writeTfFile(ctx, false, "import.tf", ctx.imports)
		writeTfFile(ctx, false, "outputs.tf", ctx.outputs)
		writeTfFile(ctx, false, "providers.tf", ctx.providers)
		writeTfFile(ctx, false, "variables.tf", ctx.variables)
		writeTfFile(ctx, true, "main.tf", ctx.main)
		writeTfFile(ctx, true, "terraform.tf", ctx.terraform)
	}

	if !strings.HasSuffix(ctx.version, "-0") {
		errR := os.RemoveAll(ctx.tempDir)
		if errR != nil {
			panic(errR)
		}
	}
}

func setErrExit(ctx *context) {
	if ctx.exit == 0 {
		ctx.exit = 1
	}
}

func sortAttributeKeys(keys []string, metaArguments map[string]bool) []string {
	slices.Sort(keys)
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
	return keys
}

func sortAttributes(
	ctx *context, block *hclwrite.Block, metaArgMode bool) *hclwrite.Block {

	metaArguments := map[string]bool{
		"count":      true,
		"depends_on": true,
		"for_each":   true,
		"provider":   true,
		"source":     true,
	}

	if block.Type() == "module" {
		metaArguments["version"] = true
	}

	if !metaArgMode {
		metaArguments = map[string]bool{}
	}

	keys := getKeys(block)
	keys = sortAttributeKeys(keys, metaArguments)

	block = removeTrailingComments(ctx, block, len(keys) == 0, "")
	temp := writeBlock(ctx, block)
	lines, start := readLines(temp, "#")

	multiLineKeys := []string{}
	multiLineMetaKeys := []string{}
	singleLineKeys := []string{}
	singleLineMetaKeys := []string{}

	for k1 := range keys {
		key := keys[k1]

		tempBlock, err := readBlockX(temp)
		if err != nil {
			panic(err)
		}

		for k2 := range keys {
			if keys[k1] != keys[k2] {
				tempBlock.Body().RemoveAttribute(keys[k2])
			}
		}

		tempBlock = removeTrailingComments(ctx, tempBlock, true, key)

		lines2, _ := readLines(writeBlock(ctx, tempBlock), "")
		isMultiLine := ifMultiline(lines2)
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

		tempBlock, err := readBlock(temp)
		if err != nil {
			panic(err)
		}
		for k2 := range keys {
			if keys[k2] != key {
				tempBlock.Body().RemoveAttribute(keys[k2])
			}
		}
		tempBlock = removeTrailingComments(ctx, tempBlock, true, keys[0])

		temp := writeBlock(ctx, tempBlock)
		lines2, _ := readLines(temp, "")
		isMultiLine := ifMultiline(lines2)

		if isMultiLine {
			lines4 := []string{}
			for n := range lines2 {
				line := lines2[n]

				if strings.HasPrefix(strings.TrimLeft(line, " "),
					fmt.Sprintf("%s = ", key)) {

					if strings.HasSuffix(line, "= {") {
						body := []string{"map {"}
						body = append(body, lines2[n+1:]...)
						body = body[:len(body)-2]
						body = append(body, "}")
						temp3 := fmt.Sprintf("%s/%d.hcl", ctx.tempDir, num(ctx))
						writeLines(temp3, body)
						mapBlock, err := readBlock(temp3)
						if err == nil {
							mapBlock = rewriteBlock(ctx, mapBlock, false)
							body, _ = readLines(writeBlock(ctx, mapBlock), "")
							n1 := 0
							lines3 := []string{}
							for {
								lines3 = append(lines3, lines2[n1])
								n1++
								if n1 > n {
									break
								}
							}
							lines3 = append(lines3, body[1:]...)
							lines3 = append(lines3, "}")
							temp5 := fmt.Sprintf("%s/%d.hcl", ctx.tempDir, num(ctx))
							writeLines(temp5, lines3)
							mapBlock, err = readBlock(temp5)
							if err == nil {
								lines4, _ = readLines(writeBlock(ctx, mapBlock), "")
							}
						}
					}
				}
			}

			if len(lines4) > 0 {
				lines2 = lines4
			}
		}

		if hasProcessedOneKey {
			if isMultiLine {
				lines = append(lines, "")
			}
		}

		for n := range lines2 {
			if n > start && n < len(lines2)-1 {
				lines = append(lines, lines2[n])
			}
		}

		hasProcessedOneKey = true
	}

	if !strings.HasSuffix(lines[start], "{}") {
		lines = append(lines, "}")
	}

	temp3 := fmt.Sprintf("%s/%d.hcl", ctx.tempDir, num(ctx))
	writeLines(temp3, lines)
	rewrittenBlock, err := readBlock(temp3)
	if err == nil {
		block = rewrittenBlock
	}
	return block
}

func writeBlock(ctx *context, block *hclwrite.Block) string {
	temp := fmt.Sprintf("%s/%d.hcl", ctx.tempDir, num(ctx))
	tempName := block.Type()
	tempFile := hclwrite.NewFile()
	tempFile.Body().AppendBlock(block)
	tempBytes := tempFile.Bytes()
	labels := block.Labels()
	for i := range labels {
		tempName += fmt.Sprintf("-%s", labels[i])
	}
	errW := os.WriteFile(temp, tempBytes, 0644)
	if errW != nil {
		panic(errW)
	}
	return temp
}

func writeLines(filename string, lines []string) {
	openFile, err := os.OpenFile(
		filename, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
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

func writeTfFile(ctx *context, always bool, filename string, blocks []*hclwrite.Block) {
	if !always {
		if len(blocks) == 0 {
			_, errS := os.Stat(filename)
			if errS == nil {
				errR := os.Remove(filename)
				if errR != nil {
					panic(errR)
				}
			}
			return
		}
	}

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
			text := fmt.Sprintf("Managed file mismatch: \"%s\"", filename)
			out(ctx, text)
			setErrExit(ctx)
		}

		err := os.WriteFile(filename, newBytes, 0644)
		if err != nil {
			panic(err)
		}
	}
}
