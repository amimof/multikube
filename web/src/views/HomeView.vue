<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { Monitor, Guide, Key, Lock, Document, List } from '@element-plus/icons-vue'
import { Line } from 'vue-chartjs'
import {
	Chart as ChartJS,
	CategoryScale,
	LinearScale,
	PointElement,
	LineElement,
	Title,
	Tooltip,
	Legend,
	Filler,
} from 'chart.js'
import { useBackendStore } from '@/stores/backend'
import { useRouteStore } from '@/stores/route'
import { useCaStore } from '@/stores/ca'
import { useCredentialStore } from '@/stores/credential'
import { useCertificateStore } from '@/stores/certificate'
import { usePolicyStore } from '@/stores/policy'
import { useMetricsStore } from '@/stores/metrics'
import type { V1Int64Series, V1Float64Series, V1Int64HistogramSeries, V1GaugeSeries } from '@/generated/metrics'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler)

const backendStore = useBackendStore()
const routeStore = useRouteStore()
const caStore = useCaStore()
const credentialStore = useCredentialStore()
const certificateStore = useCertificateStore()
const policyStore = usePolicyStore()
const metricsStore = useMetricsStore()

onMounted(() => {
	backendStore.fetchBackends().catch(() => { })
	routeStore.fetchRoutes().catch(() => { })
	caStore.fetchCas().catch(() => { })
	credentialStore.fetchCredentials().catch(() => { })
	certificateStore.fetchCertificates().catch(() => { })
	policyStore.fetchPolicies().catch(() => { })
	metricsStore.fetchMetrics().catch(() => { })
})

function formatTime(d?: Date): string {
	if (!d) return ''
	return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

const chartOptions = {
	responsive: true,
	maintainAspectRatio: false,
	interaction: { intersect: false, mode: 'index' as const },
	plugins: {
		legend: { display: false },
	},
	scales: {
		x: {
			display: false,
			grid: { display: false },
			ticks: { maxTicksLimit: 8, font: { size: 11 } },
		},
		y: {
			grid: { display: false },
			beginAtZero: true,
			ticks: { font: { size: 11 } },
		},
	},
	elements: {
		point: { radius: 0, hoverRadius: 4 },
		line: { tension: 0.4, borderWidth: 1.5, stepped: false },
	},
}

function makeCounterChart(buckets: V1Int64Series[] | undefined, color: string) {
	const sorted = [...(buckets ?? [])].sort((a, b) => (a.start?.getTime() ?? 0) - (b.start?.getTime() ?? 0))
	return {
		labels: sorted.map(b => formatTime(b.start)),
		datasets: [{
			data: sorted.map(b => Number(b.value ?? 0)),
			borderColor: color,
			backgroundColor: color + '1a',
			fill: true,
		}],
	}
}

function makeHistogramCountChart(buckets: V1Float64Series[] | undefined, color: string) {
	const sorted = [...(buckets ?? [])].sort((a, b) => (a.start?.getTime() ?? 0) - (b.start?.getTime() ?? 0))
	return {
		labels: sorted.map(b => formatTime(b.start)),
		datasets: [{
			label: 'Avg',
			data: sorted.map(b => {
				const count = Number(b.count ?? 0)
				return count > 0 ? (b.sum ?? 0) / count : 0
			}),
			borderColor: color,
			backgroundColor: color + '1a',
			fill: true,
		}],
	}
}

function makeInt64HistogramChart(buckets: V1Int64HistogramSeries[] | undefined, color: string) {
	const sorted = [...(buckets ?? [])].sort((a, b) => (a.start?.getTime() ?? 0) - (b.start?.getTime() ?? 0))
	return {
		labels: sorted.map(b => formatTime(b.start)),
		datasets: [{
			label: 'Avg',
			data: sorted.map(b => {
				const count = Number(b.count ?? 0)
				return count > 0 ? Number(b.sum ?? 0) / count : 0
			}),
			borderColor: color,
			backgroundColor: color + '1a',
			fill: true,
		}],
	}
}

function makeGaugeChart(buckets: V1GaugeSeries[] | undefined, color: string) {
	const sorted = [...(buckets ?? [])].sort((a, b) => (a.start?.getTime() ?? 0) - (b.start?.getTime() ?? 0))
	return {
		labels: sorted.map(b => formatTime(b.start)),
		datasets: [{
			data: sorted.map(b => Number(b.max ?? 0)),
			borderColor: color,
			backgroundColor: color + '1a',
			fill: true,
		}],
	}
}

// Requests section
const requestsTotalChart = computed(() => makeCounterChart(metricsStore.data?.requestsTotal?.buckets, '#409eff'))
const requestDurationChart = computed(() => makeHistogramCountChart(metricsStore.data?.requestDuration?.buckets, '#e6a23c'))
const activeRequestsChart = computed(() => makeGaugeChart(metricsStore.data?.activeRequests?.buckets, '#67c23a'))
const requestSizeChart = computed(() => makeInt64HistogramChart(metricsStore.data?.requestSizeBytes?.buckets, '#909399'))
const responseSizeChart = computed(() => makeInt64HistogramChart(metricsStore.data?.responseSizeBytes?.buckets, '#909399'))

// Backend section
const backendRequestsTotalChart = computed(() => makeCounterChart(metricsStore.data?.backendRequestsTotal?.buckets, '#409eff'))
const backendRequestDurationChart = computed(() => makeHistogramCountChart(metricsStore.data?.backendRequestDuration?.buckets, '#e6a23c'))
const backendActiveRequestsChart = computed(() => makeGaugeChart(metricsStore.data?.backendActiveRequests?.buckets, '#67c23a'))

// Auth & Policy section
const authRequestsTotalChart = computed(() => makeCounterChart(metricsStore.data?.authRequestsTotal?.buckets, '#f56c6c'))
const policyEvaluationsTotalChart = computed(() => makeCounterChart(metricsStore.data?.policyEvaluationsTotal?.buckets, '#b37feb'))

// Routing section
const routeMatchesTotalChart = computed(() => makeCounterChart(metricsStore.data?.routeMatchesTotal?.buckets, '#67c23a'))
const routeNoMatchTotalChart = computed(() => makeCounterChart(metricsStore.data?.routeNoMatchTotal?.buckets, '#f56c6c'))
</script>

<template>
	<div>
		<h1 style="margin-bottom: 24px">Dashboard</h1>
		<el-row :gutter="20">
			<el-col :xs="24" :sm="12" :md="8" style="margin-bottom: 20px">
				<router-link to="/backends" style="text-decoration: none">
					<el-card shadow="hover">
						<template #header>
							<div style="display: flex; align-items: center; gap: 8px">
								<el-icon :size="20">
									<Monitor />
								</el-icon>
								<span>Backends</span>
							</div>
						</template>
						<div style="text-align: center">
							<el-skeleton v-if="backendStore.loading" :rows="0" animated>
								<template #template>
									<el-skeleton-item variant="text" style="width: 60px; height: 32px; margin: 0 auto" />
								</template>
							</el-skeleton>
							<el-statistic v-else :value="backendStore.items.length" />
						</div>
					</el-card>
				</router-link>
			</el-col>
			<el-col :xs="24" :sm="12" :md="8" style="margin-bottom: 20px">
				<router-link to="/routes" style="text-decoration: none">
					<el-card shadow="hover">
						<template #header>
							<div style="display: flex; align-items: center; gap: 8px">
								<el-icon :size="20">
									<Guide />
								</el-icon>
								<span>Routes</span>
							</div>
						</template>
						<div style="text-align: center">
							<el-skeleton v-if="routeStore.loading" :rows="0" animated>
								<template #template>
									<el-skeleton-item variant="text" style="width: 60px; height: 32px; margin: 0 auto" />
								</template>
							</el-skeleton>
							<el-statistic v-else :value="routeStore.items.length" />
						</div>
					</el-card>
				</router-link>
			</el-col>
			<el-col :xs="24" :sm="12" :md="8" style="margin-bottom: 20px">
				<router-link to="/cas" style="text-decoration: none">
					<el-card shadow="hover">
						<template #header>
							<div style="display: flex; align-items: center; gap: 8px">
								<el-icon :size="20">
									<Key />
								</el-icon>
								<span>Certificate Authorities</span>
							</div>
						</template>
						<div style="text-align: center">
							<el-skeleton v-if="caStore.loading" :rows="0" animated>
								<template #template>
									<el-skeleton-item variant="text" style="width: 60px; height: 32px; margin: 0 auto" />
								</template>
							</el-skeleton>
							<el-statistic v-else :value="caStore.items.length" />
						</div>
					</el-card>
				</router-link>
			</el-col>
			<el-col :xs="24" :sm="12" :md="8" style="margin-bottom: 20px">
				<router-link to="/credentials" style="text-decoration: none">
					<el-card shadow="hover">
						<template #header>
							<div style="display: flex; align-items: center; gap: 8px">
								<el-icon :size="20">
									<Lock />
								</el-icon>
								<span>Credentials</span>
							</div>
						</template>
						<div style="text-align: center">
							<el-skeleton v-if="credentialStore.loading" :rows="0" animated>
								<template #template>
									<el-skeleton-item variant="text" style="width: 60px; height: 32px; margin: 0 auto" />
								</template>
							</el-skeleton>
							<el-statistic v-else :value="credentialStore.items.length" />
						</div>
					</el-card>
				</router-link>
			</el-col>
			<el-col :xs="24" :sm="12" :md="8" style="margin-bottom: 20px">
				<router-link to="/certificates" style="text-decoration: none">
					<el-card shadow="hover">
						<template #header>
							<div style="display: flex; align-items: center; gap: 8px">
								<el-icon :size="20">
									<Document />
								</el-icon>
								<span>Certificates</span>
							</div>
						</template>
						<div style="text-align: center">
							<el-skeleton v-if="certificateStore.loading" :rows="0" animated>
								<template #template>
									<el-skeleton-item variant="text" style="width: 60px; height: 32px; margin: 0 auto" />
								</template>
							</el-skeleton>
							<el-statistic v-else :value="certificateStore.items.length" />
						</div>
					</el-card>
				</router-link>
			</el-col>
			<el-col :xs="24" :sm="12" :md="8" style="margin-bottom: 20px">
				<router-link to="/policies" style="text-decoration: none">
					<el-card shadow="hover">
						<template #header>
							<div style="display: flex; align-items: center; gap: 8px">
								<el-icon :size="20">
									<List />
								</el-icon>
								<span>Policies</span>
							</div>
						</template>
						<div style="text-align: center">
							<el-skeleton v-if="policyStore.loading" :rows="0" animated>
								<template #template>
									<el-skeleton-item variant="text" style="width: 60px; height: 32px; margin: 0 auto" />
								</template>
							</el-skeleton>
							<el-statistic v-else :value="policyStore.items.length" />
						</div>
					</el-card>
				</router-link>
			</el-col>
		</el-row>

		<!-- Metrics Graphs -->
		<el-skeleton v-if="metricsStore.loading" :rows="6" animated style="margin-top: 12px" />
		<template v-else-if="metricsStore.data">
			<!-- Requests -->
			<h2 style="margin: 32px 0 16px">Requests</h2>
			<el-row :gutter="20">
				<el-col :xs="24" :md="12" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Requests Total</span></template>
						<div style="height: 140px">
							<Line :data="requestsTotalChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
				<el-col :xs="24" :md="12" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Request Duration (avg)</span></template>
						<div style="height: 140px">
							<Line :data="requestDurationChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
				<el-col :xs="24" :md="12" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Active Requests (max)</span></template>
						<div style="height: 140px">
							<Line :data="activeRequestsChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
				<el-col :xs="24" :md="12" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Request Size (avg bytes)</span></template>
						<div style="height: 140px">
							<Line :data="requestSizeChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
				<el-col :xs="24" :md="12" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Response Size (avg bytes)</span></template>
						<div style="height: 140px">
							<Line :data="responseSizeChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
			</el-row>

			<!-- Backend -->
			<h2 style="margin: 32px 0 16px">Backend</h2>
			<el-row :gutter="20">
				<el-col :xs="24" :md="12" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Backend Requests Total</span></template>
						<div style="height: 140px">
							<Line :data="backendRequestsTotalChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
				<el-col :xs="24" :md="12" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Backend Request Duration (avg)</span></template>
						<div style="height: 140px">
							<Line :data="backendRequestDurationChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
				<el-col :xs="24" :md="12" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Backend Active Requests (max)</span></template>
						<div style="height: 140px">
							<Line :data="backendActiveRequestsChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
			</el-row>

			<!-- Auth & Policy -->
			<h2 style="margin: 32px 0 16px">Auth &amp; Policy</h2>
			<el-row :gutter="20">
				<el-col :xs="24" :md="12" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Auth Requests Total</span></template>
						<div style="height: 140px">
							<Line :data="authRequestsTotalChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
				<el-col :xs="24" :md="12" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Policy Evaluations Total</span></template>
						<div style="height: 140px">
							<Line :data="policyEvaluationsTotalChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
			</el-row>

			<!-- Routing -->
			<h2 style="margin: 32px 0 16px">Routing</h2>
			<el-row :gutter="20">
				<el-col :xs="24" :md="12" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Route Matches Total</span></template>
						<div style="height: 140px">
							<Line :data="routeMatchesTotalChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
				<el-col :xs="24" :md="12" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Route No-Match Total</span></template>
						<div style="height: 140px">
							<Line :data="routeNoMatchTotalChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
			</el-row>
		</template>
	</div>
</template>
