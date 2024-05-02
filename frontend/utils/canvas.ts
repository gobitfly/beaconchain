function drawArc (ctx: CanvasRenderingContext2D, centerX:number, centerY:number, radius:number, startAngle:number, endAngle:number, color = '#000000') {
  ctx.save()
  ctx.strokeStyle = color
  ctx.beginPath()
  ctx.arc(centerX, centerY, radius, startAngle, endAngle)
  ctx.stroke()
  ctx.restore()
}

function drawPieSlice (
  ctx: CanvasRenderingContext2D,
  centerX: number,
  centerY: number,
  radius: number,
  startAngle: number,
  endAngle: number,
  fillColor: string,
  strokeColor = '#000000'
) {
  ctx.save()
  ctx.fillStyle = fillColor
  ctx.strokeStyle = strokeColor
  ctx.beginPath()
  ctx.moveTo(centerX, centerY)
  ctx.arc(centerX, centerY, radius, startAngle, endAngle)
  ctx.closePath()
  ctx.fill()
  ctx.restore()
}

// drawas a pie chart where all slizes have the same size
export function drawEqualPieChart (
  colors?: string[],
  size: number = 20
): HTMLCanvasElement {
  const canvas = document.createElement('canvas') as HTMLCanvasElement
  canvas.width = size
  canvas.height = size
  if (!colors?.length) {
    return canvas
  }
  const ctx = canvas.getContext('2d')!
  const radius = size / 2

  const totalValue = 360
  const slice = 360 / colors.length
  let startAngle = -Math.PI / 2
  const sliceAngle = (2 * Math.PI * slice) / totalValue
  colors.forEach((color) => {
    drawPieSlice(
      ctx,
      radius,
      radius,
      radius,
      startAngle,
      startAngle + sliceAngle,
      color
    )
    startAngle += sliceAngle
  })

  // outer border
  drawArc(
    ctx,
    radius,
    radius,
    radius,
    0,
    360
  )

  return canvas
}
