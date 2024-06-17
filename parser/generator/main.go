package main

import (
	"bufio"
	"bytes"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/iancoleman/strcase"
	"github.com/oleiade/lane/v2"
	"github.com/thoas/go-funk"
	"go/format"
	"golang.org/x/exp/slices"
	"html/template"
	"io"
	"log"
	"os"
	"shinya.click/cvm/common"
	"strconv"
	"strings"
)

func main() {
	lines := readLines()
	productions := genProductions(lines)
	checkProductions(productions)
	dfa := constructLR0(productions)
	addLookAheadSymbol(dfa, productions)
	//shaveDFA(dfa)
	//checkDFA(dfa)
	generateFile(dfa, productions)
}

func shaveDFA(dfa LRDFA) {
	// handle the ELSE shift-reduce conflict
	for _, node := range dfa.AllNodes {
		shiftELSE := false
		for edge := range node.Edges {
			if edge == "ELSE" {
				shiftELSE = true
				break
			}
		}
		if !shiftELSE {
			continue
		}
		problem := -1
		for i, item := range node.Items {
			if item.isReducible() && item.LookAhead == "ELSE" {
				problem = i
				break
			}
		}
		if problem != -1 {
			node.Items = append(node.Items[:problem], node.Items[problem+1:]...)
		}
	}
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
	// add dummy
	lines = append([]string{"dummy := program|1"}, lines...)
	var productions []*Production
	lastLeft := ""
	expectIdx := int64(1)
	for _, line := range lines {
		indexSplits := strings.Split(line, "|")
		if len(indexSplits) != 2 {
			log.Panicf("Invalid line format: %s", line)
		}
		idx, _ := strconv.ParseInt(indexSplits[1], 10, 64)
		splits := strings.Split(indexSplits[0], ":=")
		if len(splits) != 2 {
			log.Panicf("Invalid line format: %s", line)
		}
		left := strings.TrimSpace(splits[0])
		right := strings.Split(strings.TrimSpace(splits[1]), " ")
		if left != lastLeft {
			lastLeft = left
			expectIdx = 1
		}
		if expectIdx != idx {
			log.Panicf("Invalid line format: %s", line)
		}
		expectIdx++
		productions = append(productions, &Production{
			Left:  left,
			Index: idx,
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
			if common.IsTerminalSymbol(part) {
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
		Edges:  map[string][]*LRNode{},
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
				currentNode.Edges[s] = append(currentNode.Edges[s], existedNode)
				continue
			}
			nextItems := closure(nextKernel, productions)
			nNode := &LRNode{
				Kernel: nextKernel,
				Items:  nextItems,
				Edges:  map[string][]*LRNode{},
			}
			currentNode.Edges[s] = append(currentNode.Edges[s], nNode)
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
	}] = []string{string(common.EOF)}

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
				targetNodes := node.Edges[symbol]
				shifted := LRItem{
					Production: closureItem.Production,
					DotIndex:   closureItem.DotIndex + 1,
				}
				for _, targetNode := range targetNodes {
					targetSSLRItem := StateSpecificLRItem{Node: targetNode, Item: shifted}
					if closureItem.LookAhead == externalSymbol {
						propagateMap[stateSpecificKernelItem] = append(propagateMap[stateSpecificKernelItem], targetSSLRItem)
					} else {
						resultMap[targetSSLRItem] = append(resultMap[targetSSLRItem], closureItem.LookAhead)
					}
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
		if common.IsTerminalSymbol(left) {
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
		for symbol, nodes := range node.Edges {
			if len(nodes) != 0 {
				fmt.Println(symbol + " contain shift-shift conflicts!")
			}
			_, ok := symbols[symbol]
			if ok {
				fmt.Println(symbol + " contain shift-shift conflicts!")
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
				fmt.Println(lookahead + " contain shift-reduce conflicts!")
			}
			symbols[lookahead] = struct{}{}
		}
	}
}

func generateFile(dfa LRDFA, productions *Productions) {
	var nonTerminals []string
	set := mapset.NewSet[string]()
	productionIndex := map[*Production]int{}
	for i, prod := range productions.Productions {
		productionIndex[prod] = i
		var newRight []string
		for _, r := range prod.Right {
			if !common.IsTerminalSymbol(r) {
				newRight = append(newRight, strcase.ToCamel(r))
			} else {
				newRight = append(newRight, r)
			}
		}
		prod.Left = strcase.ToCamel(prod.Left)
		prod.Right = newRight
		if set.Contains(prod.Left) {
			continue
		}
		nonTerminals = append(nonTerminals, strcase.ToCamel(prod.Left))
		set.Add(prod.Left)
	}

	NodeStateMap := map[*LRNode]int{}
	for i, node := range dfa.AllNodes {
		NodeStateMap[node] = i
	}
	action := map[int]map[string][]DFAOperator{}
	gotos := map[int]map[string]int{}
	for i, node := range dfa.AllNodes {
		action[i] = map[string][]DFAOperator{}
		gotos[i] = map[string]int{}
		for symbol, lrNodes := range node.Edges {
			if !common.IsTerminalSymbol(symbol) {
				symbol = strcase.ToCamel(symbol)
			}
			for _, lrNode := range lrNodes {
				if common.IsTerminalSymbol(symbol) {
					action[i][symbol] = append(action[i][symbol], DFAOperator{
						OperatorType: SHIFT,
						StateIndex:   NodeStateMap[lrNode],
					})
				} else {
					gotos[i][symbol] = NodeStateMap[lrNode]
				}
			}
		}
		for _, item := range node.Items {
			if item.isReducible() {
				op := REDUCE
				if productionIndex[item.Production] == 1 {
					op = ACC
				}
				action[i][item.LookAhead] = append(action[i][item.LookAhead], DFAOperator{
					OperatorType: op,
					ReduceIndex:  productionIndex[item.Production],
				})
			}
		}
	}

	tmpl := `// Code generated by LALR Generator; DO NOT EDIT.
package glr

import (
	"shinya.click/cvm/common"
	"shinya.click/cvm/parser/entity"
)

const (
{{ range $_, $v := .NonTerminals }}	{{ $v }} common.TokenType = "{{ $v }}"
{{ end }}
)

var Productions = []entity.Production{
{{ range $_, $v := .Productions }}	{ Left: {{if isTerminal $v.Left}}common.{{end}}{{ $v.Left }}, Index: {{ $v.Index }}, Right: []common.TokenType{ {{ range $i, $r := $v.Right }}{{if $i}},{{end}}{{if isTerminal $r}}common.{{end}}{{ $r }}{{ end }} } },
{{ end }}
}

var LalrAction = map[int]map[common.TokenType][]entity.DFAOperator{
{{ range $index, $allOps := .Action }}	{{ $index }}: { {{ $first := true }}{{ range $k, $ops := $allOps }}{{if not $first}}, {{else}}{{$first = false}}{{end}}common.{{ $k }}: []entity.DFAOperator{ {{ $lfirst := true }}{{ range $_, $op := $ops }} {{if not $lfirst}}, {{else}}{{$lfirst = false}}{{end}} { {{ if eq $op.OperatorType 1 }}OperatorType: entity.SHIFT, StateIndex: {{ $op.StateIndex }}{{ else if eq $op.OperatorType 2 }}OperatorType: entity.REDUCE, ReduceIndex: {{ $op.ReduceIndex }}{{ else }}OperatorType: entity.ACC, ReduceIndex: {{ $op.ReduceIndex }}{{ end }} } {{end}} } {{end}}},
{{ end }}
}

var LalrGoto = map[int]map[common.TokenType]int{
{{ range $index, $gotos := .Goto }}	{{ $index }}: { {{ $first := true }}{{ range $k, $goto := $gotos }}{{if not $first}}, {{else}}{{$first = false}}{{end}}{{ $k }}: {{ $goto }}{{end}} },
{{ end }}
}
`
	data := struct {
		Productions  []*Production
		NonTerminals []string
		Action       map[int]map[string][]DFAOperator
		Goto         map[int]map[string]int
	}{
		Productions:  productions.Productions,
		NonTerminals: nonTerminals,
		Action:       action,
		Goto:         gotos,
	}
	file, err := os.OpenFile("./parser/glr/glr_table.go", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var buf bytes.Buffer
	err = template.Must(template.New("test").Funcs(template.FuncMap{"isTerminal": common.IsTerminalSymbol}).Parse(tmpl)).Execute(&buf, data)
	if err != nil {
		panic(err)
	}
	source, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}
	file.Write(source)
}
