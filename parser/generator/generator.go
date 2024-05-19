package main

import (
	"bufio"
	"bytes"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/oleiade/lane/v2"
	"github.com/thoas/go-funk"
	"go/format"
	"golang.org/x/exp/slices"
	"html/template"
	"io"
	"log"
	"os"
	"shinya.click/cvm/common"
	"strings"
)

func main() {
	lines := readLines()
	productions := genProductions(lines)
	checkProductions(productions)
	dfa := constructLR0(productions)
	addLookAheadSymbol(dfa, productions)
	checkDFA(dfa)
	generateFile(dfa, productions)
}

func readLines() []string {
	file, err := os.OpenFile("./parser/phrase_structure_grammar.txt", os.O_RDONLY, 0666)
	if err != nil {
		log.Panicf("open file failed: %v", err.Error())
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	var results []string
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Panicf("read line failed: %v", err.Error())
		}
		if len(line) == 0 {
			continue
		}
		results = append(results, string(line))
	}
	return results
}

func genProductions(lines []string) *Productions {
	var productions []*Production
	for _, line := range lines {
		splits := strings.Split(line, ":=")
		if len(splits) != 2 {
			log.Panicf("Invalid line format: %s", line)
		}
		left := strings.TrimSpace(splits[0])
		right := strings.Split(strings.TrimSpace(splits[1]), " ")
		productions = append(productions, &Production{
			Left:  left,
			Right: right,
		})
	}
	return NewProductions(productions)
}

func checkProductions(prods *Productions) {
	nonTerminal := map[string]struct{}{}
	for _, prod := range prods.Productions {
		nonTerminal[prod.Left] = struct{}{}
	}
	for _, prod := range prods.Productions {
		for _, part := range prod.Right {
			if isTerminalSymbol(part) {
				continue
			}
			if _, ok := nonTerminal[part]; !ok {
				log.Panicf("Invalid nonTerminal: %s", part)
			}
		}
	}
}

func constructLR0(productions *Productions) LRDFA {
	startItem := LRItem{
		Production: productions.Productions[0],
		DotIndex:   0,
	}
	startKernel := []LRItem{startItem}
	startClosure := closure(startKernel, productions)

	var nodes []*LRNode
	stack := lane.NewStack[*LRNode]()
	startNode := &LRNode{
		Kernel: startKernel,
		Items:  startClosure,
		Edges:  map[string]*LRNode{},
	}

	nodes = append(nodes, startNode)
	stack.Push(startNode)

	for stack.Size() != 0 {
		currentNode, _ := stack.Pop()
		nextMap := map[string][]LRItem{}
		for _, item := range currentNode.Items {
			if item.isReducible() {
				continue
			}
			next := item.getSymbolAfterDot()
			nextMap[next] = append(nextMap[next], LRItem{
				Production: item.Production,
				DotIndex:   item.DotIndex + 1,
			})
		}

		for s, nextKernel := range nextMap {
			var existedNode *LRNode
			for _, node := range nodes {
				if node.hasSameKernel(nextKernel) {
					existedNode = node
					break
				}
			}
			if existedNode != nil {
				currentNode.Edges[s] = existedNode
				continue
			}
			nextItems := closure(nextKernel, productions)
			nNode := &LRNode{
				Kernel: nextKernel,
				Items:  nextItems,
				Edges:  map[string]*LRNode{},
			}
			currentNode.Edges[s] = nNode
			nodes = append(nodes, nNode)
			stack.Push(nNode)
		}
	}
	var reducibleNodes []*LRNode
	for _, node := range nodes {
		reducible := false
		for _, item := range node.Items {
			if item.isReducible() {
				reducible = true
				break
			}
		}
		if reducible {
			reducibleNodes = append(reducibleNodes, node)
		}
	}

	return LRDFA{
		StartState: startNode,
		EndStates:  reducibleNodes,
		AllNodes:   nodes,
	}
}

func addLookAheadSymbol(dfa LRDFA, productions *Productions) {
	propagateMap := map[StateSpecificLRItem][]StateSpecificLRItem{}
	resultMap := map[StateSpecificLRItem][]string{}

	resultMap[StateSpecificLRItem{
		Node: dfa.StartState,
		Item: dfa.StartState.Kernel[0],
	}] = []string{common.EOF}

	externalSymbol := "#"

	for _, node := range dfa.AllNodes {
		for _, item := range node.Kernel {
			stateSpecificKernelItem := StateSpecificLRItem{
				Node: node,
				Item: item,
			}

			// add external symbol to watch its propagate
			probeItem := LRItem{
				Production: item.Production,
				DotIndex:   item.DotIndex,
				LookAhead:  externalSymbol,
			}
			probeClosure := closure([]LRItem{probeItem}, productions)
			for _, closureItem := range probeClosure {
				if closureItem.isReducible() {
					continue
				}
				symbol := closureItem.getSymbolAfterDot()
				targetNode := node.Edges[symbol]
				shifted := LRItem{
					Production: closureItem.Production,
					DotIndex:   closureItem.DotIndex + 1,
				}
				targetSSLRItem := StateSpecificLRItem{Node: targetNode, Item: shifted}
				if closureItem.LookAhead == externalSymbol {
					propagateMap[stateSpecificKernelItem] = append(propagateMap[stateSpecificKernelItem], targetSSLRItem)
				} else {
					resultMap[targetSSLRItem] = append(resultMap[targetSSLRItem], closureItem.LookAhead)
				}
			}
		}
	}

	for {
		added := false
		for item, lookAheads := range resultMap {
			propagateTargetLookAheads := propagateMap[item]
			if len(propagateTargetLookAheads) == 0 {
				continue
			}
			for _, propagateTarget := range propagateTargetLookAheads {
				for _, lookAhead := range lookAheads {
					if slices.Contains(resultMap[propagateTarget], lookAhead) {
						continue
					}
					resultMap[propagateTarget] = append(resultMap[propagateTarget], lookAhead)
					added = true
				}
			}
		}

		if !added {
			break
		}
	}

	for _, node := range dfa.AllNodes {
		node.Kernel = []LRItem{}
		node.Items = []LRItem{}
	}
	for key, list := range resultMap {
		var newKernel []LRItem
		for _, symbol := range list {
			newKernel = append(newKernel, LRItem{
				Production: key.Item.Production,
				DotIndex:   key.Item.DotIndex,
				LookAhead:  symbol,
			})
		}
		newKernel = funk.Uniq(newKernel).([]LRItem)
		clo := closure(newKernel, productions)
		key.Node.Kernel = append(key.Node.Kernel, newKernel...)
		key.Node.Items = append(key.Node.Items, clo...)
	}
}

func closure(kernel []LRItem, productions *Productions) []LRItem {
	var result []LRItem
	for _, item := range kernel {
		result = append(result, item)
	}

	stack := lane.NewStack[LRItem]()
	for _, item := range kernel {
		stack.Push(item)
	}

	for stack.Size() != 0 {
		item, _ := stack.Pop()

		if item.isReducible() {
			continue
		}

		left := item.getSymbolAfterDot()
		if isTerminalSymbol(left) {
			continue
		}

		prods := productions.LeftMap[left]
		for _, prod := range prods {
			n := LRItem{Production: prod, DotIndex: 0}
			if len(item.LookAhead) == 0 {
				// cal lr0 closure
				if slices.Contains(result, n) {
					continue
				}
				result = append(result, n)
				stack.Push(n)
				continue
			}
			// cal lalr closure
			shifted := LRItem{
				Production: item.Production,
				DotIndex:   item.DotIndex + 1,
				LookAhead:  item.LookAhead,
			}
			lookaheadSymbols := mapset.NewSet[string]()
			if shifted.isReducible() {
				lookaheadSymbols.Add(shifted.LookAhead)
			} else {
				lookaheadSymbols.Append(productions.first(shifted.getSymbolAfterDot())...)
			}
			for _, x := range lookaheadSymbols.ToSlice() {
				newItem := LRItem{
					Production: n.Production,
					DotIndex:   n.DotIndex,
					LookAhead:  x,
				}
				if slices.Contains(result, newItem) {
					continue
				}
				result = append(result, newItem)
				stack.Push(newItem)
			}
		}
	}
	return result
}

func checkDFA(dfa LRDFA) {
	// check conflicts
	for _, node := range dfa.AllNodes {
		symbols := map[string]struct{}{}
		for symbol := range node.Edges {
			_, ok := symbols[symbol]
			if ok {
				for _, item := range node.Items {
					fmt.Println(item)
				}
				panic("a" + symbol)
			}
			symbols[symbol] = struct{}{}
		}
		for _, item := range node.Items {
			if !item.isReducible() {
				continue
			}
			lookahead := item.LookAhead
			_, ok := symbols[lookahead]
			if ok {
				for _, item := range node.Items {
					fmt.Println(item)
				}
				for symbol, _ := range node.Edges {
					fmt.Printf("%s :\n", symbol)
				}
				panic(lookahead)
			}
			symbols[lookahead] = struct{}{}
		}
	}
}

func generateFile(dfa LRDFA, productions *Productions) {
	productionIndex := map[*Production]int{}
	for i, prod := range productions.Productions {
		productionIndex[prod] = i
	}

	NodeStateMap := map[*LRNode]int{}
	for i, node := range dfa.AllNodes {
		NodeStateMap[node] = i
	}
	table := map[int]map[string]DFAOperator{}
	for i, node := range dfa.AllNodes {
		table[i] = map[string]DFAOperator{}
		for symbol, lrNode := range node.Edges {
			table[i][symbol] = DFAOperator{
				OperatorType: SHIFT,
				StateIndex:   NodeStateMap[lrNode],
			}
		}
		for _, item := range node.Items {
			if item.isReducible() {
				if productionIndex[item.Production] == 0 {
					table[i][item.LookAhead] = DFAOperator{
						OperatorType: ACC,
					}
					continue
				}
				table[i][item.LookAhead] = DFAOperator{
					OperatorType: REDUCE,
					ReduceIndex:  productionIndex[item.Production],
				}
			}
		}

	}

	tmpl := `package parser

var productions = []Production{
{{ range $_, $v := .Productions }}	{ Left: "{{ $v.Left }}", Right: []string{ {{ range $i, $r := $v.Right }}{{if $i}},{{end}}"{{ $r }}"{{ end }} } },
{{ end }}
}
`
	data := struct {
		Productions []*Production
	}{
		Productions: productions.Productions,
	}
	var buf bytes.Buffer
	template.Must(template.New("test").Parse(tmpl)).Execute(&buf, data)
	source, _ := format.Source(buf.Bytes())
	fmt.Println(string(source))
}
