package engine

type VelocityComponent struct {
	velocity Vector3
}

var VelocityCID = RegisterComponent[VelocityComponent]()

func GetVelocityComponents(w *World) *ComponentArray[VelocityComponent] {
	return w.Store(VelocityCID).(*ComponentArray[VelocityComponent])
}

func UpdatePhysicsSystem(ctx *Context, deltaTime uint64) {

	for e, velocityCpnt := range GetVelocityComponents(ctx.W).All() {
		transformCpnt, _ := GetTransformComponents(ctx.W).Get(e)

		velocityCpnt.velocity.X += velocityCpnt.velocity.X * float32(deltaTime) / 1e9
		velocityCpnt.velocity.Y += velocityCpnt.velocity.Y * float32(deltaTime) / 1e9
		velocityCpnt.velocity.Z += velocityCpnt.velocity.Z * float32(deltaTime) / 1e9

		if transformCpnt.Position.X > 320.0 || transformCpnt.Position.X < 0.0 {
			velocityCpnt.velocity.X = -velocityCpnt.velocity.X
		}
		if transformCpnt.Position.Y > 200.0 || transformCpnt.Position.Y < 0.0 {
			velocityCpnt.velocity.Y = -velocityCpnt.velocity.Y
		}
	}
}
