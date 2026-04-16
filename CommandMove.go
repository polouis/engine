package engine

type MoveCommand struct {
	move Vector3
}

func NewMoveCommand(x, y, z float32) *MoveCommand {
	return &MoveCommand{move: Vector3{X: x, Y: y, Z: z}}
}

func (mc MoveCommand) Execute(ctx *Context, e EntityID) {
	t, _ := ctx.W.TransformStore.Get(e)
	t.Position.X += mc.move.X
	t.Position.Y += mc.move.Y
	t.Position.Z += mc.move.Z
}
