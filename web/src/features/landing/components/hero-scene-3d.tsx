/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useEffect, useRef } from 'react'
import * as THREE from 'three'

import { cn } from '@/lib/utils'

interface HeroScene3DProps {
  className?: string
  isDark?: boolean
}

export default function HeroScene3D(props: HeroScene3DProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const isDark = props.isDark ?? true

  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    let renderer: THREE.WebGLRenderer | null = null
    try {
      renderer = new THREE.WebGLRenderer({
        antialias: true,
        alpha: true,
        powerPreference: 'high-performance',
      })
    } catch {
      return
    }

    const width = Math.max(container.clientWidth, 1)
    const height = Math.max(container.clientHeight, 1)
    renderer.setSize(width, height)
    renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2))
    container.innerHTML = ''
    container.appendChild(renderer.domElement)

    const scene = new THREE.Scene()
    const fogColor = isDark ? 0x0b1016 : 0xf1eee6
    scene.fog = new THREE.FogExp2(fogColor, 0.032)

    const camera = new THREE.PerspectiveCamera(58, width / height, 0.1, 100)
    camera.position.set(0, 1.8, 9.5)

    const primaryHex = isDark ? 0x2bbfb4 : 0x087c74
    const secondaryHex = isDark ? 0x16857d : 0x0d9488
    const accentHex = isDark ? 0x67e8f9 : 0x0f766e

    const rootGroup = new THREE.Group()
    scene.add(rootGroup)

    // 1. Undulating wave grid at the bottom
    const gridGeometry = new THREE.PlaneGeometry(42, 32, 44, 32)
    gridGeometry.rotateX(-Math.PI * 0.48)
    const gridBasePositions = Float32Array.from(
      gridGeometry.attributes.position.array as ArrayLike<number>
    )
    const gridMaterial = new THREE.MeshBasicMaterial({
      color: primaryHex,
      wireframe: true,
      transparent: true,
      opacity: isDark ? 0.22 : 0.16,
    })
    const gridMesh = new THREE.Mesh(gridGeometry, gridMaterial)
    gridMesh.position.set(0, -3.2, -4)
    rootGroup.add(gridMesh)

    // 2. Forward-flying glowing rings
    const ringCount = 6
    const ringTrackLength = 26
    const rings: THREE.Mesh[] = []
    const ringGeometries: THREE.BufferGeometry[] = []
    const ringMaterials: THREE.Material[] = []

    for (let i = 0; i < ringCount; i++) {
      const radius = 2.2 + (i % 3) * 0.65
      const ringGeo = new THREE.TorusGeometry(radius, 0.018, 12, 72)
      const ringMat = new THREE.MeshBasicMaterial({
        color: i % 2 === 0 ? primaryHex : accentHex,
        transparent: true,
        opacity: 0.35,
      })
      const ring = new THREE.Mesh(ringGeo, ringMat)
      ring.position.set(0, 0.4, -i * (ringTrackLength / ringCount))
      ring.rotation.z = (i * Math.PI) / 6
      rings.push(ring)
      ringGeometries.push(ringGeo)
      ringMaterials.push(ringMat)
      rootGroup.add(ring)
    }

    // 3. Central wireframe polyhedron + inner core
    const polyGroup = new THREE.Group()
    polyGroup.position.set(2.8, 0.7, -1.2)
    rootGroup.add(polyGroup)

    const outerPolyGeo = new THREE.IcosahedronGeometry(1.75, 1)
    const outerPolyMat = new THREE.MeshBasicMaterial({
      color: primaryHex,
      wireframe: true,
      transparent: true,
      opacity: isDark ? 0.34 : 0.26,
    })
    const outerPolyMesh = new THREE.Mesh(outerPolyGeo, outerPolyMat)
    polyGroup.add(outerPolyMesh)

    const innerPolyGeo = new THREE.OctahedronGeometry(1.05, 0)
    const innerPolyMat = new THREE.MeshBasicMaterial({
      color: accentHex,
      wireframe: true,
      transparent: true,
      opacity: isDark ? 0.5 : 0.38,
    })
    const innerPolyMesh = new THREE.Mesh(innerPolyGeo, innerPolyMat)
    polyGroup.add(innerPolyMesh)

    // 4. Ambient particles
    const particleCount = 110
    const particlePositions = new Float32Array(particleCount * 3)
    for (let i = 0; i < particleCount; i++) {
      const i3 = i * 3
      // Deterministic pseudo-random distribution so tests/renders stay predictable
      const seed1 = Math.sin(i * 12.9898) * 43758.5453
      const seed2 = Math.sin(i * 78.233) * 43758.5453
      const seed3 = Math.sin(i * 39.346) * 43758.5453
      const rx = seed1 - Math.floor(seed1)
      const ry = seed2 - Math.floor(seed2)
      const rz = seed3 - Math.floor(seed3)
      particlePositions[i3] = (rx - 0.5) * 26
      particlePositions[i3 + 1] = (ry - 0.35) * 12
      particlePositions[i3 + 2] = -rz * 22 + 2
    }
    const particleGeo = new THREE.BufferGeometry()
    particleGeo.setAttribute(
      'position',
      new THREE.BufferAttribute(particlePositions, 3)
    )
    const particleMat = new THREE.PointsMaterial({
      color: secondaryHex,
      size: 0.065,
      transparent: true,
      opacity: isDark ? 0.65 : 0.45,
    })
    const particles = new THREE.Points(particleGeo, particleMat)
    rootGroup.add(particles)

    // Mouse parallax tracking
    const targetPointer = { x: 0, y: 0 }
    const currentPointer = { x: 0, y: 0 }

    const handlePointerMove = (event: PointerEvent) => {
      const nx = (event.clientX / Math.max(window.innerWidth, 1)) * 2 - 1
      const ny = -(event.clientY / Math.max(window.innerHeight, 1)) * 2 + 1
      targetPointer.x = nx
      targetPointer.y = ny
    }

    window.addEventListener('pointermove', handlePointerMove, { passive: true })

    const handleResize = () => {
      if (!container || !renderer) return
      const nextW = Math.max(container.clientWidth, 1)
      const nextH = Math.max(container.clientHeight, 1)
      camera.aspect = nextW / nextH
      camera.updateProjectionMatrix()
      renderer.setSize(nextW, nextH)
    }

    window.addEventListener('resize', handleResize, { passive: true })

    let frameId = 0
    const startTime = performance.now()

    const tick = (now: number) => {
      frameId = window.requestAnimationFrame(tick)
      const elapsed = (now - startTime) * 0.001

      // Smooth pointer damping
      currentPointer.x += (targetPointer.x - currentPointer.x) * 0.045
      currentPointer.y += (targetPointer.y - currentPointer.y) * 0.045

      rootGroup.rotation.y = currentPointer.x * 0.14
      rootGroup.rotation.x = -currentPointer.y * 0.08

      // Wave grid vertex animation
      const posAttr = gridGeometry.attributes.position
      const posArray = posAttr.array as Float32Array
      for (let i = 0; i < posAttr.count; i++) {
        const i3 = i * 3
        const bx = gridBasePositions[i3]
        const bz = gridBasePositions[i3 + 2]
        const wave =
          Math.sin(bx * 0.32 + elapsed * 1.4) * 0.36 +
          Math.cos(bz * 0.28 + elapsed * 1.1) * 0.28
        posArray[i3 + 1] = gridBasePositions[i3 + 1] + wave
      }
      posAttr.needsUpdate = true

      // Flying rings toward camera
      for (let i = 0; i < rings.length; i++) {
        const ring = rings[i]
        const mat = ringMaterials[i] as THREE.MeshBasicMaterial
        ring.position.z += 0.032
        ring.rotation.z += 0.0025 * (i % 2 === 0 ? 1 : -1)
        if (ring.position.z > 6.5) {
          ring.position.z = -ringTrackLength + 6.5
        }
        // Fade in from deep z and fade out near camera
        const normalizedZ =
          (ring.position.z + ringTrackLength - 6.5) / ringTrackLength
        const edgeFade = Math.sin(
          Math.min(Math.max(normalizedZ, 0), 1) * Math.PI
        )
        mat.opacity = edgeFade * (isDark ? 0.36 : 0.24)
      }

      // Rotate polyhedron
      outerPolyMesh.rotation.y = elapsed * 0.24
      outerPolyMesh.rotation.x = elapsed * 0.15
      innerPolyMesh.rotation.y = -elapsed * 0.38
      innerPolyMesh.rotation.z = elapsed * 0.22

      // Gentle particle drift
      particles.rotation.y = elapsed * 0.03

      renderer?.render(scene, camera)
    }

    frameId = window.requestAnimationFrame(tick)

    return () => {
      window.cancelAnimationFrame(frameId)
      window.removeEventListener('pointermove', handlePointerMove)
      window.removeEventListener('resize', handleResize)
      gridGeometry.dispose()
      gridMaterial.dispose()
      for (const geo of ringGeometries) geo.dispose()
      for (const mat of ringMaterials) mat.dispose()
      outerPolyGeo.dispose()
      outerPolyMat.dispose()
      innerPolyGeo.dispose()
      innerPolyMat.dispose()
      particleGeo.dispose()
      particleMat.dispose()
      renderer?.dispose()
      if (
        renderer?.domElement &&
        renderer.domElement.parentNode === container
      ) {
        container.removeChild(renderer.domElement)
      }
    }
  }, [isDark])

  return (
    <div
      ref={containerRef}
      data-testid='landing-hero-3d-canvas'
      aria-hidden='true'
      className={cn(
        'pointer-events-none absolute inset-0 overflow-hidden',
        props.className
      )}
    />
  )
}
