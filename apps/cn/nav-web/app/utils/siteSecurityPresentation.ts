import type { CollectorEnvelope, TargetLatestResponse } from '~/types/nav'
import type { SiteDetailPageData } from '~/composables/useSiteDetailPage'
import { normalizeObservationHeaders, observationRecord as record, observationTime, type ObservationFact } from './siteObservationPresentation'

type Source = Pick<SiteDetailPageData, 'domain' | 'targetLatestCore' | 'targetHealthSummary' | 'lightProbeState'>
type Translate = (key: string) => string
const text = (value: unknown) => typeof value === 'string' ? value.trim() : ''
const strings = (value: unknown) => Array.isArray(value) ? value.map(text).filter(Boolean) : []
const rows = (value: unknown) => Array.isArray(value) ? value.map(record) : []
const bool = (value: unknown) => typeof value === 'boolean' ? value : null
const number = (value: unknown) => typeof value === 'number' && Number.isFinite(value) ? value : null
const count = (value: unknown) => number(value) !== null && (value as number) >= 0 ? value as number : null
const headerCatalog = [
  ['hsts', 'strict_transport_security', 'strict-transport-security', 'HSTS'],
  ['content_security_policy', 'content_security_policy', 'content-security-policy', 'Content-Security-Policy'],
  ['x_frame_options', 'x_frame_options', 'x-frame-options', 'X-Frame-Options'],
  ['x_content_type_options', 'x_content_type_options', 'x-content-type-options', 'X-Content-Type-Options'],
  ['referrer_policy', 'referrer_policy', 'referrer-policy', 'Referrer-Policy'],
  ['permissions_policy', 'permissions_policy', 'permissions-policy', 'Permissions-Policy'],
] as const

/** Shared by Security and Target context: collector defaults are not observations. */
export function readSiteCertificateEvidence(value: unknown): Record<string, unknown> | null {
  const payload = record(value), collected = bool(payload.cert_collected)
  if (collected === false || text(payload.tls_handshake) === 'not_tls') return null
  const observed = collected === true || bool(payload.cert_verified) !== null || number(payload.cert_days_left) !== null
    || Boolean(text(payload.cert_subject_cn) || text(payload.cert_not_after) || text(payload.cert_fingerprint_sha256))
  return observed ? payload : null
}

/** Current Target evidence only. No I/O, client-clock validity calculation or security verdict. */
export function presentSiteSecurity(source: Source, t: Translate) {
  const target = source.domain
  const latest = source.targetLatestCore?.target === target ? source.targetLatestCore : null
  const light = source.lightProbeState?.target === target ? source.lightProbeState : null
  const summary = source.targetHealthSummary?.target === target ? source.targetHealthSummary : null
  const envelope = (snapshot: TargetLatestResponse | null, protocol: string) => {
    const item = snapshot?.protocols?.[protocol]
    return item && (!item.target || item.target === target) && (!item.protocol || item.protocol === protocol) ? item : undefined
  }
  const state = (value: string) => ({ state: value, label: t('siteSecurity.states.' + value) })
  const display = (value: unknown): string => typeof value === 'boolean' ? t('siteSecurity.' + (value ? 'yes' : 'no'))
    : number(value) !== null ? String(value) : Array.isArray(value) ? strings(value).join('\n') || '—' : text(value) || '—'
  const fact = (key: string, value: unknown): ObservationFact => ({ key, label: t('siteSecurity.fields.' + key), value: display(value) })
  const facts = (data: Record<string, unknown>, keys: string[]) => keys.map(key => fact(key, data[key]))
  const compact = (items: ObservationFact[]) => items.filter(item => item.value !== '—')
  const ms = (value: unknown) => count(value) === null ? '—' : `${value} ms`
  const statusOf = (item?: CollectorEnvelope) => !item ? 'not_observed' : item.status === 'failure' ? 'unavailable'
    : item.status === 'skipped' ? 'skipped' : item.status === 'success' ? 'present' : 'unknown'
  const meta = (item?: CollectorEnvelope, stale = false) => ({ ...state(statusOf(item)), stale,
    observed: observationTime(item?.observed_at),
    facts: [fact('observed_at', observationTime(item?.observed_at)), fact('duration_ms', ms(item?.duration_ms))],
    errors: compact([fact('error_code', item?.error_code), fact('error_message', item?.error_message)]),
    truncated: item?.payload_truncated === true,
  })
  const http = envelope(latest, 'http'), payload = record(http?.payload)
  const httpMeta = meta(http, summary?.protocols?.http?.stale === true || latest?.state === 'stale')
  const handshake = text(payload.tls_handshake)
  const transportState = handshake === 'failed' ? 'failed' : handshake === 'not_tls' ? 'not_applicable'
    : text(payload.tls_version) || handshake === 'collected' ? 'present' : httpMeta.state === 'present' ? 'not_observed' : httpMeta.state
  const transport = { ...httpMeta, ...state(transportState), facts: [fact('tls_version', payload.tls_version), fact('cipher_suite', payload.cipher_suite),
    fact('tls_handshake', handshake), fact('observed_at', httpMeta.observed)] }

  const collected = bool(payload.cert_collected)
  const evidence = readSiteCertificateEvidence(payload), hasCertificate = evidence !== null
  const cert = evidence ?? {}
  const verified = bool(cert.cert_verified), days = number(cert.cert_days_left)
  const verification = verified === true ? 'verified' : verified === false ? 'failed'
    : httpMeta.state === 'unavailable' && !hasCertificate ? 'unavailable' : handshake === 'not_tls' ? 'not_applicable' : 'not_observed'
  const expiry = days === null ? 'not_observed' : days <= 0 ? 'expired' : days <= 7 ? 'warning' : days <= 30 ? 'attention' : 'normal'
  const certificate = {
    verification: state(verification), expiry: state(expiry), days, collected,
    errors: compact(facts(cert, ['verify_error_category', 'verify_error'])),
    validity: [fact('cert_not_before', observationTime(cert.cert_not_before)), fact('cert_not_after', observationTime(cert.cert_not_after)), fact('cert_days_left', days)],
    subject: facts(cert, ['cert_subject_cn', 'cert_subject_org', 'cert_san_count']), san: strings(cert.cert_dns_names),
    issuer: facts(cert, ['cert_issuer_cn', 'cert_issuer_org', 'cert_chain_length']), chain: strings(cert.cert_chain_issuers),
    crypto: facts(cert, ['cert_public_key_algorithm', 'cert_public_key_bits', 'cert_signature_algorithm', 'ocsp_stapled', 'sct_count', 'cert_serial_number', 'cert_fingerprint_sha256', 'cert_spki_sha256']),
  }

  const headerSummary = record(payload.security_header_summary), headerFlags = record(payload.security_headers)
  const rawHeaders = normalizeObservationHeaders(payload.headers)
  const hasRawHeaders = payload.headers !== null && typeof payload.headers === 'object' && !Array.isArray(payload.headers)
  const headerRows = headerCatalog.map(([key, flag, rawKey, label]) => {
    const item = record(headerSummary[key])
    const value = text(headerFlags[flag]) || rawHeaders.find(row => row.key === rawKey)?.value || ''
    const present = bool(item.present) ?? bool(headerFlags[flag]) ?? (typeof headerFlags[flag] === 'string' || hasRawHeaders ? Boolean(value) : null)
    const headerState = httpMeta.state !== 'present' ? httpMeta.state : present === null ? 'not_observed' : present ? 'present' : 'missing'
    return { key, ...state(headerState), name: label, value: value || '—',
      details: compact(facts(item, ['max_age', 'include_subdomains', 'preload', 'has_default_src', 'unsafe_inline', 'unsafe_eval', 'wildcard_source', 'mode', 'nosniff', 'policy', 'directive_count'])) }
  })
  const headers = { ...httpMeta, rows: headerRows }

  const txtEnvelope = envelope(light, 'security_txt'), txt = record(txtEnvelope?.payload)
  const txtMeta = meta(txtEnvelope, light?.state === 'stale'), validation = strings(txt.validation_errors)
  const txtState = txtMeta.state !== 'present' ? txtMeta.state : bool(txt.exists) === true ? validation.length ? 'found_with_issues' : 'found'
    : bool(txt.exists) === false ? 'not_found' : 'unknown'
  const securityTxt = { ...txtMeta, ...state(txtState), metaFacts: txtMeta.facts, validation: txtMeta.state === 'present' ? validation : [],
    facts: facts(txt, ['exists', 'recognition', 'path_used', 'status_code', 'content_type', 'contact', 'expires', 'policy', 'canonical', 'preferred_languages']),
    truncated: txtMeta.truncated || txt.body_truncated === true }

  const portEnvelope = envelope(light, 'port_check'), ports = record(portEnvelope?.payload)
  const portMeta = meta(portEnvelope, light?.state === 'stale')
  const portRows = rows(ports.results).map((item, index) => ({ key: String(index),
    status: ['open', 'closed', 'timeout', 'filtered_suspected', 'skipped'].includes(text(item.status)) ? text(item.status) : 'unknown',
    tone: 'neutral' as const,
    facts: [fact('port', item.port), fact('service_hint', item.service_hint), fact('status', t('siteSecurity.states.' + (['open', 'closed', 'timeout', 'filtered_suspected', 'skipped'].includes(text(item.status)) ? text(item.status) : 'unknown'))), fact('duration_ms', ms(item.duration_ms))],
    errors: compact(facts(item, ['error_code', 'error_message'])) }))
  const portCheck = { ...portMeta, ...state(portMeta.state === 'present' && Array.isArray(ports.results) && !portRows.length ? 'empty' : portMeta.state),
    summary: facts(ports, ['ports_configured', 'ports_checked', 'open_count', 'closed_count', 'timeout_count', 'filtered_suspected_count']),
    results: portRows, metadata: facts(ports, ['skipped_count', 'invalid_port_count', 'duplicate_port_count', 'truncated_port_count', 'truncated', 'skipped_reason']),
    truncated: portMeta.truncated || ports.truncated === true }

  const wafEnvelope = envelope(light, 'waf_canary'), waf = record(wafEnvelope?.payload)
  const wafMeta = meta(wafEnvelope, light?.state === 'stale')
  const truncated = wafMeta.truncated || waf.target_run_truncated === true || (count(waf.truncated_target_count) ?? 0) > 0
  const mismatchKeys = ['unexpected_pass_count', 'network_error_count', 'status_code_unexpected_count']
  const mismatches = wafMeta.state === 'present' ? mismatchKeys.filter(key => (count(waf[key]) ?? 0) > 0) : []
  // Matching is bounded to this complete, successful sample, never proof of a WAF.
  const matched = wafMeta.state === 'present' && !truncated && (count(waf.expected_blocked_count) ?? 0) > 0
    && count(waf.expected_blocked_matched_count) === count(waf.expected_blocked_count)
    && count(waf.cases_executed) === count(waf.cases_total) && (count(waf.cases_total) ?? 0) > 0
    && mismatchKeys.every(key => count(waf[key]) === 0)
  const wafState = wafMeta.state !== 'present' ? wafMeta.state : truncated ? 'incomplete' : mismatches.length ? 'mismatch' : matched ? 'matched' : 'present'
  const wafCanary = { ...wafMeta, ...state(wafState), truncated,
    summary: facts(waf, ['cases_total', 'cases_executed', 'blocked_count', 'expected_blocked_count', 'expected_blocked_matched_count', ...mismatchKeys]),
    metadata: facts(waf, ['target_run_truncated', 'truncated_target_count']),
    cases: rows(waf.cases).map((item, index) => ({ key: String(index), facts: [...facts(item, ['case_id', 'category', 'method', 'status_code', 'blocked', 'expected_blocked', 'matched_expected']), fact('duration_ms', ms(item.duration_ms))], errors: compact(facts(item, ['error_code', 'error_message'])) })),
  }

  const attention: { key: string; message: string; detail: string }[] = []
  const addAttention = (key: string, detail = '') => attention.push({ key, message: t('siteSecurity.attention.' + key), detail })
  if (verification === 'failed') addAttention('verification', text(cert.verify_error) || text(cert.verify_error_category))
  if (['attention', 'warning', 'expired'].includes(expiry)) addAttention('expiry', `${days} ${t('siteDetail.days')}`)
  if (txtState === 'found_with_issues') addAttention('securityTxt', validation.join('; '))
  for (const key of mismatches) addAttention(key, display(waf[key]))
  const overview = { attention, sections: [
    { key: 'transport', title: t('siteSecurity.transport'), value: transport.label, detail: [text(payload.tls_version), handshake].filter(Boolean).join(' · ') || '—' },
    { key: 'certificate', title: t('siteSecurity.certificate'), value: `${t('siteSecurity.verification')}: ${certificate.verification.label}`, detail: days === null ? '—' : `${days} ${t('siteDetail.days')}` },
    { key: 'headers', title: t('siteSecurity.headers'), value: headerRows.map(row => `${row.name}: ${row.label}`).join('\n'), detail: '' },
    { key: 'securityTxt', title: 'security.txt', value: securityTxt.label, detail: '' },
    { key: 'portCheck', title: t('siteSecurity.portCheck'), value: portCheck.label, detail: `${t('siteSecurity.fields.ports_checked')}: ${display(ports.ports_checked)}` },
    { key: 'wafCanary', title: t('siteSecurity.wafCanary'), value: wafCanary.label, detail: '' },
  ] }
  return { target, overview, transport, certificate, headers, securityTxt, portCheck, wafCanary }
}
export type SiteSecurityPresentation = ReturnType<typeof presentSiteSecurity>
