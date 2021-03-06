package expreduce

import "fmt"
import "strings"
import "sort"
import "hash/fnv"

// Symbols are defined by a string-based name
type Symbol struct {
	Name       string
	cachedHash uint64
}

func (this *Symbol) Eval(es *EvalState) Ex {
	//definition, isdefined := es.defined[this.Name]
	definition, isdefined, _ := es.GetDef(this.Name, this)
	if isdefined {
		// We must call Eval because, at this point, the expression has broken
		// out of the evaluation loop.
		toReturn := definition
		// To handle the case where we set a variable to itself.
		if sym, isSym := definition.(*Symbol); isSym {
			if sym.Name == this.Name {
				return toReturn
			}
		}
		toReturn = toReturn.Eval(es)
		retVal, isReturn := tryReturnValue(toReturn, nil, es)
		if isReturn {
			return retVal
		}
		return toReturn
	}
	return this
}

func (this *Symbol) StringForm(params ToStringParams) string {
	if len(this.Name) == 0 {
		return "<EMPTYSYM>"
	}
	if strings.HasPrefix(this.Name, params.context.Val) {
		return fmt.Sprintf("%v", this.Name[len(params.context.Val):])
	}
	for _, pathPart := range params.contextPath.Parts[1:] {
		path := pathPart.(*String).Val
		if strings.HasPrefix(this.Name, path) {
			return fmt.Sprintf("%v", this.Name[len(path):])
		}
	}
	return fmt.Sprintf("%v", this.Name)
}

func (this *Symbol) String() string {
	context, contextPath := DefaultStringFormArgs()
	return this.StringForm(ToStringParams{form: "InputForm", context: context, contextPath: contextPath})
}

func (this *Symbol) IsEqual(other Ex, cl *CASLogger) string {
	otherConv, ok := other.(*Symbol)
	if !ok {
		return "EQUAL_UNK"
	}
	if this.Name == "System`False" && otherConv.Name == "System`True" {
		return "EQUAL_FALSE"
	}
	if this.Name == "System`True" && otherConv.Name == "System`False" {
		return "EQUAL_FALSE"
	}
	if this.Name != otherConv.Name {
		return "EQUAL_UNK"
	}
	return "EQUAL_TRUE"
}

func (this *Symbol) DeepCopy() Ex {
	thiscopy := *this
	return &thiscopy
}

func (this *Symbol) Copy() Ex {
	return this
}

// Functions for working with the attributes of symbols:
type Attributes struct {
	Orderless       bool
	Flat            bool
	OneIdentity     bool
	Listable        bool
	Constant        bool
	NumericFunction bool
	Protected       bool
	Locked          bool
	ReadProtected   bool
	HoldFirst       bool
	HoldRest        bool
	HoldAll         bool
	HoldAllComplete bool
	NHoldFirst      bool
	NHoldRest       bool
	NHoldAll        bool
	SequenceHold    bool
	Temporary       bool
	Stub            bool
}

func (this *Symbol) Attrs(dm *DefMap) Attributes {
	def, isDef := (*dm)[this.Name]
	if !isDef {
		return Attributes{}
	}
	return def.attributes
}

func (this *Symbol) Default(dm *DefMap) Ex {
	def, isDef := (*dm)[this.Name]
	if !isDef {
		return nil
	}
	return def.defaultExpr
}

func stringsToAttributes(strings []string) Attributes {
	attrs := Attributes{}
	for _, s := range strings {
		if s == "Orderless" {
			attrs.Orderless = true
		}
		if s == "Flat" {
			attrs.Flat = true
		}
		if s == "OneIdentity" {
			attrs.OneIdentity = true
		}
		if s == "Listable" {
			attrs.Listable = true
		}
		if s == "Constant" {
			attrs.Constant = true
		}
		if s == "NumericFunction" {
			attrs.NumericFunction = true
		}
		if s == "Protected" {
			attrs.Protected = true
		}
		if s == "Locked" {
			attrs.Locked = true
		}
		if s == "ReadProtected" {
			attrs.ReadProtected = true
		}
		if s == "HoldFirst" {
			attrs.HoldFirst = true
		}
		if s == "HoldRest" {
			attrs.HoldRest = true
		}
		if s == "HoldAll" {
			attrs.HoldAll = true
		}
		if s == "HoldAllComplete" {
			attrs.HoldAllComplete = true
		}
		if s == "NHoldFirst" {
			attrs.NHoldFirst = true
		}
		if s == "NHoldRest" {
			attrs.NHoldRest = true
		}
		if s == "NHoldAll" {
			attrs.NHoldAll = true
		}
		if s == "SequenceHold" {
			attrs.SequenceHold = true
		}
		if s == "Temporary" {
			attrs.Temporary = true
		}
		if s == "Stub" {
			attrs.Stub = true
		}
	}
	return attrs
}

func (this *Attributes) toStrings() []string {
	var strings []string
	if this.Orderless {
		strings = append(strings, "Orderless")
	}
	if this.Flat {
		strings = append(strings, "Flat")
	}
	if this.OneIdentity {
		strings = append(strings, "OneIdentity")
	}
	if this.Listable {
		strings = append(strings, "Listable")
	}
	if this.Constant {
		strings = append(strings, "Constant")
	}
	if this.NumericFunction {
		strings = append(strings, "NumericFunction")
	}
	if this.Protected {
		strings = append(strings, "Protected")
	}
	if this.Locked {
		strings = append(strings, "Locked")
	}
	if this.ReadProtected {
		strings = append(strings, "ReadProtected")
	}
	if this.HoldFirst {
		strings = append(strings, "HoldFirst")
	}
	if this.HoldRest {
		strings = append(strings, "HoldRest")
	}
	if this.HoldAll {
		strings = append(strings, "HoldAll")
	}
	if this.HoldAllComplete {
		strings = append(strings, "HoldAllComplete")
	}
	if this.NHoldFirst {
		strings = append(strings, "NHoldFirst")
	}
	if this.NHoldRest {
		strings = append(strings, "NHoldRest")
	}
	if this.NHoldAll {
		strings = append(strings, "NHoldAll")
	}
	if this.SequenceHold {
		strings = append(strings, "SequenceHold")
	}
	if this.Temporary {
		strings = append(strings, "Temporary")
	}
	if this.Stub {
		strings = append(strings, "Stub")
	}

	sort.Strings(strings)
	return strings
}

func (this *Attributes) toSymList() *Expression {
	toReturn := E(S("List"))
	for _, s := range this.toStrings() {
		toReturn.appendEx(S(s))
	}
	return toReturn
}

func (this *Symbol) NeedsEval() bool {
	return false
}

func (this *Symbol) Hash() uint64 {
	if this.cachedHash > 0 {
		return this.cachedHash
	}
	h := fnv.New64a()
	h.Write([]byte{107, 10, 247, 23, 33, 221, 163, 156})
	h.Write([]byte(this.Name))
	this.cachedHash = h.Sum64()
	return h.Sum64()
}

func ContainsSymbol(e Ex, name string) bool {
	asSym, isSym := e.(*Symbol)
	if isSym {
		return asSym.Name == name
	}
	asExpr, isExpr := e.(*Expression)
	if isExpr {
		for _, part := range asExpr.Parts {
			if ContainsSymbol(part, name) {
				return true
			}
		}
	}
	return false
}

func NewSymbol(name string) *Symbol {
	return &Symbol{Name: name}
}

func S(name string) Ex {
	return NewSymbol("System`" + name)
}
