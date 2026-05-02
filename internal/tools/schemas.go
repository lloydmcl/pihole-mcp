package tools

// BlockingStatusOutput is the structured output for pihole_dns_get_blocking.
type BlockingStatusOutput struct {
	Blocking string   `json:"blocking" jsonschema:"Current blocking state: enabled or disabled"`
	Timer    *float64 `json:"timer,omitempty" jsonschema:"Seconds remaining until automatic state revert"`
}

// StatsSummaryOutput is the structured output for pihole_stats_summary.
type StatsSummaryOutput struct {
	TotalQueries     int     `json:"total_queries" jsonschema:"Total DNS queries processed"`
	BlockedQueries   int     `json:"blocked_queries" jsonschema:"Number of blocked queries"`
	PercentBlocked   float64 `json:"percent_blocked" jsonschema:"Percentage of queries blocked"`
	CachedQueries    int     `json:"cached_queries" jsonschema:"Number of cached queries"`
	ForwardedQueries int     `json:"forwarded_queries" jsonschema:"Number of forwarded queries"`
	ActiveClients    int     `json:"active_clients" jsonschema:"Number of active clients"`
	TotalClients     int     `json:"total_clients" jsonschema:"Total known clients"`
	GravityDomains   int     `json:"gravity_domains" jsonschema:"Number of domains in gravity blocklist"`
}

// DomainOutput represents a single domain in structured output.
type DomainOutput struct {
	Domain  string `json:"domain" jsonschema:"Domain name or regex pattern"`
	Type    string `json:"type" jsonschema:"allow or deny"`
	Kind    string `json:"kind" jsonschema:"exact or regex"`
	Enabled bool   `json:"enabled" jsonschema:"Whether the entry is active"`
	Comment string `json:"comment,omitempty" jsonschema:"Optional comment"`
}

// DomainsListOutput is the structured output for pihole_domains_list.
type DomainsListOutput struct {
	Domains []DomainOutput `json:"domains" jsonschema:"List of domain entries"`
	Count   int            `json:"count" jsonschema:"Total number of domains"`
}
