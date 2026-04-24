<script setup lang="ts">
import { onMounted, computed, ref, watch } from 'vue'
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
import type { Metricsv1Label, V1MetricBucket, V1MetricSeries } from '@/generated/metrics'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend, Filler)

const backendStore = useBackendStore()
const routeStore = useRouteStore()
const caStore = useCaStore()
const credentialStore = useCredentialStore()
const certificateStore = useCertificateStore()
const policyStore = usePolicyStore()
const metricsStore = useMetricsStore()

const selectedRoute = ref<string[]>([])
const selectedBackend = ref<string[]>([])

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

function labelValue(labels: Metricsv1Label[] | undefined, name: string): string {
	return labels?.find(label => label.name === name)?.value ?? ''
}

function seriesLabel(series: V1MetricSeries, primary: string): string {
	const value = labelValue(series.labels, primary)
	if (value) return value
	const fallback = (series.labels ?? [])
		.map(label => `${label.name}=${label.value}`)
		.join(', ')
	return fallback || series.metric || 'series'
}

function sortedBuckets(series?: V1MetricSeries) {
	return [...(series?.buckets ?? [])].sort((a, b) => (a.start?.getTime() ?? 0) - (b.start?.getTime() ?? 0))
}

function counterValue(bucket: V1MetricBucket): number {
	return Number(bucket.value ?? 0)
}

function gaugeValue(bucket: V1MetricBucket): number {
	return Number(bucket.value ?? 0)
}

function histogramAverage(bucket: V1MetricBucket): number {
	const count = Number(bucket.count ?? 0)
	return count > 0 ? Number(bucket.sum ?? 0) / count : 0
}

function aggregateSeriesBuckets(seriesList: V1MetricSeries[]): V1MetricBucket[] {
	const buckets = new Map<number, V1MetricBucket>()
	for (const series of seriesList) {
		for (const bucket of sortedBuckets(series)) {
			const startTime = bucket.start?.getTime()
			if (startTime == null) continue
			const existing = buckets.get(startTime)
			if (existing) {
				existing.value = Number(existing.value ?? 0) + Number(bucket.value ?? 0)
				existing.count = String(Number(existing.count ?? 0) + Number(bucket.count ?? 0))
				existing.sum = Number(existing.sum ?? 0) + Number(bucket.sum ?? 0)
				continue
			}
			buckets.set(startTime, {
				start: bucket.start,
				value: Number(bucket.value ?? 0),
				count: bucket.count == null ? undefined : String(Number(bucket.count ?? 0)),
				sum: Number(bucket.sum ?? 0),
			})
		}
	}
	return [...buckets.values()].sort((a, b) => (a.start?.getTime() ?? 0) - (b.start?.getTime() ?? 0))
}

function chartFromGroupedSeries(
	groups: Array<{ label: string, series: V1MetricSeries[] }>,
	color: string,
	valueForBucket: (bucket: V1MetricBucket) => number,
) {
	const aggregatedGroups = groups.map(group => ({
		label: group.label,
		buckets: aggregateSeriesBuckets(group.series),
	}))
	const labels = aggregatedGroups[0]?.buckets.map(bucket => formatTime(bucket.start)) ?? []
	return {
		labels,
		datasets: aggregatedGroups.map((group, index) => ({
			label: group.label,
			data: group.buckets.map(valueForBucket),
			borderColor: index === 0 ? color : palette[index % palette.length],
			backgroundColor: (index === 0 ? color : palette[index % palette.length]) + '1a',
			fill: true,
		})),
	}
}

const palette = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#909399', '#b37feb']

const chartOptions = {
	responsive: true,
	maintainAspectRatio: false,
	interaction: { intersect: false, mode: 'index' as const },
	plugins: {
		legend: { display: true, position: 'bottom' as const },
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
		line: { tension: 0.2, borderWidth: 1, stepped: false },
	},
}

const allSeries = computed(() => metricsStore.data?.series ?? [])

const routeOptions = computed(() => {
	const values = new Set<string>()
	for (const series of allSeries.value) {
		const route = labelValue(series.labels, 'route')
		if (route && route !== '__unmatched__') values.add(route)
	}
	return [...values].sort()
})

const backendOptions = computed(() => {
	const values = new Set<string>()
	for (const series of allSeries.value) {
		const backend = labelValue(series.labels, 'backend')
		if (backend) values.add(backend)
	}
	return [...values].sort()
})

watch(routeOptions, values => {
	selectedRoute.value = selectedRoute.value.filter(value => values.includes(value))
}, { immediate: true })

watch(backendOptions, values => {
	selectedBackend.value = selectedBackend.value.filter(value => values.includes(value))
}, { immediate: true })

function seriesByMetric(metric: string) {
	return allSeries.value.filter(series => series.metric === metric)
}

function seriesByMetricAndLabel(metric: string, labelName: string, labelValueText: string) {
	return seriesByMetric(metric).filter(series => labelValue(series.labels, labelName) === labelValueText)
}

function seriesByMetricWithoutLabel(metric: string, labelName: string) {
	return seriesByMetric(metric).filter(series => !labelValue(series.labels, labelName))
}

function groupedSeriesByMetricSelection(
	metric: string,
	labelName: string,
	selectedValues: string[],
	allLabel: string,
	options?: { exclude?: string },
) {
	const grouped = new Map<string, V1MetricSeries[]>()
	for (const series of seriesByMetric(metric)) {
		const value = labelValue(series.labels, labelName)
		if (!value || value === options?.exclude) continue
		const existing = grouped.get(value)
		if (existing) {
			existing.push(series)
		} else {
			grouped.set(value, [series])
		}
	}

	if (selectedValues.length === 0) {
		return grouped.size > 0
			? [{ label: allLabel, series: [...grouped.values()].flat() }]
			: []
	}

	return selectedValues
		.map(value => {
			const series = grouped.get(value)
			if (!series || series.length === 0) return null
			return { label: value, series }
		})
		.filter((group): group is { label: string, series: V1MetricSeries[] } => group !== null)
}

const requestRateSeries = computed(() => groupedSeriesByMetricSelection('proxy.http.requests.total', 'route', selectedRoute.value, 'All routes', { exclude: '__unmatched__' }))
const requestDurationSeries = computed(() => groupedSeriesByMetricSelection('proxy.http.request.duration.seconds', 'route', selectedRoute.value, 'All routes', { exclude: '__unmatched__' }))
const requestSizeSeries = computed(() => groupedSeriesByMetricSelection('proxy.http.request.size.bytes', 'route', selectedRoute.value, 'All routes', { exclude: '__unmatched__' }))
const responseSizeSeries = computed(() => groupedSeriesByMetricSelection('proxy.http.response.size.bytes', 'route', selectedRoute.value, 'All routes', { exclude: '__unmatched__' }))

const backendRequestSeries = computed(() => groupedSeriesByMetricSelection('proxy.backend.requests.total', 'backend', selectedBackend.value, 'All backends'))
const backendDurationSeries = computed(() => groupedSeriesByMetricSelection('proxy.backend.request.duration.seconds', 'backend', selectedBackend.value, 'All backends'))
const backendActiveSeries = computed(() => groupedSeriesByMetricSelection('proxy.backend.active.requests', 'backend', selectedBackend.value, 'All backends'))

const authSeries = computed(() => seriesByMetric('proxy.auth.requests.total'))
const policySeries = computed(() => seriesByMetric('proxy.policy.evaluations.total'))
const routeMatchSeries = computed(() => seriesByMetric('proxy.route.matches.total'))
const routeNoMatchSeries = computed(() => seriesByMetric('proxy.route.no_match.total'))
const activeRequestsSeries = computed(() => seriesByMetricWithoutLabel('proxy.http.active.requests', 'route'))

const requestRateChart = computed(() => chartFromGroupedSeries(requestRateSeries.value, '#409eff', counterValue))
const requestDurationChart = computed(() => chartFromGroupedSeries(requestDurationSeries.value, '#e6a23c', histogramAverage))
const activeRequestsChart = computed(() => chartFromGroupedSeries(activeRequestsSeries.value.map(series => ({ label: seriesLabel(series, 'route'), series: [series] })), '#67c23a', gaugeValue))
const requestSizeChart = computed(() => chartFromGroupedSeries(requestSizeSeries.value, '#909399', histogramAverage))
const responseSizeChart = computed(() => chartFromGroupedSeries(responseSizeSeries.value, '#909399', histogramAverage))

const backendRequestsChart = computed(() => chartFromGroupedSeries(backendRequestSeries.value, '#409eff', counterValue))
const backendDurationChart = computed(() => chartFromGroupedSeries(backendDurationSeries.value, '#e6a23c', histogramAverage))
const backendActiveChart = computed(() => chartFromGroupedSeries(backendActiveSeries.value, '#67c23a', gaugeValue))

const authChart = computed(() => chartFromGroupedSeries(authSeries.value.map(series => ({ label: seriesLabel(series, 'result'), series: [series] })), '#f56c6c', counterValue))
const policyChart = computed(() => chartFromGroupedSeries(policySeries.value.map(series => ({ label: seriesLabel(series, 'result'), series: [series] })), '#b37feb', counterValue))
const routeMatchesChart = computed(() => chartFromGroupedSeries(routeMatchSeries.value.map(series => ({ label: seriesLabel(series, 'match_kind'), series: [series] })), '#67c23a', counterValue))
const routeNoMatchChart = computed(() => chartFromGroupedSeries(routeNoMatchSeries.value.map(series => ({ label: seriesLabel(series, 'route'), series: [series] })), '#f56c6c', counterValue))
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


		<el-skeleton v-if="metricsStore.loading" :rows="6" animated style="margin-top: 12px" />
		<template v-else-if="metricsStore.data">

			<el-row justify="space-between" align="middle" style="margin-bottom: 16px">
				<el-col :span="12">
					<div style="display: flex; align-items: center; gap: 12px">
						<h2>Requests</h2>
						<div>
							<el-select v-model="selectedRoute" placeholder="Select routes" style="width: 280px" multiple collapse-tags
								collapse-tags-tooltip clearable>
								<el-option v-for="route in routeOptions" :key="route" :label="route" :value="route" />
							</el-select>
						</div>
					</div>
				</el-col>
				<el-col :span="12" style="text-align: left">
				</el-col>
			</el-row>

			<el-row :gutter="20">
				<el-col :xs="24" :md="8" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Requests Total</span></template>
						<div style="height: 180px">
							<Line :data="requestRateChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
				<el-col :xs="24" :md="8" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Request Duration (avg)</span></template>
						<div style="height: 180px">
							<Line :data="requestDurationChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
				<el-col :xs="24" :md="8" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Active Requests (max)</span></template>
						<div style="height: 180px">
							<Line :data="activeRequestsChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
				<el-col :xs="24" :md="8" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Request Size (avg bytes)</span></template>
						<div style="height: 180px">
							<Line :data="requestSizeChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
				<el-col :xs="24" :md="8" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Response Size (avg bytes)</span></template>
						<div style="height: 180px">
							<Line :data="responseSizeChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
			</el-row>

			<el-row justify="space-between" align="middle" style="margin-bottom: 16px">
				<el-col :span="12">
					<div style="display: flex; align-items: center; gap: 12px">
						<h2>Backend</h2>
						<div>
							<el-select v-model="selectedBackend" placeholder="Select backends" style="width: 280px" multiple
								collapse-tags collapse-tags-tooltip clearable>
								<el-option v-for="backend in backendOptions" :key="backend" :label="backend" :value="backend" />
							</el-select>
						</div>
					</div>
				</el-col>
				<el-col :span="12" style="text-align: left">
				</el-col>
			</el-row>

			<el-row :gutter="20">
				<el-col :xs="24" :md="8" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Backend Requests Total</span></template>
						<div style="height: 180px">
							<Line :data="backendRequestsChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
				<el-col :xs="24" :md="8" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Backend Request Duration (avg)</span></template>
						<div style="height: 180px">
							<Line :data="backendDurationChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
				<el-col :xs="24" :md="8" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Backend Active Requests (max)</span></template>
						<div style="height: 180px">
							<Line :data="backendActiveChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
			</el-row>

			<h2>Auth &amp; Policy</h2>
			<el-row :gutter="20">
				<el-col :xs="24" :md="8" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Auth Requests Total</span></template>
						<div style="height: 180px">
							<Line :data="authChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
				<el-col :xs="24" :md="8" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Policy Evaluations Total</span></template>
						<div style="height: 180px">
							<Line :data="policyChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
			</el-row>

			<h2>Routing</h2>
			<el-row :gutter="20">
				<el-col :xs="24" :md="8" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Route Matches Total</span></template>
						<div style="height: 180px">
							<Line :data="routeMatchesChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
				<el-col :xs="24" :md="8" style="margin-bottom: 20px">
					<el-card shadow="hover">
						<template #header><span>Route No-Match Total</span></template>
						<div style="height: 180px">
							<Line :data="routeNoMatchChart" :options="chartOptions" />
						</div>
					</el-card>
				</el-col>
			</el-row>
		</template>
	</div>
</template>
