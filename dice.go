package dice

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"unicode"
)

// name types "kind", because lowercase type is reserved word
var (
	diceKind       = reflect.TypeOf(diceNode{})
	arithmeticKind = reflect.TypeOf(arithmeticNode{})
	numericKind    = reflect.TypeOf(numericNode{})
)

type token string

// abstract syntax tree (AST) node
// the root node of the tree will call child nodes' Result() function and
// compute own Result() from it
type ResultProvider interface {
	Result() Result
	kind() reflect.Type
}

type numericNode struct {
	num int
}

type arithmeticNode struct {
	right     ResultProvider
	left      ResultProvider
	operation token
}

type diceNode struct {
	repetitions int
	faces       int
	keep        int
}

type Result struct {
	Value   int
	Details string
}

func (n numericNode) Result() Result {
	return Result{Value: n.num, Details: fmt.Sprintf("%d", n.num)}
}

func (n arithmeticNode) Result() Result {
	right := n.right.Result()
	left := n.left.Result()

	switch n.operation {
	case "-":
		return Result{
			Value:   right.Value - left.Value,
			Details: fmt.Sprintf("(%s)-(%s)", right.Details, left.Details),
		}
	default:
		return Result{
			Value:   right.Value + left.Value,
			Details: fmt.Sprintf("(%s)+(%s)", right.Details, left.Details),
		}
	}
}

func (n diceNode) Result() Result {
	details := ""

	var rolls []int

	for range n.repetitions {
		roll := rand.Intn(n.faces) + 1

		if details != "" {
			details += ","
		}
		details += fmt.Sprintf("%d", roll)
		rolls = append(rolls, roll)
	}

	keep := ""
	if n.keep < 0 {
		keep = fmt.Sprintf("kl%d", (-1 * n.keep))
	} else if n.keep > 0 {
		keep = fmt.Sprintf("kh%d", n.keep)
	}

	details = fmt.Sprintf("%dd%d%s(%s)", n.repetitions, n.faces, keep, details)

	// kh1 => rolls[len(rolls)-1-1 : len(rolls)]
	// hl1 => rolls[0 : len(rolls)-1]
	sort.Ints(rolls)
	value := 0
	start := 0
	end := len(rolls)
	if n.keep != 0 {
		if n.keep < 0 {
			end += n.keep
		} else {
			start += n.keep
		}
	}

	selected := ""

	for _, roll := range rolls[start:end] {
		if selected != "" {
			selected += ","
		}
		selected += fmt.Sprintf("%d", roll)
		value += roll
	}

	details = fmt.Sprintf("%s=>(%s)", details, selected)

	return Result{
		Value:   value,
		Details: details,
	}
}

func (n arithmeticNode) kind() reflect.Type {
	return reflect.TypeOf(n)
}

func (n numericNode) kind() reflect.Type {
	return reflect.TypeOf(n)
}

func (n diceNode) kind() reflect.Type {
	return reflect.TypeOf(n)
}

func tokenize(expression string) []token {
	var tokens []token
	var currentToken token = ""
	previousTokenEnd := false
	currentTokenEnd := false

	for _, char := range expression {

		if unicode.IsSpace(char) {
			previousTokenEnd = true
		}

		if char == '-' || char == '+' {
			previousTokenEnd = true
			currentTokenEnd = true
		}

		if char == 'd' {
			for _, char := range currentToken {
				if char < '0' || char > '9' {
					previousTokenEnd = true
					break
				}
			}
		}

		if previousTokenEnd {
			tokens = append(tokens, currentToken)
			currentToken = ""
			previousTokenEnd = false
		}

		currentToken += token(char)

		if currentTokenEnd {
			tokens = append(tokens, currentToken)
			currentToken = ""
			currentTokenEnd = false
		}
	}

	if currentToken != "" {
		tokens = append(tokens, currentToken)
	}

	return tokens
}

func (t token) toNumeric() (numericNode, error) {
	num, err := strconv.Atoi(string(t))
	if err != nil {
		return numericNode{}, err
	}
	return numericNode{num: num}, nil
}

func (t token) toArithmetic() (arithmeticNode, error) {
	if t != "+" && t != "-" {
		return arithmeticNode{}, fmt.Errorf("'%s' is not valid arithmetic operation", t)
	}

	return arithmeticNode{
		operation: t,
	}, nil
}

func (t token) toDice() (diceNode, error) {
	var dice diceNode
	var err error
	var repetitionsLength int
	var dPosition int = -1
	var char rune

	for repetitionsLength, char = range t {
		if char == 'd' {
			dPosition = repetitionsLength
			break
		}
		if unicode.IsDigit(char) {
			continue
		}
		return dice, fmt.Errorf("dice token must start with 'd' rune or digit")
	}

	// if dPosition is still unset there was no 'd' rune in the string
	if dPosition == -1 {
		return dice, fmt.Errorf("dice token must contain 'd' rune")
	}

	if repetitionsLength > 0 {
		dice.repetitions, err = strconv.Atoi(string(t[:repetitionsLength]))
		if err != nil {
			return dice, fmt.Errorf("could not convert repetitions '%s' to int", t[:repetitionsLength])
		}
	}

	var facesLength int
	for facesLength, char = range t[repetitionsLength+1:] {
		if !unicode.IsDigit(char) {
			// at least one numeric digit is expected
			if facesLength == 0 {
				return dice, fmt.Errorf("dice token must specify face count")
			}
			facesLength--
			break
		}
	}
	facesLength++

	facesStart := dPosition + 1
	facesEnd := dPosition + 1 + facesLength
	dice.faces, err = strconv.Atoi(string(t[facesStart:facesEnd]))
	if err != nil {
		return dice, fmt.Errorf("could not convert repetitions '%s' to int", t[dPosition+1:dPosition+1+facesLength])
	}

	diceLength := repetitionsLength + 1 + facesLength
	if len(t) == diceLength {
		return dice, nil
	}

	// TODO: test the possibilities after faces digits
	if t[diceLength] == 'k' {

		// must be at least k(h | l)[0-9]
		if len(t[diceLength:]) < 3 {
			return dice, fmt.Errorf("could not evaluate kh/kl in token '%s'", t)
		}

		num, err := strconv.Atoi(string(t[diceLength+2:]))
		if err != nil {
			return dice, fmt.Errorf("the kh/hl parameter must be numeric")
		}

		if t[diceLength] == 'l' {
			num *= -1
		}

		dice.keep = num
	}

	return dice, nil
}

func (t token) toNode() (ResultProvider, error) {
	if node, err := t.toNumeric(); err == nil {
		return node, err
	}

	if node, err := t.toArithmetic(); err == nil {
		return node, err
	}

	if node, err := t.toDice(); err == nil {
		return node, err
	}

	return nil, fmt.Errorf("unknown token '%s'", t)
}

func createAST(tokens []token) (ResultProvider, error) {
	var astRoot ResultProvider
	for _, token := range tokens {

		node, err := token.toNode()
		if err != nil {
			return nil, err
		}

		switch node.kind() {

		case numericKind:
			if astRoot == nil {
				astRoot = node
				continue
			}
			if astRoot.kind() == arithmeticKind {
				tmp := astRoot.(arithmeticNode)
				tmp.right = node
				astRoot = tmp
				continue
			}
		case arithmeticKind:
			if astRoot == nil {
				return nil, fmt.Errorf("expression can not start with arithmetic token")
			}
			tmp := node.(arithmeticNode)
			tmp.left = astRoot
			astRoot = tmp
			continue
		case diceKind:
			if astRoot == nil {
				astRoot = node
				continue
			}
			if astRoot.kind() != arithmeticKind {
				return nil, fmt.Errorf("non starting dice token '%s' must be directly preceded by arithmetic token", token)
			}

			tmp := astRoot.(arithmeticNode)
			tmp.right = node
			astRoot = tmp

		default:
			return nil, nil
		}
	}
	return astRoot, nil
}

func Evaluate(expression string) (Result, error) {
	ast, err := createAST(tokenize(expression))
	if err != nil {
		return Result{}, err
	}
	return ast.Result(), nil
}
