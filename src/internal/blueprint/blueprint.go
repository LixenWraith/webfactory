package blueprint

import (
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type Block struct {
	Path  string
	Index []int
	ID    int
	Vars  map[string][]string
}

type Node struct {
	Block    Block
	Children []*Node
}

// New creates a blueprint tree from content
func New(content string) (*Node, error) {
	lines := strings.Split(content, "\n")
	blocks := make([]Block, 0, len(lines))
	id := 0
	var currentBlock *Block

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, ".") {
			if currentBlock == nil {
				continue
			}

			eqIndex := strings.IndexByte(line, '=')
			if eqIndex == -1 {
				continue
			}

			varName := strings.TrimSpace(line[:eqIndex])
			valueStart := eqIndex + 1
			for ; valueStart < len(line); valueStart++ {
				if !unicode.IsSpace(rune(line[valueStart])) {
					break
				}
			}
			value := line[valueStart:]

			if strings.HasPrefix(varName, ".") {
				name := varName[1:] // Remove the dot
				if _, exists := currentBlock.Vars[name]; !exists {
					currentBlock.Vars[name] = make([]string, 0)
				}
				currentBlock.Vars[name] = append(currentBlock.Vars[name], value)
			}
			continue
		}

		if block, ok := parseLine(line, id); ok {
			blocks = append(blocks, block)
			currentBlock = &blocks[len(blocks)-1]
			id++
		}
	}

	return buildTree(blocks), nil
}

func parseLine(line string, id int) (Block, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return Block{}, false
	}

	// Variable line
	if strings.HasPrefix(line, ".") {
		return Block{}, false
	}

	parts := strings.Fields(line)
	if len(parts) != 2 {
		return Block{}, false
	}

	indexStr := strings.Split(strings.TrimRight(parts[0], "."), ".")
	index := make([]int, 0, len(indexStr))
	for _, str := range indexStr {
		num, err := strconv.Atoi(str)
		if err != nil {
			return Block{}, false
		}
		index = append(index, num)
	}

	return Block{
		Path:  strings.TrimSpace(parts[1]),
		Index: index,
		ID:    id,
		Vars:  make(map[string][]string),
	}, true
}

func buildTree(blocks []Block) *Node {
	if len(blocks) == 0 {
		return nil
	}

	root := &Node{
		Block:    Block{ID: -1},
		Children: make([]*Node, 0),
	}

	nodeMap := make(map[string]*Node)
	nodeMap[""] = root

	indexKey := func(index []int) string {
		parts := make([]string, len(index))
		for i, v := range index {
			parts[i] = strconv.Itoa(v)
		}
		return strings.Join(parts, ".")
	}

	for _, block := range blocks {
		node := &Node{
			Block:    block,
			Children: make([]*Node, 0),
		}

		key := indexKey(block.Index)
		// Duplicate Index is not allowed
		if _, exists := nodeMap[key]; exists {
			return nil
		}
		nodeMap[key] = node

		if len(block.Index) > 0 {
			parentKey := indexKey(block.Index[:len(block.Index)-1])
			if parent, exists := nodeMap[parentKey]; exists {
				parent.Children = append(parent.Children, node)
			} else {
				root.Children = append(root.Children, node)
			}
		} else {
			root.Children = append(root.Children, node)
		}
	}

	var sortNodes func(*Node)
	sortNodes = func(node *Node) {
		if len(node.Children) > 0 {
			sort.Slice(node.Children, func(i, j int) bool {
				iIdx := node.Children[i].Block.Index
				jIdx := node.Children[j].Block.Index
				if node == root {
					return iIdx[0] < jIdx[0]
				}
				return iIdx[len(iIdx)-1] < jIdx[len(jIdx)-1]
			})
			for _, child := range node.Children {
				sortNodes(child)
			}
		}
	}
	sortNodes(root)

	return root
}