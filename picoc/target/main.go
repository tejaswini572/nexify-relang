// pico is a small, self-contained interpreter for the C subset used by this project.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type Tok struct {
	s    string
	line int
}

func lex(s string) []Tok {
	var r []Tok
	line := 1
	for i := 0; i < len(s); {
		c := s[i]
		if c == '\n' {
			line++
			i++
			continue
		}
		if unicode.IsSpace(rune(c)) {
			i++
			continue
		}
		if c == '/' && i+1 < len(s) && s[i+1] == '/' {
			for i < len(s) && s[i] != '\n' {
				i++
			}
			continue
		}
		if c == '/' && i+1 < len(s) && s[i+1] == '*' {
			i += 2
			for i+1 < len(s) && !(s[i] == '*' && s[i+1] == '/') {
				if s[i] == '\n' {
					line++
				}
				i++
			}
			i += 2
			continue
		}
		if unicode.IsLetter(rune(c)) || c == '_' {
			j := i + 1
			for j < len(s) && (unicode.IsLetter(rune(s[j])) || unicode.IsDigit(rune(s[j])) || s[j] == '_') {
				j++
			}
			r = append(r, Tok{s[i:j], line})
			i = j
			continue
		}
		if unicode.IsDigit(rune(c)) {
			j := i + 1
			for j < len(s) && (unicode.IsDigit(rune(s[j])) || strings.ContainsRune(".xXabcdefABCDEF", rune(s[j]))) {
				j++
			}
			r = append(r, Tok{s[i:j], line})
			i = j
			continue
		}
		if c == '\'' || c == '"' {
			q := c
			j := i + 1
			for j < len(s) && s[j] != q {
				if s[j] == '\\' && j+1 < len(s) {
					j += 2
				} else {
					j++
				}
			}
			if j < len(s) {
				j++
			}
			r = append(r, Tok{s[i:j], line})
			i = j
			continue
		}
		op := ""
		for _, candidate := range []string{"<<=", ">>=", "==", "!=", "<=", ">=", "++", "--", "+=", "-=", "*=", "/=", "%=", "&&", "||", "<<", ">>", "->", "&=", "|=", "^="} {
			if strings.HasPrefix(s[i:], candidate) {
				op = candidate
				break
			}
		}
		if op == "" {
			op = s[i : i+1]
		}
		r = append(r, Tok{op, line})
		i += len(op)
	}
	return append(r, Tok{"<eof>", line})
}

func preprocess(in string) string {
	macros := map[string]string{}
	defined := map[string]bool{}
	active := []bool{true}
	var out []string
	for _, ln := range strings.Split(in, "\n") {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "#") {
			f := strings.Fields(t)
			if len(f) == 0 {
				continue
			}
			on := active[len(active)-1]
			switch f[0] {
			case "#include":
				continue
			case "#define":
				if on && len(f) >= 2 {
					defined[f[1]] = true
					macros[f[1]] = strings.TrimSpace(strings.TrimPrefix(t, "#define "+f[1]))
				}
			case "#ifdef":
				ok := len(f) > 1 && defined[f[1]]
				active = append(active, on && ok)
			case "#ifndef":
				ok := len(f) > 1 && !defined[f[1]]
				active = append(active, on && ok)
			case "#if":
				ok := false
				if len(f) > 1 {
					v := f[1]
					if m, exists := macros[v]; exists {
						v = strings.TrimSpace(m)
					}
					ok = v != "" && v != "0"
				}
				active = append(active, on && ok)
			case "#else":
				if len(active) > 1 {
					parent := active[len(active)-2]
					active[len(active)-1] = parent && !active[len(active)-1]
				}
			case "#endif":
				if len(active) > 1 {
					active = active[:len(active)-1]
				}
			}
			continue
		}
		if !active[len(active)-1] {
			continue
		}
		for k, v := range macros {
			if v != "" {
				ln = replaceIdent(ln, k, v)
			}
		}
		out = append(out, ln)
	}
	return strings.Join(out, "\n")
}

// replaceIdent deliberately leaves quoted literals alone: C macro expansion does
// not rewrite the text inside a string or character constant.
func replaceIdent(s, name, value string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '"' || s[i] == '\'' {
			q := s[i]
			j := i + 1
			for j < len(s) && s[j] != q {
				if s[j] == '\\' && j+1 < len(s) {
					j += 2
				} else {
					j++
				}
			}
			if j < len(s) {
				j++
			}
			b.WriteString(s[i:j])
			i = j
			continue
		}
		if unicode.IsLetter(rune(s[i])) || s[i] == '_' {
			j := i + 1
			for j < len(s) && (unicode.IsLetter(rune(s[j])) || unicode.IsDigit(rune(s[j])) || s[j] == '_') {
				j++
			}
			w := s[i:j]
			if w == name {
				b.WriteString(value)
			} else {
				b.WriteString(w)
			}
			i = j
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

type Type struct {
	name   string
	ptr    int
	arr    int
	arr2   int
	fields map[string]Type
	order  []string
	union  bool
}

func (t Type) size() int {
	if t.ptr > 0 {
		return 8
	}
	if t.arr > 0 {
		return t.arr * t.elem().size()
	}
	if t.name == "char" || t.name == "unsigned char" {
		return 1
	}
	if t.name == "short" || t.name == "unsigned short" {
		return 2
	}
	if t.name == "double" || t.name == "float" {
		return 8
	}
	if t.fields != nil {
		n := 0
		align := 1
		if t.union {
			for _, name := range t.order {
				f := t.fields[name]
				if f.size() > n {
					n = f.size()
				}
				if a := typeAlign(f); a > align {
					align = a
				}
			}
		} else {
			for _, name := range t.order {
				f := t.fields[name]
				a := typeAlign(f)
				n = (n + a - 1) / a * a
				n += f.size()
				if a > align {
					align = a
				}
			}
		}
		return (n + align - 1) / align * align
	}
	return 4
}
func typeAlign(t Type) int {
	a := t.size()
	if a > 8 {
		return 8
	}
	if a < 1 {
		return 1
	}
	return a
}
func (t Type) elem() Type {
	u := t
	if u.arr2 > 0 {
		u.arr = u.arr2
		u.arr2 = 0
	} else {
		u.arr = 0
	}
	if u.ptr > 0 {
		u.ptr--
	}
	return u
}

type Cell struct{ v Value }
type Ptr struct {
	cells []*Cell
	i     int
	field string
}
type Value struct {
	typ Type
	i   int64
	f   float64
	str string
	p   *Ptr
	a   []*Cell
	obj map[string]*Cell
	mem []byte
}

func zero(t Type) Value {
	v := Value{typ: t}
	if t.arr > 0 {
		v.a = make([]*Cell, t.arr)
		e := t.elem()
		for i := range v.a {
			v.a[i] = &Cell{zero(e)}
		}
	}
	if t.fields != nil {
		if t.union {
			v.mem = make([]byte, t.size())
		} else {
			v.obj = map[string]*Cell{}
			for n, x := range t.fields {
				v.obj[n] = &Cell{zero(x)}
			}
		}
	}
	return v
}
func unionValue(mem []byte, t Type) Value {
	v := zero(t)
	v.mem = mem
	if t.arr > 0 && (t.name == "char" || t.name == "unsigned char") {
		n := t.arr
		if n > len(mem) {
			n = len(mem)
		}
		v.str = string(mem[:n])
		return v
	}
	var u uint64
	for i := 0; i < t.size() && i < len(mem) && i < 8; i++ {
		u |= uint64(mem[i]) << (8 * i)
	}
	v = intval(int64(u), t)
	v.mem = mem
	return v
}
func (c *Cell) store(v Value) {
	if c.v.mem == nil {
		c.v = v
		return
	}
	for i := 0; i < c.v.typ.size() && i < len(c.v.mem) && i < 8; i++ {
		c.v.mem[i] = byte(uint64(v.i) >> (8 * i))
	}
	c.v = unionValue(c.v.mem, c.v.typ)
}
func (v Value) num() float64 {
	if v.typ.name == "double" || v.typ.name == "float" {
		return v.f
	}
	return float64(v.i)
}
func (v Value) truth() bool {
	if v.p != nil {
		return true
	}
	if v.typ.name == "double" || v.typ.name == "float" {
		return v.f != 0
	}
	return v.i != 0
}
func intval(x int64, t Type) Value {
	v := zero(t)
	if t.name == "unsigned int" {
		x = int64(uint32(x))
	}
	if t.name == "unsigned short" {
		x = int64(uint16(x))
	}
	if t.name == "unsigned char" {
		x = int64(uint8(x))
	}
	v.i = x
	if t.name == "char" {
		v.i = int64(int8(x))
	}
	return v
}
func coerce(v Value, t Type) Value {
	if t.ptr > 0 || t.arr > 0 || t.fields != nil {
		v.typ = t
		return v
	}
	return numv(v.num(), t)
}
func floatval(x float64, t Type) Value { v := zero(t); v.f = x; return v }

type Expr interface{ eval(*Env) *Cell }
type Lit struct{ v Value }

func (x Lit) eval(e *Env) *Cell { return &Cell{x.v} }

type Name struct{ s string }

func (x Name) eval(e *Env) *Cell { return e.get(x.s) }

type Unary struct {
	op string
	x  Expr
}

func (x Unary) eval(e *Env) *Cell {
	c := x.x.eval(e)
	v := c.v
	switch x.op {
	case "&":
		return &Cell{Value{typ: Type{name: v.typ.name, ptr: v.typ.ptr + 1}, p: &Ptr{cells: []*Cell{c}}}}
	case "*":
		if v.p != nil && v.p.i < len(v.p.cells) {
			return v.p.cells[v.p.i]
		}
	case "+":
		return &Cell{v}
	case "-":
		return &Cell{numv(-v.num(), v.typ)}
	case "!":
		return &Cell{intval(booli(!v.truth()), Type{name: "int"})}
	case "~":
		return &Cell{intval(^v.i, v.typ)}
	case "++", "--":
		d := int64(1)
		if x.op == "--" {
			d = -1
		}
		c.store(add(c.v, d))
		return c
	}
	return &Cell{zero(Type{name: "int"})}
}

type Binary struct {
	op   string
	l, r Expr
}

func (x Binary) eval(e *Env) *Cell {
	a := x.l.eval(e)
	if x.op == "&&" && !a.v.truth() {
		return &Cell{intval(0, Type{name: "int"})}
	}
	if x.op == "||" && a.v.truth() {
		return &Cell{intval(1, Type{name: "int"})}
	}
	b := x.r.eval(e)
	if x.op == "=" {
		a.store(coerce(b.v, a.v.typ))
		return a
	}
	if strings.HasSuffix(x.op, "=") && x.op != "==" && x.op != "!=" && x.op != "<=" && x.op != ">=" {
		op := strings.TrimSuffix(x.op, "=")
		a.store(bin(op, a.v, b.v))
		return a
	}
	return &Cell{bin(x.op, a.v, b.v)}
}
func numv(x float64, t Type) Value {
	if t.name == "double" || t.name == "float" {
		return floatval(x, t)
	}
	return intval(int64(x), t)
}
func add(v Value, d int64) Value {
	if v.p != nil {
		q := *v.p
		q.i += int(d)
		v.p = &q
		return v
	}
	return numv(v.num()+float64(d), v.typ)
}
func booli(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
func bin(op string, a, b Value) Value {
	t := a.typ
	if a.typ.name == "double" || b.typ.name == "double" || a.typ.name == "float" || b.typ.name == "float" {
		t = Type{name: "double"}
	} else if a.typ.name == "char" || a.typ.name == "unsigned char" || a.typ.name == "short" || a.typ.name == "unsigned short" || b.typ.name == "char" || b.typ.name == "unsigned char" || b.typ.name == "short" || b.typ.name == "unsigned short" {
		// C promotes integer types narrower than int before arithmetic.
		t = Type{name: "int"}
	}
	if a.p != nil && (op == "+" || op == "-") {
		d := b.i
		if op == "-" {
			d = -d
		}
		return add(a, d)
	}
	if op == "==" || op == "!=" {
		var equal bool
		if a.p != nil || b.p != nil {
			equal = a.p == b.p
		} else {
			equal = a.num() == b.num()
		}
		if op == "!=" {
			equal = !equal
		}
		return intval(booli(equal), Type{name: "int"})
	}
	x, y := a.num(), b.num()
	switch op {
	case "+":
		return numv(x+y, t)
	case "-":
		return numv(x-y, t)
	case "*":
		return numv(x*y, t)
	case "/":
		if y == 0 {
			return numv(0, t)
		}
		return numv(x/y, t)
	case "%":
		if b.i == 0 {
			return intval(0, t)
		}
		return intval(a.i%b.i, t)
	case "<<":
		return intval(a.i<<uint(b.i), t)
	case ">>":
		return intval(a.i>>uint(b.i), t)
	case "&":
		return intval(a.i&b.i, t)
	case "|":
		return intval(a.i|b.i, t)
	case "^":
		return intval(a.i^b.i, t)
	case "<":
		return intval(booli(x < y), Type{name: "int"})
	case ">":
		return intval(booli(x > y), Type{name: "int"})
	case "<=":
		return intval(booli(x <= y), Type{name: "int"})
	case ">=":
		return intval(booli(x >= y), Type{name: "int"})
	case "&&":
		return intval(booli(a.truth() && b.truth()), Type{name: "int"})
	case "||":
		return intval(booli(a.truth() || b.truth()), Type{name: "int"})
	}
	return zero(t)
}

type Post struct {
	op string
	x  Expr
}

func (x Post) eval(e *Env) *Cell {
	c := x.x.eval(e)
	old := c.v
	if x.op == "++" {
		c.store(add(c.v, 1))
	} else {
		c.store(add(c.v, -1))
	}
	return &Cell{old}
}

type Index struct{ a, i Expr }

func (x Index) eval(e *Env) *Cell {
	v := x.a.eval(e).v
	n := int(x.i.eval(e).v.i)
	if v.a != nil && n >= 0 && n < len(v.a) {
		return v.a[n]
	}
	if v.p != nil && v.p.i+n >= 0 && v.p.i+n < len(v.p.cells) {
		return v.p.cells[v.p.i+n]
	}
	return &Cell{zero(Type{name: "int"})}
}

type Field struct {
	x     Expr
	n     string
	arrow bool
}

func (x Field) eval(e *Env) *Cell {
	v := x.x.eval(e).v
	if x.arrow && v.p != nil && len(v.p.cells) > v.p.i {
		v = v.p.cells[v.p.i].v
	}
	if v.obj != nil && v.obj[x.n] != nil {
		return v.obj[x.n]
	}
	if v.typ.union && v.mem != nil {
		if t, ok := v.typ.fields[x.n]; ok {
			return &Cell{unionValue(v.mem, t)}
		}
	}
	return &Cell{zero(Type{name: "int"})}
}

type Cast struct {
	t Type
	x Expr
}

func (x Cast) eval(e *Env) *Cell {
	v := x.x.eval(e).v
	if x.t.ptr > 0 {
		return &Cell{Value{typ: x.t, p: v.p}}
	}
	return &Cell{numv(v.num(), x.t)}
}

type Call struct {
	n    string
	args []Expr
}

func (x Call) eval(e *Env) *Cell {
	av := []Value{}
	for _, a := range x.args {
		av = append(av, a.eval(e).v)
	}
	if f := e.funcs[x.n]; f != nil {
		return f.call(e, av)
	}
	return builtin(x.n, av, e)
}

type Stmt interface{ run(*Env) Signal }
type Signal struct {
	k     string
	v     Value
	label string
}
type Block struct{ xs []Stmt }

func (b Block) run(e *Env) Signal {
	labels := map[string]int{}
	for i, x := range b.xs {
		if l, ok := x.(Label); ok {
			labels[l.n] = i
		}
	}
	for i := 0; i < len(b.xs); i++ {
		s := b.xs[i].run(e)
		if s.k == "goto" {
			if at, ok := labels[s.label]; ok {
				i = at - 1
				continue
			}
		}
		if s.k != "" {
			return s
		}
	}
	return Signal{}
}

type Decl struct {
	t    Type
	n    string
	init Expr
	stat bool
}

func (x Decl) run(e *Env) Signal {
	if x.stat && e.static[x.n] != nil {
		e.vars[x.n] = e.static[x.n]
		return Signal{}
	}
	v := zero(x.t)
	if x.init != nil {
		init := x.init.eval(e).v
		if x.t.arr == 0 && init.a != nil {
			v = init
		} else {
			v = coerce(init, x.t)
		}
	}
	c := &Cell{v}
	e.vars[x.n] = c
	if x.stat {
		e.static[x.n] = c
	}
	return Signal{}
}

type ExprStmt struct{ x Expr }

func (x ExprStmt) run(e *Env) Signal {
	if x.x != nil {
		x.x.eval(e)
	}
	return Signal{}
}

type If struct {
	c    Expr
	a, b Stmt
}

func (x If) run(e *Env) Signal {
	if x.c.eval(e).v.truth() {
		return x.a.run(e)
	}
	if x.b != nil {
		return x.b.run(e)
	}
	return Signal{}
}

type While struct {
	c    Expr
	b    Stmt
	post Stmt
}

func (x While) run(e *Env) Signal {
	for x.c == nil || x.c.eval(e).v.truth() {
		s := x.b.run(e)
		if s.k == "return" || s.k == "goto" {
			return s
		}
		if s.k == "break" {
			break
		}
		if x.post != nil {
			x.post.run(e)
		}
	}
	return Signal{}
}

type Return struct{ x Expr }

func (x Return) run(e *Env) Signal {
	v := zero(Type{name: "int"})
	if x.x != nil {
		v = x.x.eval(e).v
	}
	return Signal{k: "return", v: v}
}

type Flow struct{ k, label string }

func (x Flow) run(e *Env) Signal { return Signal{k: x.k, label: x.label} }

type Label struct {
	n string
	s Stmt
}

func (x Label) run(e *Env) Signal { return x.s.run(e) }

type Case struct {
	v    Expr
	body []Stmt
	def  bool
}
type Switch struct {
	x     Expr
	cases []Case
}

func (x Switch) run(e *Env) Signal {
	want, start, fallback := x.x.eval(e).v.i, -1, -1
	for i, c := range x.cases {
		if c.def {
			fallback = i
		} else if c.v.eval(e).v.i == want && start < 0 {
			start = i
		}
	}
	if start < 0 {
		start = fallback
	}
	if start < 0 {
		return Signal{}
	}
	for i := start; i < len(x.cases); i++ {
		for _, st := range x.cases[i].body {
			s := st.run(e)
			if s.k == "break" {
				return Signal{}
			}
			if s.k != "" {
				return s
			}
		}
	}
	return Signal{}
}

type Function struct {
	n       string
	params  []string
	body    Block
	ret     Type
	statics map[string]*Cell
}

func (f *Function) call(parent *Env, args []Value) *Cell {
	e := newEnv(parent)
	if f.statics == nil {
		f.statics = map[string]*Cell{}
	}
	e.static = f.statics
	for i, n := range f.params {
		v := zero(Type{name: "int"})
		if i < len(args) {
			v = args[i]
		}
		e.vars[n] = &Cell{v}
	}
	s := f.body.run(e)
	return &Cell{s.v}
}

type Env struct {
	vars   map[string]*Cell
	funcs  map[string]*Function
	parent *Env
	static map[string]*Cell
}

func newEnv(p *Env) *Env {
	e := &Env{vars: map[string]*Cell{}, funcs: map[string]*Function{}, parent: p, static: map[string]*Cell{}}
	if p != nil {
		e.funcs = p.funcs
	}
	return e
}
func (e *Env) get(n string) *Cell {
	for q := e; q != nil; q = q.parent {
		if c := q.vars[n]; c != nil {
			return c
		}
	}
	return &Cell{zero(Type{name: "int"})}
}

type Parser struct {
	ts    []Tok
	i     int
	types map[string]Type
	funcs map[string]*Function
}

func (p *Parser) peek() string { return p.ts[p.i].s }
func (p *Parser) next() string { x := p.peek(); p.i++; return x }
func (p *Parser) eat(x string) bool {
	if p.peek() == x {
		p.i++
		return true
	}
	return false
}
func (p *Parser) isType(s string) bool {
	_, ok := p.types[s]
	return ok || s == "unsigned" || s == "signed" || s == "long" || s == "short"
}
func (p *Parser) typex() Type {
	parts := []string{}
	for p.peek() == "unsigned" || p.peek() == "signed" || p.peek() == "long" || p.peek() == "short" || p.peek() == "const" {
		parts = append(parts, p.next())
	}
	if p.peek() == "struct" || p.peek() == "union" {
		u := p.next()
		n := p.next()
		t := p.types[u+" "+n]
		if p.eat("{") {
			fields := map[string]Type{}
			order := []string{}
			for !p.eat("}") && p.peek() != "<eof>" {
				ft := p.typex()
				fn := p.next()
				for p.eat("*") {
					ft.ptr++
				}
				if p.eat("[") {
					ft.arr = atoi(p.next())
					p.eat("]")
				}
				p.eat(";")
				fields[fn] = ft
				order = append(order, fn)
			}
			t = Type{name: u + " " + n, fields: fields, order: order, union: u == "union"}
			p.types[t.name] = t
		}
		parts = append(parts, t.name)
	} else if p.peek() == "void" || p.peek() == "char" || p.peek() == "int" || p.peek() == "float" || p.peek() == "double" {
		parts = append(parts, p.next())
	} else if len(parts) == 0 {
		// A typedef name is itself a complete type.
		parts = append(parts, p.next())
	}
	name := strings.Join(parts, " ")
	if t, ok := p.types[name]; ok {
		return t
	}
	return Type{name: name}
}
func (p *Parser) program() *Env {
	e := newEnv(nil)
	e.funcs = p.funcs
	for p.peek() != "<eof>" {
		p.external(e)
	}
	return e
}
func (p *Parser) external(e *Env) {
	stat := p.eat("static")
	if p.peek() == "typedef" {
		p.next()
		t := p.typex()
		n := p.next()
		p.eat(";")
		p.types[n] = t
		return
	}
	if !p.isType(p.peek()) && p.peek() != "struct" && p.peek() != "union" {
		p.next()
		return
	}
	t := p.typex()
	for p.eat("*") {
		t.ptr++
	}
	n := p.next()
	if p.eat("(") {
		ps := []string{}
		if !p.eat(")") {
			for {
				pt := p.typex()
				for p.eat("*") {
					pt.ptr++
				}
				pn := p.next()
				_ = pt
				ps = append(ps, pn)
				if p.eat(")") {
					break
				}
				p.eat(",")
			}
		}
		if p.eat(";") {
			return
		}
		b := p.block()
		p.funcs[n] = &Function{n: n, params: ps, body: b, ret: t}
		return
	}
	d := Decl{t: t, n: n, stat: stat}
	if p.eat("[") {
		d.t.arr = atoi(p.next())
		p.eat("]")
	}
	if p.eat("=") {
		d.init = p.expr(0)
	}
	p.eat(";")
	d.run(e)
}
func (p *Parser) block() Block {
	p.eat("{")
	b := Block{}
	for p.peek() != "}" && p.peek() != "<eof>" {
		b.xs = append(b.xs, p.stmt())
	}
	p.eat("}")
	return b
}
func (p *Parser) stmt() Stmt {
	s := p.peek()
	if s == "{" {
		return p.block()
	}
	if s == ";" {
		p.next()
		return ExprStmt{}
	}
	if s == "if" {
		p.next()
		p.eat("(")
		c := p.expr(0)
		p.eat(")")
		a := p.stmt()
		var b Stmt
		if p.eat("else") {
			b = p.stmt()
		}
		return If{c, a, b}
	}
	if s == "while" {
		p.next()
		p.eat("(")
		c := p.expr(0)
		p.eat(")")
		return While{c: c, b: p.stmt()}
	}
	if s == "do" {
		p.next()
		b := p.stmt()
		p.eat("while")
		p.eat("(")
		c := p.expr(0)
		p.eat(");")
		return While{c: c, b: b}
	}
	if s == "switch" {
		return p.switchStmt()
	}
	if s == "for" {
		p.next()
		p.eat("(")
		var init Stmt
		if p.isType(p.peek()) || p.peek() == "static" {
			init = p.decl()
		} else {
			init = ExprStmt{p.expr(0)}
			p.eat(";")
		}
		c := p.expr(0)
		p.eat(";")
		post := ExprStmt{p.expr(0)}
		p.eat(")")
		b := p.stmt()
		return Block{[]Stmt{init, While{c: c, b: b, post: post}}}
	}
	if s == "return" {
		p.next()
		x := p.expr(0)
		p.eat(";")
		return Return{x}
	}
	if s == "break" || s == "continue" {
		p.next()
		p.eat(";")
		return Flow{k: s}
	}
	if s == "goto" {
		p.next()
		n := p.next()
		p.eat(";")
		return Flow{k: "goto", label: n}
	}
	if p.i+1 < len(p.ts) && p.ts[p.i+1].s == ":" {
		n := p.next()
		p.next()
		return Label{n, p.stmt()}
	}
	if s == "static" || p.isType(s) || s == "struct" || s == "union" {
		return p.decl()
	}
	x := p.expr(0)
	p.eat(";")
	return ExprStmt{x}
}

func (p *Parser) switchStmt() Stmt {
	p.next()
	p.eat("(")
	x := p.expr(0)
	p.eat(")")
	p.eat("{")
	cases := []Case{}
	var cur *Case
	for p.peek() != "}" && p.peek() != "<eof>" {
		if p.eat("case") {
			v := p.expr(0)
			p.eat(":")
			cases = append(cases, Case{v: v})
			cur = &cases[len(cases)-1]
			continue
		}
		if p.eat("default") {
			p.eat(":")
			cases = append(cases, Case{def: true})
			cur = &cases[len(cases)-1]
			continue
		}
		st := p.stmt()
		if cur == nil {
			cases = append(cases, Case{})
			cur = &cases[len(cases)-1]
		}
		cur.body = append(cur.body, st)
	}
	p.eat("}")
	return Switch{x: x, cases: cases}
}
func (p *Parser) decl() Stmt {
	stat := p.eat("static")
	base := p.typex()
	var ds []Stmt
	for {
		t := base
		for p.eat("*") {
			t.ptr++
		}
		n := p.next()
		for p.eat("[") {
			n := atoi(p.next())
			p.eat("]")
			if t.arr == 0 {
				t.arr = n
			} else {
				t.arr2 = n
			}
		}
		var x Expr
		if p.eat("=") {
			if p.peek() == "{" {
				x = p.arrayInit(t)
			} else {
				x = p.expr(0)
			}
		}
		ds = append(ds, Decl{t, n, x, stat})
		if !p.eat(",") {
			break
		}
	}
	p.eat(";")
	if len(ds) == 1 {
		return ds[0]
	}
	return Block{xs: ds}
}

func (p *Parser) arrayInit(t Type) Expr {
	p.eat("{")
	items := []*Cell{}
	for p.peek() != "}" && p.peek() != "<eof>" {
		i := len(items)
		var item *Cell
		if t.arr2 > 0 && p.peek() == "{" {
			item = p.arrayInit(t.elem()).eval(nil)
		} else {
			item = p.expr(0).eval(nil)
			item.v.typ = t.elem()
		}
		_ = i
		items = append(items, item)
		if !p.eat(",") {
			break
		}
	}
	p.eat("}")
	if t.arr == 0 {
		t.arr = len(items)
	}
	v := zero(t)
	copy(v.a, items)
	return Lit{v}
}

var prec = map[string]int{"=": 1, "+=": 1, "-=": 1, "*=": 1, "/=": 1, "%=": 1, "||": 2, "&&": 3, "|": 4, "^": 5, "&": 6, "==": 7, "!=": 7, "<": 8, ">": 8, "<=": 8, ">=": 8, "<<": 9, ">>": 9, "+": 10, "-": 10, "*": 11, "/": 11, "%": 11}

func (p *Parser) expr(min int) Expr {
	x := p.prefix()
	for {
		op := p.peek()
		q := prec[op]
		if q < min || q == 0 {
			break
		}
		p.next()
		// Assignment is right-associative.  Parsing its RHS at the same
		// precedence keeps a=b=c as a=(b=c), rather than treating the second
		// '=' as the start of a primary expression.
		rMin := q + 1
		if q == 1 {
			rMin = q
		}
		r := p.expr(rMin)
		x = Binary{op, x, r}
	}
	return x
}
func unquote(s string) string {
	if len(s) < 2 {
		return s
	}
	s = s[1 : len(s)-1]
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\\"", "\"")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}
func (p *Parser) prefix() Expr {
	s := p.next()
	if s == "sizeof" {
		p.eat("(")
		t := p.typex()
		p.eat(")")
		return p.post(Lit{intval(int64(t.size()), Type{name: "int"})})
	}
	if s == "(" {
		if p.isType(p.peek()) || p.peek() == "struct" || p.peek() == "union" {
			t := p.typex()
			for p.eat("*") {
				t.ptr++
			}
			p.eat(")")
			return p.post(Cast{t, p.prefix()})
		}
		x := p.expr(0)
		p.eat(")")
		return p.post(x)
	}
	if strings.Contains("+ - ! ~ * & ++ --", s) {
		return p.post(Unary{s, p.prefix()})
	}
	if strings.HasPrefix(s, "\"") {
		return p.post(Lit{Value{typ: Type{name: "char", ptr: 1}, str: unquote(s)}})
	}
	if strings.HasPrefix(s, "'") {
		z := unquote(s)
		return p.post(Lit{intval(int64(z[0]), Type{name: "char"})})
	}
	if len(s) > 0 && unicode.IsDigit(rune(s[0])) {
		if strings.Contains(s, ".") {
			f, _ := strconv.ParseFloat(s, 64)
			return p.post(Lit{floatval(f, Type{name: "double"})})
		}
		n := atoi64(s)
		return p.post(Lit{intval(n, Type{name: "int"})})
	}
	return p.post(Name{s})
}
func (p *Parser) post(x Expr) Expr {
	for {
		if p.eat("(") {
			as := []Expr{}
			if !p.eat(")") {
				for {
					as = append(as, p.expr(0))
					if p.eat(")") {
						break
					}
					p.eat(",")
				}
			}
			if n, ok := x.(Name); ok {
				x = Call{n.s, as}
			}
			continue
		}
		if p.eat("[") {
			i := p.expr(0)
			p.eat("]")
			x = Index{x, i}
			continue
		}
		if p.eat(".") {
			x = Field{x, p.next(), false}
			continue
		}
		if p.eat("->") {
			x = Field{x, p.next(), true}
			continue
		}
		if p.peek() == "++" || p.peek() == "--" {
			x = Post{p.next(), x}
			continue
		}
		break
	}
	return x
}
func atoi(s string) int { n, _ := strconv.Atoi(s); return n }
func atoi64(s string) int64 {
	n, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		u, _ := strconv.ParseUint(s, 0, 64)
		return int64(u)
	}
	return n
}

func builtin(n string, a []Value, e *Env) *Cell {
	switch n {
	case "printf", "sprintf":
		if len(a) == 0 {
			return &Cell{intval(0, Type{name: "int"})}
		}
		out := format(a[0].str, a[1:])
		if n == "printf" {
			fmt.Print(out)
		}
		return &Cell{intval(int64(len(out)), Type{name: "int"})}
	case "malloc", "calloc", "realloc":
		sz := 0
		if len(a) > 0 {
			sz = int(a[0].i)
		}
		if n == "calloc" && len(a) > 1 {
			sz *= int(a[1].i)
		}
		cnt := sz/4 + 2
		cs := make([]*Cell, cnt)
		for i := range cs {
			cs[i] = &Cell{zero(Type{name: "int"})}
		}
		return &Cell{Value{typ: Type{name: "void", ptr: 1}, p: &Ptr{cells: cs}}}
	case "free":
		return &Cell{zero(Type{name: "void"})}
	case "strlen":
		if len(a) > 0 {
			return &Cell{intval(int64(len(a[0].str)), Type{name: "int"})}
		}
	}
	return &Cell{intval(0, Type{name: "int"})}
}
func format(f string, a []Value) string {
	var b strings.Builder
	k := 0
	for i := 0; i < len(f); i++ {
		if f[i] != '%' || i+1 >= len(f) {
			b.WriteByte(f[i])
			continue
		}
		i++
		if f[i] == '%' {
			b.WriteByte('%')
			continue
		}
		zeroPad := false
		width := 0
		prec := -1
		if f[i] == '0' {
			zeroPad = true
			i++
		}
		for i < len(f) && f[i] >= '0' && f[i] <= '9' {
			width = width*10 + int(f[i]-'0')
			i++
		}
		if i < len(f) && f[i] == '.' {
			i++
			prec = 0
			for i < len(f) && f[i] >= '0' && f[i] <= '9' {
				prec = prec*10 + int(f[i]-'0')
				i++
			}
		}
		for i < len(f) && strings.ContainsRune("hlL", rune(f[i])) {
			i++
		}
		if k >= len(a) {
			continue
		}
		v := a[k]
		k++
		s := ""
		switch f[i] {
		case 'd', 'i':
			s = strconv.FormatInt(v.i, 10)
		case 'u':
			s = strconv.FormatUint(uint64(v.i), 10)
		case 'x', 'X':
			s = strconv.FormatUint(uint64(v.i), 16)
			if f[i] == 'X' {
				s = strings.ToUpper(s)
			}
		case 'c':
			s = string(rune(v.i))
		case 's':
			s = v.str
		case 'f', 'g':
			if prec < 0 {
				prec = 6
			}
			s = strconv.FormatFloat(v.num(), 'f', prec, 64)
		case 'p':
			s = "0x" + strconv.FormatInt(int64(uintptr(0)), 16)
		}
		if width > len(s) {
			pad := strings.Repeat(map[bool]string{true: "0", false: " "}[zeroPad], width-len(s))
			if zeroPad && strings.HasPrefix(s, "-") {
				s = "-" + pad + s[1:]
			} else {
				s = pad + s
			}
		}
		b.WriteString(s)
	}
	return b.String()
}
func main() {
	if len(os.Args) < 2 {
		return
	}
	path := os.Args[len(os.Args)-1]
	raw, err := os.ReadFile(path)
	if err != nil {
		return
	}
	p := &Parser{ts: lex(preprocess(string(raw))), types: map[string]Type{}, funcs: map[string]*Function{}}
	for _, n := range []string{"void", "char", "short", "int", "long", "float", "double", "unsigned int", "unsigned char", "unsigned short", "unsigned long"} {
		p.types[n] = Type{name: n}
	}
	e := p.program()
	if f := e.funcs["main"]; f != nil {
		f.call(e, nil)
	}
}
