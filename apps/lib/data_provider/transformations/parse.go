package transformations

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/Stork-Oracle/stork-external/apps/lib/data_provider/types"
	"github.com/alecthomas/participle/v2"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

type evaluable interface {
	Eval(ctx map[string]types.DataSourceValueUpdate) float64
	getDependencies() []string
}

type Operator int

const (
	OpMul Operator = iota
	OpDiv
	OpAdd
	OpSub
)

var operatorMap = map[string]Operator{"+": OpAdd, "-": OpSub, "*": OpMul, "/": OpDiv}

func (o *Operator) Capture(s []string) error {
	*o = operatorMap[s[0]]
	return nil
}

// E --> T {( "+" | "-" ) T}
// T --> F {( "*" | "/" ) F}
// F --> P ["^" F]
// P --> v | "(" E ")" | "-" T

type Value struct {
	Function      *Function   `@@`
	Number        *float64    `| @(Float|Int)`
	Variable      *string     `| @(Ident ("." Ident)*)`
	Subexpression *Expression `| "(" @@ ")"`
}

type Function struct {
	Name      string        `@Ident`
	Arguments *Expression   `"(" @@`
	RestArgs  []*Expression `( "," @@ )*")"`
}

type Factor struct {
	Base     *Value `@@`
	Exponent *Value `( "^" @@ )?`
}

type OpFactor struct {
	Operator Operator `@("*" | "/")`
	Factor   *Factor  `@@`
}

type Term struct {
	Left  *Factor     `@@`
	Right []*OpFactor `@@*`
}

type OpTerm struct {
	Operator Operator `@("+" | "-")`
	Term     *Term    `@@`
}

type Expression struct {
	Left  *Term     `@@`
	Right []*OpTerm `@@*`
}

// Display

func (o Operator) String() string {
	switch o {
	case OpMul:
		return "*"
	case OpDiv:
		return "/"
	case OpSub:
		return "-"
	case OpAdd:
		return "+"
	}
	panic("unsupported operator")
}

func (v *Value) String() string {
	if v.Number != nil {
		return fmt.Sprintf("%g", *v.Number)
	}
	if v.Variable != nil {
		return *v.Variable
	}
	if v.Function != nil {
		return v.Function.String()
	}
	return "(" + v.Subexpression.String() + ")"
}

func (f *Factor) String() string {
	out := f.Base.String()
	if f.Exponent != nil {
		out += " ^ " + f.Exponent.String()
	}
	return out
}

func (o *OpFactor) String() string {
	return fmt.Sprintf("%s %s", o.Operator, o.Factor)
}

func (t *Term) String() string {
	out := []string{t.Left.String()}
	for _, r := range t.Right {
		out = append(out, r.String())
	}
	return strings.Join(out, " ")
}

func (o *OpTerm) String() string {
	return fmt.Sprintf("%s %s", o.Operator, o.Term)
}

func (e *Expression) String() string {
	out := []string{e.Left.String()}
	for _, r := range e.Right {
		out = append(out, r.String())
	}
	return strings.Join(out, " ")
}

func (f *Function) String() string {
	args := []string{f.Arguments.String()}
	for _, arg := range f.RestArgs {
		args = append(args, arg.String())
	}
	return fmt.Sprintf("%s(%s)", f.Name, strings.Join(args, ", "))
}

// Evaluation

func (o Operator) Eval(l, r float64) float64 {
	switch o {
	case OpMul:
		return l * r
	case OpDiv:
		return l / r
	case OpAdd:
		return l + r
	case OpSub:
		return l - r
	}
	panic("unsupported operator")
}

func (v *Value) Eval(ctx map[string]types.DataSourceValueUpdate) float64 {
	switch {
	case v.Number != nil:
		return *v.Number
	case v.Variable != nil:
		value, ok := ctx[string(*v.Variable)]
		if !ok {
			return math.NaN()
		}
		return value.Value
	case v.Function != nil:
		return v.Function.Eval(ctx)
	default:
		return v.Subexpression.Eval(ctx)
	}
}

func (f *Factor) Eval(ctx map[string]types.DataSourceValueUpdate) float64 {
	b := f.Base.Eval(ctx)
	if f.Exponent != nil {
		return math.Pow(b, f.Exponent.Eval(ctx))
	}
	return b
}

func (t *Term) Eval(ctx map[string]types.DataSourceValueUpdate) float64 {
	n := t.Left.Eval(ctx)
	for _, r := range t.Right {
		n = r.Operator.Eval(n, r.Factor.Eval(ctx))
	}
	return n
}

func (e *Expression) Eval(ctx map[string]types.DataSourceValueUpdate) float64 {
	l := e.Left.Eval(ctx)
	for _, r := range e.Right {
		l = r.Operator.Eval(l, r.Term.Eval(ctx))
	}
	return l
}

func (f *Function) Eval(ctx map[string]types.DataSourceValueUpdate) float64 {
	// Collect all arguments
	args := []float64{f.Arguments.Eval(ctx)}
	for _, arg := range f.RestArgs {
		args = append(args, arg.Eval(ctx))
	}

	switch f.Name {
	case "median":
		// Sort the values
		sort.Float64s(args)
		// Calculate median
		n := len(args)
		if n%2 == 0 {
			return (args[n/2-1] + args[n/2]) / 2
		}
		return args[n/2]

	case "mean":
		sum := 0.0
		for _, v := range args {
			sum += v
		}
		return sum / float64(len(args))

	case "sum":
		sum := 0.0
		for _, v := range args {
			sum += v
		}
		return sum

	case "product":
		product := 1.0
		for _, v := range args {
			product *= v
		}
		return product
	default:
		panic("unknown function: " + f.Name)
	}
}

func (f *Function) getDependencies() []string {
	deps := f.Arguments.getDependencies()
	for _, arg := range f.RestArgs {
		deps = append(deps, arg.getDependencies()...)
	}
	return deps
}

func (f *Factor) getDependencies() []string {
	deps := f.Base.getDependencies()
	if f.Exponent != nil {
		deps = append(deps, f.Exponent.getDependencies()...)
	}
	return deps
}

func (f *Value) getDependencies() []string {
	switch {
	case f.Variable != nil:
		return []string{*f.Variable}
	case f.Function != nil:
		return f.Function.getDependencies()
	case f.Subexpression != nil:
		return f.Subexpression.getDependencies()
	default:
		return []string{}
	}
}

func (f *Term) getDependencies() []string {
	deps := f.Left.getDependencies()
	for _, r := range f.Right {
		deps = append(deps, r.Factor.getDependencies()...)
	}
	return deps
}

func (f *OpTerm) getDependencies() []string {
	return f.Term.getDependencies()
}

func (f *Expression) getDependencies() []string {
	deps := f.Left.getDependencies()
	for _, r := range f.Right {
		deps = append(deps, r.Term.getDependencies()...)
	}
	return deps
}

var parser = participle.MustBuild[Expression]()

func parse(formula string) (*Expression, error) {
	return parser.ParseString("", formula)
}

type OrderedTransformation struct {
	Id             types.ValueId
	Transformation *Expression
}

func (o *OrderedTransformation) String() string {
	return fmt.Sprintf("%s: %s", o.Id, o.Transformation.String())
}

func BuildTransformations(transformations []types.DataProviderTransformationConfig, sourceIds map[types.ValueId]interface{}) ([]OrderedTransformation, error) {
	g := simple.NewDirectedGraph()

	// allow translating node <-> price id
	nodeToTransformationId := make(map[graph.Node]types.ValueId)
	transformationIdToNode := make(map[types.ValueId]graph.Node)

	parsedTransformations := make(map[types.ValueId]*Expression)
	for _, transformation := range transformations {
		expr, err := parse(transformation.Formula)
		if err != nil {
			return nil, err
		}
		parsedTransformations[transformation.Id] = expr

		node := g.NewNode()
		g.AddNode(node)
		nodeToTransformationId[node] = transformation.Id
		transformationIdToNode[transformation.Id] = node
	}

	for _, transformation := range transformations {
		expr, ok := parsedTransformations[transformation.Id]
		if !ok {
			return nil, fmt.Errorf("no such transformation: %s", transformation.Id)
		}

		deps := expr.getDependencies()
		for _, dep := range deps {
			if strings.HasPrefix(dep, "t.") {
				dep = dep[2:]
				if _, ok := transformationIdToNode[types.ValueId(dep)]; !ok {
					return nil, fmt.Errorf("no such transformation: %s", dep)
				}
				g.SetEdge(g.NewEdge(transformationIdToNode[types.ValueId(dep)], transformationIdToNode[transformation.Id]))
			} else if strings.HasPrefix(dep, "s.") {
				dep = dep[2:]
				if _, ok := sourceIds[types.ValueId(dep)]; !ok {
					return nil, fmt.Errorf("no such source: %s", dep)
				}
			} else {
				return nil, fmt.Errorf("unknown dependency: %s", dep)
			}
		}
	}

	nodes, err := topo.Sort(g)
	if err != nil {
		return nil, fmt.Errorf("could not linearize price id graph - there may be circular dependencies: %v", err)
	}

	transformationsInOrder := make([]OrderedTransformation, len(nodes))
	for i, node := range nodes {
		transformationsInOrder[i] = OrderedTransformation{
			Id:             nodeToTransformationId[node],
			Transformation: parsedTransformations[nodeToTransformationId[node]],
		}
	}

	return transformationsInOrder, nil
}
