package engine

type Rect struct {
	x uint
	y uint
	w uint
	h uint
}

type Vector3 struct {
	x, y, z float32
}

// typedef struct Matrix4x4
// {
// 	float m11, m12, m13, m14;
// 	float m21, m22, m23, m24;
// 	float m31, m32, m33, m34;
// 	float m41, m42, m43, m44;
// } Matrix4x4;
// Matrix4x4 Matrix4x4_CreateOrthographicOffCenter(
//   float left,
//   float right,
//   float bottom,
//   float top,
//   float zNear,
//   float zFar
// );
// Matrix4x4 Matrix4x4_CreateOrthographicOffCenter(
//   float left,
//   float right,
//   float bottom,
//   float top,
//   float zNear,
//   float zFar
// )
// {
//   Matrix4x4 result;
//   result.m11 = 2.0f / (right - left);
//   result.m12 = 0.0f;
//   result.m13 = 0.0f;
//   result.m14 = 0.0f;
//   result.m21 = 0.0f;
//   result.m22 = 2.0f / (top - bottom);
//   result.m23 = 0.0f;
//   result.m24 = 0.0f;
//   result.m31 = 0.0f;
//   result.m32 = 0.0f;
//   result.m33 = 1.0f / (zNear - zFar);
//   result.m34 = 0.0f;
//   result.m41 = (right + left) / (left - right);
//   result.m42 =  (top + bottom) / (bottom - top);
//   result.m43 = zNear / (zNear - zFar);
//   result.m44 = 1.0f;
//   return result;
// }
