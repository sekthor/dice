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
			name:       "single dice",
			args:       args{expression: "1d20"},
			wantTokens: []token{"1", "d20"},
		},
		{
			name:       "repeat dice",
			args:       args{expression: "2d20"},
			wantTokens: []token{"2", "d20"},
		},
		{
			name:       "with coefficient",
			args:       args{expression: "2d20+5"},
			wantTokens: []token{"2", "d20", "+", "5"},
		},
		{
			name:       "two dice",
			args:       args{expression: "1d20+1d10"},
			wantTokens: []token{"1", "d20", "+", "1", "d10"},
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

func Test_token_isDice(t *testing.T) {
	tests := []struct {
		name string
		tr   token
		want bool
	}{
		{
			name: "d20",
			tr:   "d20",
			want: true,
		},
		{
			name: "1d20",
			tr:   "1d20",
			want: false,
		},
		{
			name: "arithmetic token",
			tr:   "+",
			want: false,
		},
		{
			name: "numeric token",
			tr:   "1",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.isDice(); got != tt.want {
				t.Errorf("token.isDice() = %v, want %v", got, tt.want)
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
			args: args{[]token{"1", "d20"}},
			want: arithmeticNode{
				operation: "*",
				left:      numericNode{1},
				right:     diceNode{faces: 20},
			},
		},
		{
			name: "1d20+5",
			args: args{[]token{"1", "d20", "+", "5"}},
			want: arithmeticNode{
				operation: "+",
				left: arithmeticNode{
					operation: "*",
					left:      numericNode{1},
					right:     diceNode{faces: 20},
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
