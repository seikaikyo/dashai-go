package career

// ContextMeta holds display metadata for a career context.
type ContextMeta struct {
	Key         ContextType `json:"key"`
	Name        string      `json:"name"`
	Subject1    string      `json:"subject1"`
	Subject2    string      `json:"subject2"`
	Description string      `json:"description"`
}

// Contexts defines all six career analysis scenarios.
var Contexts = map[ContextType]ContextMeta{
	ContextEmployment: {
		Key:         ContextEmployment,
		Name:        "direct_employment",
		Subject1:    "job_seeker",
		Subject2:    "company",
		Description: "Long-term employment: daily friction, career development, retention",
	},
	ContextConsulting: {
		Key:         ContextConsulting,
		Name:        "consulting",
		Subject1:    "consultant",
		Subject2:    "client_site",
		Description: "Short-term project: communication efficiency, deliverables, exit timing",
	},
	ContextOutsourcing: {
		Key:         ContextOutsourcing,
		Name:        "outsourcing",
		Subject1:    "vendor_company",
		Subject2:    "client_company",
		Description: "Contract relationship: business complementarity, payment risk, scope creep",
	},
	ContextB2B: {
		Key:         ContextB2B,
		Name:        "b2b",
		Subject1:    "company_a",
		Subject2:    "company_b",
		Description: "Strategic alliance: long-term partnership, mutual benefit, power balance",
	},
	ContextHR: {
		Key:         ContextHR,
		Name:        "hr_talent",
		Subject1:    "candidate",
		Subject2:    "company",
		Description: "Talent selection: fit level, team complementarity, retention rate",
	},
	ContextHeadhunter: {
		Key:         ContextHeadhunter,
		Name:        "headhunter",
		Subject1:    "candidate",
		Subject2:    "target_company",
		Description: "Third-party matching: match rate, pitch angles, risk disclosure",
	},
}

// ValidContext checks if a context type is supported.
func ValidContext(ctx ContextType) bool {
	_, ok := Contexts[ctx]
	return ok
}

// DefaultContext returns the default context if none specified.
func DefaultContext(ctx ContextType) ContextType {
	if ctx == "" {
		return ContextEmployment
	}
	return ctx
}
