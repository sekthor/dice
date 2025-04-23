package dice

import (
	"reflect"
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
		want    node
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
