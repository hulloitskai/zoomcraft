// A majority of this code is stolen gracelessly from:
// https://www.khanacademy.org/computing/computer-programming/programming-games-visualizations/programming-3d-shapes/a/rotating-3d-shapes
//
// I stole this because I do not have a basic understanding of university level
// maths.

export function deg2rad(degrees) {
  return (degrees * Math.PI) / 180;
}

export function rotate(vec, ay, ax) {
  vec = rotateY(vec, ay);
  vec = rotateX(vec, ax);
  return vec;
}

function rotateX(vec, theta) {
  const sin = Math.sin(theta);
  const cos = Math.cos(theta);
  const [x, y, z] = vec;
  return [x, y * cos - z * sin, z * cos + y * sin];
}

function rotateY(vec, theta) {
  const sin = Math.sin(theta);
  const cos = Math.cos(theta);
  const [x, y, z] = vec;
  return [x * cos + z * sin, y, z * cos - x * sin];
}
