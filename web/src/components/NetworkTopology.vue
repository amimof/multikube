<script setup lang="ts">
import { computed } from 'vue'

export interface NormalizedServer {
	url: string
	phase: string
	reason: string
	lastTransitionTime?: Date
}

const props = defineProps<{
	backendName: string
	lbType: string
	servers: NormalizedServer[]
}>()

// Layout constants
const PAD_X = 40
const PAD_Y = 40
const ROOT_MIN_W = 140
const ROOT_H = 60
const CHAR_WIDTH = 8.5 // approximate average char width at 14px
const ROOT_PAD = 32 // horizontal padding inside root node
const TARGET_W = 280
const TARGET_H = 50
const SPACING_Y = 70
const MIN_SVG_W = 700

const rootW = computed(() => {
  const nameW = props.backendName.length * CHAR_WIDTH + ROOT_PAD
  const badgeW = props.lbType.length * 7 + ROOT_PAD // badge font is smaller
  return Math.max(ROOT_MIN_W, Math.ceil(Math.max(nameW, badgeW)))
})

const svgW = computed(() => Math.max(MIN_SVG_W, rootW.value + TARGET_W + PAD_X * 3 + 60))

const rootX = PAD_X
const targetX = computed(() => svgW.value - PAD_X - TARGET_W)

const svgH = computed(() => {
  const count = props.servers.length || 1
  return Math.max(200, count * SPACING_Y + PAD_Y * 2)
})

const rootY = computed(() => svgH.value / 2 - ROOT_H / 2)

function targetY(index: number): number {
	const count = props.servers.length || 1
	const totalH = count * TARGET_H + (count - 1) * (SPACING_Y - TARGET_H)
	const startY = (svgH.value - totalH) / 2
	return startY + index * SPACING_Y
}

function bezierPath(index: number): string {
	const x1 = rootX + rootW.value
	const y1 = rootY.value + ROOT_H / 2
	const ty = targetY(index)
	const x2 = targetX.value
	const y2 = ty + TARGET_H / 2
	const cx = (x1 + x2) / 2
	return `M ${x1} ${y1} C ${cx} ${y1}, ${cx} ${y2}, ${x2} ${y2}`
}

function pathClass(phase: string): string {
	if (phase === 'Healthy') return 'path-healthy'
	if (phase === 'Unhealthy') return 'path-unhealthy'
	return 'path-unknown'
}

function dotClass(phase: string): string {
	if (phase === 'Healthy') return 'dot-healthy'
	if (phase === 'Unhealthy') return 'dot-unhealthy'
	return 'dot-unknown'
}

function nodeClass(phase: string): string {
	if (phase === 'Healthy') return 'target-rect-healthy'
	if (phase === 'Unhealthy') return 'target-rect-unhealthy'
	return 'target-rect-unknown'
}

// Truncate long URLs for display
function truncateUrl(url: string, max = 32): string {
	if (url.length <= max) return url
	return url.slice(0, max - 1) + '\u2026'
}
</script>

<template>
	<svg :viewBox="`0 0 ${svgW} ${svgH}`" :width="svgW" :height="svgH" class="topology-svg">
		<!-- Connection paths -->
		<path v-for="(server, i) in servers" :key="'path-' + server.url" :d="bezierPath(i)" :class="pathClass(server.phase)"
			fill="none" stroke-width="2" />

		<!-- Root node -->
		<g class="root-node">
			<rect :x="rootX" :y="rootY" :width="rootW" :height="ROOT_H" rx="12" ry="12" class="root-rect" />
			<text :x="rootX + rootW / 2" :y="rootY + 24" text-anchor="middle" class="root-label">
				{{ backendName }}
			</text>
			<text :x="rootX + rootW / 2" :y="rootY + 44" text-anchor="middle" class="root-badge">
				{{ lbType }}
			</text>
		</g>

		<!-- Target server nodes -->
		<g v-for="(server, i) in servers" :key="'node-' + server.url" class="target-node">
			<rect :x="targetX" :y="targetY(i)" :width="TARGET_W" :height="TARGET_H" rx="8" ry="8"
				:class="nodeClass(server.phase)" />
			<!-- Health dot -->
			<circle :cx="targetX + 18" :cy="targetY(i) + TARGET_H / 2" r="6" :class="dotClass(server.phase)" />
			<!-- Server URL -->
			<text :x="targetX + 32" :y="targetY(i) + 21" class="target-url">
				{{ truncateUrl(server.url) }}
			</text>
			<!-- Phase label -->
			<text :x="targetX + 32" :y="targetY(i) + 38" class="target-phase" :class="'phase-' + server.phase.toLowerCase()">
				{{ server.phase }}
			</text>
		</g>
	</svg>
</template>

<style scoped>
.topology-svg {
	display: inline-block;
	font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

/* Root node */
.root-rect {
	fill: #2b5ea7;
	stroke: #3a7bd5;
	stroke-width: 2;
}

.root-label {
	fill: #ffffff;
	font-size: 14px;
	font-weight: 600;
}

.root-badge {
	fill: #a8cfff;
	font-size: 11px;
	font-weight: 400;
}

/* Target nodes */
.target-rect-healthy {
	fill: #1a2e1a;
	stroke: #67c23a;
	stroke-width: 1.5;
}

.target-rect-unhealthy {
	fill: #2e1a1a;
	stroke: #f56c6c;
	stroke-width: 1.5;
}

.target-rect-unknown {
	fill: #1f1f1f;
	stroke: #909399;
	stroke-width: 1.5;
}

.target-url {
	fill: #e0e0e0;
	font-size: 12px;
	font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
}

.target-phase {
	font-size: 10px;
	font-weight: 600;
}

.phase-healthy {
	fill: #67c23a;
}

.phase-unhealthy {
	fill: #f56c6c;
}

.phase-unknown {
	fill: #909399;
}

/* Health dots */
.dot-healthy {
	fill: #67c23a;
	filter: drop-shadow(0 0 4px rgba(103, 194, 58, 0.6));
}

.dot-unhealthy {
	fill: #f56c6c;
	filter: drop-shadow(0 0 4px rgba(245, 108, 108, 0.6));
}

.dot-unknown {
	fill: #909399;
}

/* Connection paths */
.path-healthy {
	stroke: #67c23a;
	stroke-dasharray: 8 4;
	animation: flow 1s linear infinite;
}

.path-unhealthy {
	stroke: #f56c6c;
	stroke-dasharray: 4 4;
	opacity: 0.35;
}

.path-unknown {
	stroke: #909399;
	stroke-dasharray: 4 4;
	opacity: 0.35;
}

@keyframes flow {
	to {
		stroke-dashoffset: -12;
	}
}

@media (prefers-reduced-motion: reduce) {
	.path-healthy {
		animation: none;
	}
}

/* Hover effects */
.target-node {
	transition: opacity 0.15s ease;
}

.target-node:hover {
	opacity: 0.85;
	cursor: default;
}
</style>
