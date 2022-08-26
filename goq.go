package main

import (
    "fmt"
	"unicode"
	"bufio"
	"os"
)

var r = bufio.NewReader(os.Stdin)
func step(msg string) {
	fmt.Println(msg)
	r.ReadLine()
}

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
                    //if len(pattern_args) == len(value_args){
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
type tokenkindset map[tokenkind]bool

const (
	SYM tokenkind	= "SYM"
	OPENPAREN		= "OPENPAREN"
	CLOSEPAREN		= "CLOSEPAREN"
	COMMA			= "COMMA"
	EQUAL			= "EQUAL"
	RULE			= "RULE"
	SHAPE			= "SHAPE"
	APPLY			= "APPLY"
	DONE			= "DONE"
	QUIT			= "QUIT"
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
	l.cursor = 0
	return l
}
func (l Lexer) peek_token(offset int) (Token, bool) {
	//step("peeking")
	t := Token{}
	ok := true
	for i := 0; i <= offset; i++ {
		t, ok = l.generateToken()
		if !ok {
			return Token{}, false
		}
	}
	//fmt.Printf("returning %s %s", t.kind, t.text)
	//step("")
	return t, true
}
func keyword_by_name(text string) (tokenkind, bool) {
	switch text {
	case "rule":
		return RULE, true
	case "shape":
		return SHAPE, true
	case "apply":
		return APPLY, true
	case "done":
		return DONE, true
	case "quit":
		return QUIT, true
	}
	return INVALID, false
}
func (l *Lexer) generateToken() (Token, bool) {
	if !l.advance(){
		return Token{}, false
	}
	switch l.current_chr{
	case '(':
		//step("generated '('")
		return Token{OPENPAREN, ""}, true
	case ')':
		//step("generated ')'")
		return Token{CLOSEPAREN, ""}, true
	case ',':
		//step("generated ','")
		return Token{COMMA, ""}, true
	case '=':
		//step("generated '='")
		return Token{EQUAL, ""}, true
	default:
		if rune(l.current_chr) == ' '{
			//step("skipping ' '")
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
		//fmt.Printf("generated sym : '%s", string(sym_name))
		//step("'")
		if kind, ok := keyword_by_name(string(sym_name)); ok {
			return Token{kind, string(sym_name)}, true
		} else {
			return Token{SYM, string(sym_name)}, true
		}
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

// === === ===

type Parser struct {
	/*
	token_list []Token
	current_token Token
	cursor int
	*/
}
func (p *Parser)parse(l *Lexer) Expr{
	current_token, _ := l.generateToken()
	//println("current token", current_token.kind, current_token.text)
	/*
	pt, ok := l.peek_token(0)
	if !ok {
		panic(ok)
	}
	println("sanity check: next is", pt.kind)
	*/
	if (current_token != Token{}) {
		switch current_token.kind{
		case SYM:
			_, ok := generate_if_kind(l, OPENPAREN)
			switch ok{
			case 1:
				args := []Expr{}
				if _, ok = generate_if_kind(l, CLOSEPAREN); ok == 1 {
					return Fun{current_token.text, args}
				}
				args = append(args, p.parse(l))
				for _, ok = generate_if_kind(l, COMMA); ok == 1; _, ok = generate_if_kind(l, COMMA) {
					args = append(args, p.parse(l))
				}
				if _, ok = generate_if_kind(l, CLOSEPAREN); ok != 1 {
					panic("Expected close paren")
				}
				//panic("parse functor arguments")
				return Fun{current_token.text, args}
			case 0:
				//println("parse symbol", current_token.text)
				//panic("")
				return Sym{current_token.text}
			default:
				panic("peeked in EOF")
			}
		case RULE:
			panic("got rule")
		default:
			panic("report expected symbol")
		}
	} else {
		panic("report EOF error")
	}
}

func generate_if_kind(l *Lexer, kind tokenkind) (Token, int) {
	peeked_token, ok := l.peek_token(0)
	if !ok{
		return Token{}, -1
	}
	//println("peeked token", peeked_token.text)
	//step("")
	if peeked_token.kind == kind {
		peeked_token, _ = l.generateToken() // doesn't change pt but advances lexer
		return peeked_token, 1
	}
	return peeked_token, 0
}


// === === ===

func expect_token_kind(l *Lexer, kinds tokenkindset) (Token, error) {
	token, ok := l.generateToken()
	if !ok {
		return Token{}, fmt.Errorf("Completely exhausted lexer")
	}
	_, ok = kinds[token.kind]
	if ok {
		return token, nil
	}
	return Token{}, fmt.Errorf("Unexpected Token %v", token.kind)
}

type Context struct {
	known_rules map[string]Rule
	current_expr Expr
	expected_tokens tokenkindset
	quit bool
}
func (c *Context)parse_cmd(l *Lexer) error {
	parser := Parser{}
	c.expected_tokens = tokenkindset{}
	c.expected_tokens[RULE] = true
	c.expected_tokens[SHAPE] = true
	c.expected_tokens[APPLY] = true
	c.expected_tokens[DONE] = true
	c.expected_tokens[QUIT] = true
	keyword, err := expect_token_kind(l, c.expected_tokens)
	if err != nil {
		panic(err)
	}
	switch keyword.kind {
	case RULE:
		expt := tokenkindset{}
		expt[SYM] = true
		name, err := expect_token_kind(l, expt)
		if err != nil {
			panic(err)
		}
		_, ok := c.known_rules[name.text]
		if ok {
			return fmt.Errorf("Rule %s already exists", name.text)
		}
		head := parser.parse(l)
		expt = tokenkindset{}
		expt[EQUAL] = true
		_, err = expect_token_kind(l, expt)
		if err != nil {
			panic(err)
		}
		body:= parser.parse(l)
		rule := Rule{head, body}
		c.known_rules[name.text] = rule
	case SHAPE:
		if c.current_expr != nil {
			return fmt.Errorf("Already Shaping")
		}
		expr := parser.parse(l)
		fmt.Printf("Shaping %s\n", expr)
		c.current_expr = expr
	case APPLY:
		if c.current_expr != nil {
			expt := tokenkindset{}
			expt[SYM] = true
			expt[RULE] = true
			name, err := expect_token_kind(l, expt)
			if err != nil {
				panic(err)
			}
			rule := Rule{}
			var new_expr Expr
			if name.kind == RULE {
				head := parser.parse(l)
				expt = tokenkindset{}
				expt[EQUAL] = true
				_, err = expect_token_kind(l, expt)
				if err != nil {
					panic(err)
				}
				body := parser.parse(l)
				rule = Rule{head, body}
			} else {
				var ok bool
				rule, ok = c.known_rules[name.text]
				if !ok {
					return fmt.Errorf("Rule %s does not exists", name.text)
				}
			}
			new_expr = rule.apply_all(c.current_expr)
			fmt.Println(new_expr)
			c.current_expr = new_expr
		} else {
			return fmt.Errorf("No Shaping in Place")
		}
	case DONE:
		if c.current_expr == nil {
			return fmt.Errorf("No Shaping in Place")
		}
		fmt.Printf("Finished shaping expression %s\n", c.current_expr)
		c.current_expr = nil
	case QUIT:
		c.quit = true
	default:
		panic("unreachable")
	}
	return nil
}

func handlepanic() {
	if a := recover(); a != nil {
		fmt.Println("RECOVER", a)
	} 
}
func mainloop(context *Context) {
	lexer := Lexer{}
	defer handlepanic()
	if context.current_expr != nil {
		fmt.Print("> ")
	} else {
		fmt.Print("known rules: \n")
		for rulename, rule := range context.known_rules {
			fmt.Printf("\t'%s': %v\n", rulename, rule)
		}
		//fmt.Print("current expr: ")
		//fmt.Println(context.current_expr)
		fmt.Print("goq> ")
	}
	input, _ := r.ReadString('\n')
	if len(input) > 1 {
		//res := swap.apply_all(parser.parse(lexer.fromStr(input)))
		res := context.parse_cmd(lexer.fromStr(input))
		if res != nil {
			fmt.Println("=> ", res)
		}
	}
}

func main(){
	swap := Rule{
        head : Fun{"swap",
                []Expr{
                    Fun{"pair",
                        []Expr{
                            Sym{"a"},
                            Sym{"b"},
                        },
                    },
                },
            },
        body : Fun{"pair",
                []Expr{
                    Sym{"b"},
                    Sym{"a"},
                },
            },
    }
	kr := map[string]Rule{}
	kr["swap"] = swap
	context := Context{kr, nil, tokenkindset{}, false}
	fmt.Println("Use 'quit' to exit")
	for !context.quit {
		mainloop(&context)
	}
	fmt.Println("Goodbye!")
}
