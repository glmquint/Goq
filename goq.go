package main

import (
    "fmt"
	"unicode"
)

type Expr interface{
    String() string
    getName() string
    getArgs() []Expr
	isEqual(Expr) bool
}

///--										\
type Sym struct {
    name string
}
func (sym Sym)String() string{
    return sym.name
}
func (sym Sym)getName() string{
    return sym.String()
}
func (sym Sym)getArgs() []Expr{
    return nil
}
func (sym Sym)isEqual(other Expr) bool{
	switch other.(type){
	case Sym:
		return sym.String() == other.String()
	default:
		return false
	}
}
//\											/

///									\
type Fun struct {
    name string
    args []Expr
}
func (fun Fun)String() string {
    ret := fun.name + "("
    for i, a := range fun.args {
        if i > 0{
            ret += ", "
        }
        ret += a.String()
    }
    ret += ")"
    return ret
}
func (fun Fun)getName() string{
    return fun.name
}
func (fun Fun)getArgs() []Expr{
    return fun.args
}
func (fun Fun)isEqual(other Expr) bool{
	switch other.(type){
	case Fun:
		for i, arg := range fun.args{
			if !arg.isEqual(other.getArgs()[i]){
				return false
			}
		}
		return fun.String() == other.String()
	default:
		return false
	}
}
//\									/

type Bindings map[string]Expr

///															\
type Rule struct {
    head, body Expr
}
func (rule Rule)String() string{
    return rule.head.String() + " = " + rule.body.String()
}
func (rule *Rule)apply_all(expr Expr) Expr {
    bind, err := pattern_match(rule.head, expr)
    if err == nil{
		/*
		for key, elem := range bind {
			println(key, "=>", elem.String())
		}
		*/
		expr = substitute_bindings(bind, rule.body)
	} else {
		println(err.Error())
		switch expr.(type){
		case Sym:
			return expr
		case Fun:
			new_args := []Expr{}
			for _, arg := range expr.getArgs(){
				new_args = append(new_args, rule.apply_all(arg))
			}
			return Fun{expr.getName(), new_args}
		}
	}
	return expr
}
//\															/

func substitute_bindings(bindings Bindings, expr Expr) Expr {
	switch expr.(type){
	case Sym:
		if value, ok := bindings[expr.getName()]; ok{
			return value
		} else {
			return expr
		}
	case Fun:
		new_name := ""
		if value, ok := bindings[expr.getName()]; ok{
			switch value.(type){
			case Sym:
				new_name = value.getName()
			default:
				panic("Expected symbol in the place of the functor name")
			}
		} else {
			new_name =  expr.getName()
		}
		new_args := []Expr{}
		for _, arg := range expr.getArgs(){
			new_args = append(new_args, substitute_bindings(bindings, arg))
		}
		return Fun{new_name, new_args}
	}
	return expr
}

func pattern_match_impl(pattern Expr, value Expr, bindings Bindings) bool{
    switch pattern.(type){
        case Sym:
            if bound_value, ok := bindings[pattern.getName()]; ok{
				return bound_value.isEqual(value)
            } else {
                bindings[pattern.getName()] = value
                return true
            }
        case Fun:
            switch value.(type){
                case Fun:
                    pattern_name := pattern.getName()
                    pattern_args := pattern.getArgs()
                    value_name := value.getName()
                    value_args := value.getArgs()
                    if pattern_name == value_name && len(pattern_args) == len(value_args){
                        for i := 0; i < len(pattern_args); i++{
                            if !pattern_match_impl(pattern_args[i], value_args[i], bindings){
                                return false
                            }
                        }
                        return true
                    } else {
                        return false
                    }
                default:
                    return false
            }
    }
    return false
}

func pattern_match(pattern Expr, value Expr) (Bindings, error) {
	bindings := Bindings{}
    if pattern_match_impl(pattern, value, bindings){
        return bindings, nil
    }
	return nil, fmt.Errorf("Can't find pattern match: pattern='%s' over value='%s'", pattern.String(), value.String())
}

// === === ===

type tokenkind string

const (
	SYM tokenkind	= "SYM"
	OPENPAREN		= "OPENPAREN"
	CLOSEPAREN		= "CLOSEPAREN"
	COMMA			= "COMMA"
	EQUAL			= "EQUAL"
	INVALID			= "INVALID"
)

type Token struct {
	kind tokenkind
	text string
}
func (t Token)String() string {
	return "[" + string(t.kind) + ":" + t.text + "]"
}

type Lexer struct {
	text []rune
	cursor int
	current_chr rune
}
func (l *Lexer) fromStr(s string) *Lexer{
	l.text = []rune(s)
	//fmt.Printf("now l.text = %s\n", string(l.text))
	l.cursor = 0
	return l
}
func (l *Lexer) Iter () <-chan Token {
	ch := make(chan Token)
	go func (){
		for t, ok := l.generateToken(); ok; t, ok = l.generateToken() {
			ch <- t
		}
		close(ch)
	}()
	return ch
}
func (l *Lexer) generateToken() (Token, bool) {
	if !l.advance(){
		return Token{}, false
	}
	switch l.current_chr{
	case '(':
		return Token{OPENPAREN, ""}, true
	case ')':
		return Token{CLOSEPAREN, ""}, true
	case ',':
		return Token{COMMA, ""}, true
	case '=':
		return Token{EQUAL, ""}, true
	default:
		if rune(l.current_chr) == ' '{
			return l.generateToken()
		}
		sym_name := []rune{l.current_chr}
		for next_chr, ok := l.peek(0); ok; next_chr, ok = l.peek(0){
			//fmt.Printf("peeking %c: it is ", next_chr)
			if !unicode.IsLetter(next_chr) && !unicode.IsDigit(next_chr) {
				//fmt.Println("invalid because not alphanumeric")
				break
			}
			sym_name = append(sym_name, next_chr)
			//fmt.Printf("%c is valid. Now sym_name = %s\n", next_chr, string(sym_name))
			l.advance()
		}
		//fmt.Printf("saving sym_name = %s\n", string(sym_name))
		return Token{SYM, string(sym_name)}, true
	}
	panic("unreachable")
	return Token{INVALID, ""}, false
}
func (l *Lexer) advance() bool {
	//fmt.Print("advanced and ")
	if l.cursor >= len(l.text){
		//fmt.Println("we're over")
		return false
	}
	l.current_chr = l.text[l.cursor]
	l.cursor += 1
	//fmt.Printf("current_chr = %c, cursor/len = %d/%d\n", l.current_chr, l.cursor, len(l.text))
	return true
}
func (l *Lexer) peek(offset int) (rune, bool) {
	sum := l.cursor + offset 
	//fmt.Printf("peeking @ %d -> ", sum)
	if sum >= len(l.text) || sum < 0 {
		return ' ', false
	}
	return l.text[sum], true
}

func main(){
	lexer := Lexer{}
	lexed_tokens := []Token{}
	input := " swap( pair (a ,   b))  =  pair(b, a) "
	fmt.Println(input)
	for token := range lexer.fromStr(input).Iter() {
		lexed_tokens = append(lexed_tokens, token)
	}
	fmt.Println(lexed_tokens)
}
