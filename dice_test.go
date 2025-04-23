package dice

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"
)

func Test_tokenize(t *testing.T) {
	type args struct {
		expression string
	}
	tests := []struct {
		name       string
		args       args
		wantTokens []token
	}{
		{
			name:       "single dice no repetition",
			args:       args{expression: "d20"},
			wantTokens: []token{"d20"},
		},
		{
			name:       "single dice with repetition",
			args:       args{expression: "1d20"},
			wantTokens: []token{"1d20"},
		},
		{
			name:       "repeat dice twice",
			args:       args{expression: "2d20"},
			wantTokens: []token{"2d20"},
		},
		{
			name:       "with coefficient",
			args:       args{expression: "2d20+5"},
			wantTokens: []token{"2d20", "+", "5"},
		},
		{
			name:       "two dice",
			args:       args{expression: "1d20+1d10"},
			wantTokens: []token{"1d20", "+", "1d10"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotTokens := tokenize(tt.args.expression); !reflect.DeepEqual(gotTokens, tt.wantTokens) {
				t.Errorf("tokenize() = %v, want %v", gotTokens, tt.wantTokens)
			}
		})
	}
}

func Test_createAST(t *testing.T) {
	type args struct {
		tokens []token
	}
	tests := []struct {
		name    string
		args    args
		want    ResultProvider
		wantErr bool
	}{
		{
			name: "d20",
			args: args{[]token{"d20"}},
			want: diceNode{faces: 20},
		},
		{
			name: "1d20",
			args: args{[]token{"1d20"}},
			want: diceNode{
				repetitions: 1,
				faces:       20,
			},
		},
		{
			name: "1d20+5",
			args: args{[]token{"1d20", "+", "5"}},
			want: arithmeticNode{
				operation: "+",
				left: diceNode{
					repetitions: 1,
					faces:       20,
				},
				right: numericNode{num: 5},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createAST(tt.args.tokens)
			if (err != nil) != tt.wantErr {
				t.Errorf("createAST() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createAST() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_token_toDice(t *testing.T) {
	tests := []struct {
		name    string
		tr      token
		want    diceNode
		wantErr bool
	}{
		{
			name: "single dice",
			tr:   "d20",
			want: diceNode{
				repetitions: 0,
				faces:       20,
			},
			wantErr: false,
		},
		{
			name: "single dice, one rep",
			tr:   "1d20",
			want: diceNode{
				repetitions: 1,
				faces:       20,
			},
			wantErr: false,
		},
		{
			name: "ten dice, one rep",
			tr:   "10d20",
			want: diceNode{
				repetitions: 10,
				faces:       20,
			},
			wantErr: false,
		},
		{
			name: "advantage",
			tr:   "2d20kh1",
			want: diceNode{
				repetitions: 2,
				faces:       20,
				keep:        1,
			},
			wantErr: false,
		},
		{
			name: "disadvantage",
			tr:   "2d20kl1",
			want: diceNode{
				repetitions: 2,
				faces:       20,
				keep:        -1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.tr.toDice()
			if (err != nil) != tt.wantErr {
				t.Errorf("token.toDice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("token.toDice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_diceNode_Result(t *testing.T) {
	type fields struct {
		repetitions int
		faces       int
		keep        int
	}
	tests := []struct {
		name     string
		fields   fields
		minValue int
		maxValue int
	}{
		{
			name: "single d20",
			fields: fields{
				repetitions: 1,
				faces:       20,
			},
			minValue: 1,
			maxValue: 20,
		},
		{
			name: "two d20",
			fields: fields{
				repetitions: 2,
				faces:       20,
			},
			minValue: 2,
			maxValue: 40,
		},
		{
			name: "two d1 2d1",
			fields: fields{
				repetitions: 2,
				faces:       1,
				keep:        0,
			},
			minValue: 1,
			maxValue: 2,
		},
		{
			name: "advantage 2d1kh1",
			fields: fields{
				repetitions: 2,
				faces:       1,
				keep:        -1,
			},
			minValue: 1,
			maxValue: 1,
		},
		{
			name: "keep highest 2 3d1kh2",
			fields: fields{
				repetitions: 3,
				faces:       1,
				keep:        -2,
			},
			minValue: 2,
			maxValue: 2,
		},
		{
			name: "disadvantage 2d1kl1",
			fields: fields{
				repetitions: 2,
				faces:       1,
				keep:        1,
			},
			minValue: 1,
			maxValue: 1,
		},
		{
			name: "keep lowest 2 3d1kl2",
			fields: fields{
				repetitions: 3,
				faces:       1,
				keep:        2,
			},
			minValue: 2,
			maxValue: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := diceNode{
				repetitions: tt.fields.repetitions,
				faces:       tt.fields.faces,
				keep:        tt.fields.keep,
			}
			got := n.Result()
			if (got.Value < tt.minValue) || (got.Value > tt.maxValue) {
				t.Errorf("diceNode.Result() = %v, want %v", got.Value, fmt.Sprintf("range %d-%d", tt.minValue, tt.maxValue))
			}

			pattern := fmt.Sprintf("%dd%d(?:kh\\d+|kl\\d+)?\\(\\d+(,\\d+)*\\)=>\\(\\d+(,\\d+)*\\)$", n.repetitions, n.faces)
			ok, err := regexp.MatchString(pattern, got.Details)
			if err != nil || !ok {
				t.Errorf("diceNode.Result() = %v, want pattern %v", got.Details, pattern)
			}
		})
	}
}

func Test_numericNode_Result(t *testing.T) {
	type fields struct {
		num int
	}
	tests := []struct {
		name   string
		fields fields
		want   Result
	}{
		{
			name:   "one",
			fields: fields{num: 1},
			want: Result{
				Value:   1,
				Details: "1",
			},
		},
		{
			name:   "twenty",
			fields: fields{num: 20},
			want: Result{
				Value:   20,
				Details: "20",
			},
		},
		{
			// should not be tokenizable like that, but technically valid
			name:   "minus one",
			fields: fields{num: -1},
			want: Result{
				Value:   -1,
				Details: "-1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := numericNode{
				num: tt.fields.num,
			}
			if got := n.Result(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("numericNode.Result() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_arithmeticNode_Result(t *testing.T) {
	type fields struct {
		right     ResultProvider
		left      ResultProvider
		operation token
	}
	tests := []struct {
		name   string
		fields fields
		want   Result
	}{
		{
			name: "one plus one",
			fields: fields{
				operation: "+",
				right:     numericNode{1},
				left:      numericNode{1},
			},
			want: Result{
				Value:   2,
				Details: "(1)+(1)",
			},
		},
		{
			name: "one plus two",
			fields: fields{
				operation: "+",
				right:     numericNode{1},
				left:      numericNode{2},
			},
			want: Result{
				Value:   3,
				Details: "(1)+(2)",
			},
		},
		{
			name: "two minus one",
			fields: fields{
				operation: "-",
				right:     numericNode{2},
				left:      numericNode{1},
			},
			want: Result{
				Value:   1,
				Details: "(2)-(1)",
			},
		},
		{
			name: "1d1 plus two",
			fields: fields{
				operation: "+",
				right: diceNode{
					faces:       1,
					repetitions: 1,
				},
				left: numericNode{2},
			},
			want: Result{
				Value:   3,
				Details: "(1d1(1)=>(1))+(2)",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := arithmeticNode{
				right:     tt.fields.right,
				left:      tt.fields.left,
				operation: tt.fields.operation,
			}
			if got := n.Result(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("arithmeticNode.Result() = %v, want %v", got, tt.want)
			}
		})
	}
}
