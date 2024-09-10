package bytecode

type ConstantType int

const (
	ConstString ConstantType = iota
	ConstInteger
	ConstFloat
	ConstBoolean
)

type Constant struct {
	Type  ConstantType
	Value interface{}
}
