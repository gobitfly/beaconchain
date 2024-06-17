type point = number[]
type RGB = number[]
type HSV = number[]

function RGB24toRGB01 (rgb: RGB) : RGB {
  return [
    (rgb[0] / 255) ** 2.2,
    (rgb[1] / 255) ** 2.2,
    (rgb[2] / 255) ** 2.2
  ]
}

function RGB01toHSV (rgb: RGB) : HSV {
  const tilted = [
    0.8164966 * rgb[0] - 0.4082483 * (rgb[1] + rgb[2]),
    0.5773503 * (rgb[0] + rgb[1] + rgb[2]),
    0.7071068 * (rgb[2] - rgb[1])
  ]

  return [
    ,
    ,
    0.3 * rgb[0] + 0.6 * rgb[1] + 0.1 * rgb[2]
  ]
}
