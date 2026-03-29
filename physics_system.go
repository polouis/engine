package engine

type VelocityComponent struct {
	velocity Vector3
}

var VelocityCID = RegisterComponent[VelocityComponent]()

func GetVelocityComponents(w *World) *ComponentArray[VelocityComponent] {
	return w.Store(VelocityCID).(*ComponentArray[VelocityComponent])
}

func UpdatePhysicsSystem(w *World, deltaTime uint64) {

	for e, velocityCpnt := range GetVelocityComponents(w).All() {
		transformCpnt, _ := GetTransformComponents(w).Get(e)

		velocityCpnt.velocity.x += velocityCpnt.velocity.x * float32(deltaTime) / 1e9
		velocityCpnt.velocity.y += velocityCpnt.velocity.y * float32(deltaTime) / 1e9
		velocityCpnt.velocity.z += velocityCpnt.velocity.z * float32(deltaTime) / 1e9

		if transformCpnt.position.x > 320.0 || transformCpnt.position.x < 0.0 {
			velocityCpnt.velocity.x = -velocityCpnt.velocity.x
		}
		if transformCpnt.position.y > 200.0 || transformCpnt.position.y < 0.0 {
			velocityCpnt.velocity.y = -velocityCpnt.velocity.y
		}
	}
}
