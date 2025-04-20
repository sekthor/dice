package dice

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"unicode"
)

type diceNode struct {
	faces int
}

type token string

type node interface {
	Value() int
	Type() reflect.Type
}

type numericNode struct {
	num int
}

type arithmeticNode struct {
	right     node
	left      node
	operation token
}

func (n numericNode) Value() int {
	return n.num
}

func (n numericNode) Type() reflect.Type {
	return reflect.TypeOf(n)
}

func (n arithmeticNode) Value() int {
	switch n.operation {
	case "-":
		return n.right.Value() + n.left.Value()
	default:
		return n.right.Value() + n.left.Value()
	}
}

func (n arithmeticNode) Type() reflect.Type {
	return reflect.TypeOf(n)
}

func (n diceNode) Value() int {
	return rand.Intn(n.faces) + 1
}

func (n diceNode) Type() reflect.Type {
	return reflect.TypeOf(n)
}

/* TODO
func Evaluate(expression string) {
	tokens := tokenize(expression)
	ast, error := createAST(tokens)
}
*/

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
			previousTokenEnd = true
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

func (t token) isDice() bool {
	if len(t) < 2 {
		return false
	}
	if t[0] != 'd' {
		return false
	}
	if _, err := strconv.Atoi(string(t[1:])); err != nil {
		return false
	}
	return true
}

func (t token) isArithmetic() bool {
	if len(t) != 1 {
		return false
	}

	if t == "+" || t == "-" || t == "*" {
		return true
	}

	return false
}

func (t token) isNumeric() bool {
	_, err := strconv.Atoi(string(t))
	return err == nil
}

func (t token) numeric() numericNode {
	num, _ := strconv.Atoi(string(t))
	return numericNode{num: num}
}

func (t token) arithmetic() arithmeticNode {
	return arithmeticNode{
		operation: t,
	}
}

func (t token) dice() diceNode {
	faces, _ := strconv.Atoi(string(t[1:]))
	return diceNode{faces: faces}
}

func createAST(tokens []token) (node, error) {
	var astRoot node
	for _, token := range tokens {

		var node node

		if token.isNumeric() {
			node = token.numeric()
		} else if token.isArithmetic() {
			node := token.arithmetic()
			node.left = astRoot
			astRoot = node
			continue
		} else if token.isDice() {
			node = token.dice()
		} else {
			return nil, fmt.Errorf("invalid token: %s", token)
		}

		if astRoot == nil {
			astRoot = node
			continue
		}

		if astRoot.Type() == reflect.TypeOf(arithmeticNode{}) {
			tmp, _ := astRoot.(arithmeticNode)
			tmp.right = node
			astRoot = tmp
			continue
		}

		if astRoot.Type() == reflect.TypeOf(numericNode{}) {
			parent := arithmeticNode{operation: "*"}
			parent.left = astRoot
			parent.right = node
			astRoot = parent
			continue
		}
	}
	return astRoot, nil
}
