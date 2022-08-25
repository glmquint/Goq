package main

import (
    "fmt"
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

func main(){
    // swap(pair(a, b)) = pair(b, a)
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

	// foo(swap(pair(f(a), g(b))), swap(pair(q(c), z(d))))
    expr := Fun{"foo",
        []Expr{
            Fun{"swap",
                []Expr{
					Fun{"pair",
						[]Expr{
							Fun{"f", []Expr{Sym{"a"}}},
							Fun{"g", []Expr{Sym{"b"}}},
						},
					},
                },
            },
            Fun{"swap",
                []Expr{
					Fun{"pair",
						[]Expr{
							Fun{"q", []Expr{Sym{"c"}}},
							Fun{"z", []Expr{Sym{"d"}}},
						},
					},
                },
            },

        },
    }
	// swap(pair(f(c), g(d)))
	value := Fun{"swap",
		[]Expr{
			Fun{"pair",
				[]Expr{
					Fun{"f", []Expr{Sym{"c"}}},
					Fun{"g", []Expr{Sym{"d"}}},
				},
			},
		},
	}
	fmt.Println("swap:", swap)
	fmt.Println("expr:", expr)
	pattern := swap.head
	fmt.Println("pattern:", pattern)
	fmt.Println("value:", value)
	bind, err := pattern_match(pattern, value)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("important: ", bind)
	}
	expr_after := swap.apply_all(expr)
	println("\n\n")
	fmt.Println("swap=", swap)
	fmt.Println("expr=", expr)
	fmt.Println("res=", expr_after)
}
