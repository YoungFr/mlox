package interpreter

import (
	"fmt"
	"strconv"
	"time"

	"github.com/youngfr/mlox/parser"
	"github.com/youngfr/mlox/token"
)

// functions global-environment
var globals *environment = NewEnvironment(nil)

func init() {
	globals.def("date", &date{})
	globals.def("clock", &clock{})
}

/******************** the date builtin function ********************/

type date struct{}

func (d *date) arity() int {
	return 0
}

func (d *date) call(interpreter *Interpreter, arguments []any) any {
	year, month, day := time.Now().Date()
	return fmt.Sprintf("%4d-%03s-%02d", year, month.String()[:3], day)
}

func (d *date) String() string {
	return "<native fn date>"
}

/******************** the clock builtin function ********************/

type clock struct{}

func (c *clock) arity() int {
	return 0
}

func (c *clock) call(interpreter *Interpreter, arguments []any) any {
	hour, min, sec := time.Now().Clock()
	return fmt.Sprintf("%02d:%02d:%02d", hour, min, sec)
}

func (c *clock) String() string {
	return "<native fn clock>"
}

type Interpreter struct {
	env *environment
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		env: globals,
	}
}

func (i *Interpreter) Interpret(statements []parser.Stmt) {
	for _, stmt := range statements {
		i.exec(stmt)
	}
}

// evaluating expression
func (i *Interpreter) eval(expr parser.Expr) any {
	return expr.Accept(i)
}

// The *Interpreter should implement the ExprVisitor interface.
var _ = parser.ExprVisitor(NewInterpreter())

/********** The implementation of the ExprVisitor interface **********/

func (i *Interpreter) VisitorLiteralExpr(l *parser.Literal) any {
	// literal expression like `1` `true` `"str"` ...
	// We simply return it literal vlaue.
	return l.Value
}

func (i *Interpreter) VisitorUnaryExpr(u *parser.Unary) any {
	// unary expression like `-2` `--2` `!(1 < 2)` `!!true` ...
	ropreand := i.eval(u.Ropreand)
	switch u.Operator.Ttype {
	case token.SUB:
		if f, ok := ropreand.(float64); ok {
			return -f
		}
	case token.NOT:
		return !isTruthy(ropreand)
	}
	return nil
}

// In Lox, only the nil and false are false.
func isTruthy(a any) bool {
	if a == nil {
		return false
	}
	if b, ok := a.(bool); ok {
		return b
	}
	return true
}

// E1 op E2
// op -> '+' | '-' | '*' | '/' | '>' | '>=' | '<' | '<=' | '==' | '!='
func (i *Interpreter) VisitorBinaryExpr(b *parser.Binary) any {
	lopreand := i.eval(b.Lopreand)
	roperand := i.eval(b.Ropreand)
	switch b.Operator.Ttype {
	case token.ADD:
		lf, okl := lopreand.(float64)
		rf, okr := roperand.(float64)
		if okl && okr {
			return lf + rf
		}
		ls, okl := lopreand.(string)
		rs, okr := roperand.(string)
		if okl && okr {
			return ls + rs
		}
		panic("the binary operator '+' can only be used for numbers or strings")
	case token.SUB:
		lf, okl := lopreand.(float64)
		rf, okr := roperand.(float64)
		if okl && okr {
			return lf - rf
		}
	case token.MUL:
		lf, okl := lopreand.(float64)
		rf, okr := roperand.(float64)
		if okl && okr {
			return lf * rf
		}
	case token.DIV:
		lf, okl := lopreand.(float64)
		rf, okr := roperand.(float64)
		if okl && okr {
			return lf / rf
		}
	case token.GTR:
		lf, okl := lopreand.(float64)
		rf, okr := roperand.(float64)
		if okl && okr {
			return lf > rf
		}
		ls, okl := lopreand.(string)
		rs, okr := roperand.(string)
		if okl && okr {
			return ls > rs
		}
	case token.GEQ:
		lf, okl := lopreand.(float64)
		rf, okr := roperand.(float64)
		if okl && okr {
			return lf >= rf
		}
		ls, okl := lopreand.(string)
		rs, okr := roperand.(string)
		if okl && okr {
			return ls >= rs
		}
	case token.LSS:
		lf, okl := lopreand.(float64)
		rf, okr := roperand.(float64)
		if okl && okr {
			return lf < rf
		}
		ls, okl := lopreand.(string)
		rs, okr := roperand.(string)
		if okl && okr {
			return ls < rs
		}
	case token.LEQ:
		lf, okl := lopreand.(float64)
		rf, okr := roperand.(float64)
		if okl && okr {
			return lf <= rf
		}
		ls, okl := lopreand.(string)
		rs, okr := roperand.(string)
		if okl && okr {
			return ls <= rs
		}
	case token.EQL:
		return isEqual(lopreand, roperand)
	case token.NEQ:
		return !isEqual(lopreand, roperand)
	}
	return nil
}

func isEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil {
		return false
	}
	return a == b
}

func (i *Interpreter) VisitorGroupExpr(g *parser.Group) any {
	// For group expression, simply evaluating its expression field.
	return i.eval(g.Expression)
}

func (i *Interpreter) VisitorVariableExpr(v *parser.Variable) any {
	return i.env.get(v.Name)
}

func (i *Interpreter) VisitorAssignExpr(a *parser.Assign) any {
	value := i.eval(a.Value)
	i.env.asg(a.Name, value)
	return value
}

func (i *Interpreter) VisitorLogicalExpr(l *parser.Logical) any {
	lopreand := i.eval(l.Lopreand)
	// short circuit evaluation
	if l.Operator.Ttype == token.OR {
		if isTruthy(lopreand) {
			return lopreand
		}
	} else {
		if !isTruthy(lopreand) {
			return lopreand
		}
	}
	return i.eval(l.Ropreand)
}

func (i *Interpreter) VisitorCallExpr(c *parser.Call) any {
	callee := i.eval(c.Callee)
	arguments := make([]any, 0)
	for _, argument := range c.Arguments {
		arguments = append(arguments, i.eval(argument))
	}
	function, ok := callee.(LoxCallable)
	// make sure the callee is callable
	if !ok {
		panic("can only call functions and classes")
	}
	// arity checking
	if len(arguments) != function.arity() {
		panic("expected " + strconv.Itoa(function.arity()) + " arguments" + " but got " + strconv.Itoa(len(arguments)))
	}
	return function.call(i, arguments)
}

// executing statement
func (i *Interpreter) exec(stmt parser.Stmt) {
	stmt.Accept(i)
}

// The *Interpreter should implement the StmtVisitor interface.
var _ = parser.StmtVisitor(NewInterpreter())

/********** The implementation of the StmtVisitor interface **********/

func (i *Interpreter) VisitorExpressionStmt(e *parser.Expression) any {
	// expression statement like `a;` `1+2;` `true;` ...
	// We evaluate the expression but discard its value.
	i.eval(e.ExprStmtExpr)
	return nil
}

func (i *Interpreter) VisitorPrintStmt(p *parser.Print) any {
	// print statement like `print a;` `print 1+2;` ...
	// We evaluate the expression and prints its value.
	fmt.Println(i.eval(p.Expression))
	return nil
}

func (i *Interpreter) VisitorVarStmt(v *parser.Var) any {
	// variable declaration statement like `var a;` `var b = 1;` ...
	var value any
	if v.Initializer != nil {
		value = i.eval(v.Initializer)
	}
	// define a new variable with `nil` as default value
	i.env.def(v.Name.Lexeme, value)
	return nil
}

func (i *Interpreter) VisitorBlockStmt(b *parser.Block) any {
	i.execBlock(b.Statements, NewEnvironment(i.env))
	return nil
}

func (i *Interpreter) execBlock(statements []parser.Stmt, env *environment) {
	prevEnv := i.env
	defer func() {
		i.env = prevEnv
	}()
	i.env = env
	for _, statement := range statements {
		i.exec(statement)
	}
}

func (i *Interpreter) VisitorIfStmt(ifstmt *parser.If) any {
	if isTruthy(i.eval(ifstmt.Condition)) {
		i.exec(ifstmt.ThenBranch)
	} else if ifstmt.ElseBranch != nil {
		i.exec(ifstmt.ElseBranch)
	}
	return nil
}

func (i *Interpreter) VisitorWhileStmt(w *parser.While) any {
	for isTruthy(i.eval(w.Condition)) {
		i.exec(w.LoopBody)
	}
	return nil
}

func (i *Interpreter) VisitorFunctionStmt(f *parser.Function) any {
	function := NewLoxFunction(f, i.env)
	i.env.def(f.Name.Lexeme, function)
	return nil
}

type Return struct {
	returnValue any
}

func (i *Interpreter) VisitorReturnStmt(r *parser.Return) any {
	var value any
	if r.Value != nil {
		value = i.eval(r.Value)
	}
	panic(Return{returnValue: value})
}
