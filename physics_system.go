package engine

type VelocityComponent struct {
	velocity Vector3
}

func PhysicsSystemUpdate(w *World) {

}

// void PhysicsSystem::update(Uint64 deltaTime)
// {
//   for (auto& velocityComponentOpt: componentManager->getComponents<cpn::Velocity>().getComponents())
//   {
//     if (!velocityComponentOpt.has_value())
//       continue;

//     cpn::Velocity& velocityComponent = velocityComponentOpt.value();
//     size_t entityId = &velocityComponentOpt - &componentManager->getComponents<cpn::Velocity>().getComponents()[0];
//     auto& transformComponent = componentManager->getComponent<cpn::Transform>(entityId);

//     auto newPosition = transformComponent.getPosition();
//     newPosition.x += velocityComponent.getX() * deltaTime / 1e9;
//     newPosition.y += velocityComponent.getY() * deltaTime / 1e9;
//     newPosition.z += velocityComponent.getZ() * deltaTime / 1e9;
//     if (newPosition.x > 320.0f || newPosition.x < 0.0f)
//     {
//       velocityComponent.setX(-velocityComponent.getX());
//     }
//     if (newPosition.y > 200.0f || newPosition.y < 0.0f)
//     {
//       velocityComponent.setY(-velocityComponent.getY());
//     }
//     transformComponent.setPosition(newPosition);
//   }
// }
