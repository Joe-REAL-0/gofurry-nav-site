/** SiteVo keeps the Home domain field (plain text or the stored domain JSON). */
export function siteRecommendationDomain(value: string): string {
  let parsed: unknown
  try { parsed = JSON.parse(value) } catch { return value }
  const domains = Array.isArray(parsed) ? parsed
    : parsed && typeof parsed === 'object' && 'domain' in parsed ? parsed.domain : []
  return Array.isArray(domains) ? domains.find((domain): domain is string => typeof domain === 'string' && Boolean(domain.trim())) ?? ''
    : typeof parsed === 'string' ? parsed : ''
}
