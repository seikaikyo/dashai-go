package user

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/seikaikyo/dashai-go/internal/database"
)

// Store handles all user-related database operations.
type Store struct {
	db *database.DB
}

// --- User profile ---

const userCols = `id, auth_id, email, display_name, birth_date, plan, credits_remaining, preferences, hr_company, created_at, updated_at`

func scanUser(row scannable) (*User, error) {
	var u User
	var birthDate *time.Time
	var prefsJSON, hrCompJSON []byte

	err := row.Scan(
		&u.ID, &u.AuthID, &u.Email, &u.DisplayName,
		&birthDate, &u.Plan, &u.CreditsRemaining,
		&prefsJSON, &hrCompJSON,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	u.BirthDate = datePtr(birthDate)
	u.Preferences = unmarshalJSON(prefsJSON, map[string]any{})
	u.HrCompany = unmarshalJSON(hrCompJSON, map[string]any(nil))
	return &u, nil
}

func (s *Store) GetOrCreateUser(ctx context.Context, authID string) (*User, error) {
	row := s.db.Pool.QueryRow(ctx,
		`SELECT `+userCols+` FROM users WHERE auth_id = $1`, authID)
	u, err := scanUser(row)
	if err == pgx.ErrNoRows {
		now := time.Now().UTC()
		id := newUUID()
		row = s.db.Pool.QueryRow(ctx,
			`INSERT INTO users (id, auth_id, plan, credits_remaining, preferences, created_at, updated_at)
			 VALUES ($1, $2, 'free', 0, '{}', $3, $3)
			 RETURNING `+userCols,
			id, authID, now)
		return scanUser(row)
	}
	return u, err
}

func (s *Store) UpdateUserSync(ctx context.Context, authID string, req ProfileSyncRequest) (*User, error) {
	u, err := s.GetOrCreateUser(ctx, authID)
	if err != nil {
		return nil, err
	}

	// Shallow merge preferences
	if req.Preferences != nil {
		for k, v := range req.Preferences {
			u.Preferences[k] = v
		}
	}
	prefsJSON, _ := json.Marshal(u.Preferences)

	now := time.Now().UTC()
	row := s.db.Pool.QueryRow(ctx,
		`UPDATE users SET
			birth_date = COALESCE($2, birth_date),
			display_name = COALESCE($3, display_name),
			preferences = $4,
			updated_at = $5
		 WHERE auth_id = $1
		 RETURNING `+userCols,
		authID, parseDateNullable(req.BirthDate), req.DisplayName, prefsJSON, now)
	return scanUser(row)
}

func (s *Store) GetFullProfile(ctx context.Context, authID string) (*User, []Partner, []Company, []JobSeeker, []HrCandidate, error) {
	u, err := s.GetOrCreateUser(ctx, authID)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	partners, err := s.listPartners(ctx, u.ID)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	companies, err := s.listCompanies(ctx, u.ID)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	jobSeekers, err := s.listJobSeekers(ctx, u.ID)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	hrCandidates, err := s.listHrCandidates(ctx, u.ID)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	return u, partners, companies, jobSeekers, hrCandidates, nil
}

// --- SyncFull (transaction) ---

func (s *Store) SyncFull(ctx context.Context, authID string, req FullSyncRequest) error {
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get or create user
	var userID string
	err = tx.QueryRow(ctx, `SELECT id FROM users WHERE auth_id = $1`, authID).Scan(&userID)
	if err == pgx.ErrNoRows {
		userID = newUUID()
		now := time.Now().UTC()
		_, err = tx.Exec(ctx,
			`INSERT INTO users (id, auth_id, plan, credits_remaining, preferences, created_at, updated_at)
			 VALUES ($1, $2, 'free', 0, '{}', $3, $3)`, userID, authID, now)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// Update user fields
	prefsJSON, _ := json.Marshal(req.Preferences)
	hrJSON, _ := json.Marshal(req.HrCompany)
	now := time.Now().UTC()
	_, err = tx.Exec(ctx,
		`UPDATE users SET birth_date = $2, preferences = $3, hr_company = $4, updated_at = $5
		 WHERE id = $1`,
		userID, parseDateNullable(req.BirthDate), prefsJSON, hrJSON, now)
	if err != nil {
		return err
	}

	// Sync partners
	if err := syncEntities(ctx, tx, userID, "user_partners", req.Partners,
		func(tx pgx.Tx, uid string, p PartnerData) error {
			tagsJSON, _ := json.Marshal(p.SubTags)
			id := p.ID
			if id == "" {
				id = newUUID()
			}
			_, err := tx.Exec(ctx,
				`INSERT INTO user_partners (id, user_id, nickname, birth_date, relation, sub_tags, emotion_direction, created_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				 ON CONFLICT (id) DO UPDATE SET
				   nickname = EXCLUDED.nickname, birth_date = EXCLUDED.birth_date,
				   relation = EXCLUDED.relation, sub_tags = EXCLUDED.sub_tags,
				   emotion_direction = EXCLUDED.emotion_direction`,
				id, uid, p.Nickname, parseDateString(p.BirthDate), p.Relation, tagsJSON, p.EmotionDirection, now)
			return err
		}); err != nil {
		return err
	}

	// Sync companies
	if err := syncEntities(ctx, tx, userID, "user_companies", req.Companies,
		func(tx pgx.Tx, uid string, c CompanyData) error {
			id := c.ID
			if id == "" {
				id = newUUID()
			}
			_, err := tx.Exec(ctx,
				`INSERT INTO user_companies (id, user_id, name, founding_date, country, memo, job_url, created_at)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				 ON CONFLICT (id) DO UPDATE SET
				   name = EXCLUDED.name, founding_date = EXCLUDED.founding_date,
				   country = EXCLUDED.country, memo = EXCLUDED.memo, job_url = EXCLUDED.job_url`,
				id, uid, c.Name, parseDateNullable(c.FoundingDate), c.Country, c.Memo, c.JobURL, now)
			return err
		}); err != nil {
		return err
	}

	// Sync job seekers
	if err := syncEntities(ctx, tx, userID, "user_job_seekers", req.JobSeekers,
		func(tx pgx.Tx, uid string, js JobSeekerData) error {
			id := js.ID
			if id == "" {
				id = newUUID()
			}
			companiesJSON, _ := json.Marshal(js.Companies)
			_, err := tx.Exec(ctx,
				`INSERT INTO user_job_seekers (id, user_id, name, birth_date, companies, created_at)
				 VALUES ($1, $2, $3, $4, $5, $6)
				 ON CONFLICT (id) DO UPDATE SET
				   name = EXCLUDED.name, birth_date = EXCLUDED.birth_date, companies = EXCLUDED.companies`,
				id, uid, js.Name, parseDateString(js.BirthDate), companiesJSON, now)
			return err
		}); err != nil {
		return err
	}

	// Sync HR candidates
	if err := syncEntities(ctx, tx, userID, "user_hr_candidates", req.HrCandidates,
		func(tx pgx.Tx, uid string, hc HrCandidateData) error {
			id := hc.ID
			if id == "" {
				id = newUUID()
			}
			_, err := tx.Exec(ctx,
				`INSERT INTO user_hr_candidates (id, user_id, name, birth_date, created_at)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (id) DO UPDATE SET
				   name = EXCLUDED.name, birth_date = EXCLUDED.birth_date`,
				id, uid, hc.Name, parseDateString(hc.BirthDate), now)
			return err
		}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// syncEntities is a generic helper for full-replacement sync within a transaction.
// It deletes items not in the incoming list, then upserts each incoming item.
type idGetter interface {
	getID() string
}

func (p PartnerData) getID() string     { return p.ID }
func (c CompanyData) getID() string     { return c.ID }
func (j JobSeekerData) getID() string   { return j.ID }
func (h HrCandidateData) getID() string { return h.ID }

func syncEntities[T idGetter](ctx context.Context, tx pgx.Tx, userID, table string, items []T, upsert func(pgx.Tx, string, T) error) error {
	// Collect incoming IDs (skip empty = new items)
	incomingIDs := make([]string, 0, len(items))
	for _, item := range items {
		if id := item.getID(); id != "" {
			incomingIDs = append(incomingIDs, id)
		}
	}

	// Delete items not in the incoming list
	if len(incomingIDs) > 0 {
		_, err := tx.Exec(ctx,
			`DELETE FROM `+table+` WHERE user_id = $1 AND id != ALL($2)`,
			userID, incomingIDs)
		if err != nil {
			return err
		}
	} else {
		// No incoming items: delete all
		_, err := tx.Exec(ctx,
			`DELETE FROM `+table+` WHERE user_id = $1`, userID)
		if err != nil {
			return err
		}
	}

	// Upsert each item
	for _, item := range items {
		if err := upsert(tx, userID, item); err != nil {
			return err
		}
	}
	return nil
}

// --- List helpers (for GetFullProfile) ---

func (s *Store) listPartners(ctx context.Context, userID string) ([]Partner, error) {
	rows, err := s.db.Pool.Query(ctx,
		`SELECT id, user_id, nickname, birth_date, relation, sub_tags, emotion_direction, created_at
		 FROM user_partners WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Partner
	for rows.Next() {
		var p Partner
		var bd time.Time
		var tagsJSON []byte
		if err := rows.Scan(&p.ID, &p.UserID, &p.Nickname, &bd, &p.Relation, &tagsJSON, &p.EmotionDirection, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.BirthDate = bd.Format("2006-01-02")
		p.SubTags = unmarshalJSON(tagsJSON, []string{})
		result = append(result, p)
	}
	return result, rows.Err()
}

func (s *Store) listCompanies(ctx context.Context, userID string) ([]Company, error) {
	rows, err := s.db.Pool.Query(ctx,
		`SELECT id, user_id, name, founding_date, country, memo, job_url, created_at
		 FROM user_companies WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Company
	for rows.Next() {
		var c Company
		var fd *time.Time
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &fd, &c.Country, &c.Memo, &c.JobURL, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.FoundingDate = datePtr(fd)
		result = append(result, c)
	}
	return result, rows.Err()
}

func (s *Store) listJobSeekers(ctx context.Context, userID string) ([]JobSeeker, error) {
	rows, err := s.db.Pool.Query(ctx,
		`SELECT id, user_id, name, birth_date, companies, created_at
		 FROM user_job_seekers WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []JobSeeker
	for rows.Next() {
		var js JobSeeker
		var bd time.Time
		var companiesJSON []byte
		if err := rows.Scan(&js.ID, &js.UserID, &js.Name, &bd, &companiesJSON, &js.CreatedAt); err != nil {
			return nil, err
		}
		js.BirthDate = bd.Format("2006-01-02")
		js.Companies = unmarshalJSON(companiesJSON, []map[string]any{})
		result = append(result, js)
	}
	return result, rows.Err()
}

func (s *Store) listHrCandidates(ctx context.Context, userID string) ([]HrCandidate, error) {
	rows, err := s.db.Pool.Query(ctx,
		`SELECT id, user_id, name, birth_date, created_at
		 FROM user_hr_candidates WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []HrCandidate
	for rows.Next() {
		var hc HrCandidate
		var bd time.Time
		if err := rows.Scan(&hc.ID, &hc.UserID, &hc.Name, &bd, &hc.CreatedAt); err != nil {
			return nil, err
		}
		hc.BirthDate = bd.Format("2006-01-02")
		result = append(result, hc)
	}
	return result, rows.Err()
}

// --- Company cache ---

func (s *Store) GetCompanyCache(ctx context.Context, country, name string) (*CompanyCacheEntry, error) {
	var c CompanyCacheEntry
	var fd *time.Time
	err := s.db.Pool.QueryRow(ctx,
		`SELECT id, name, country, founding_date, business_no, source, job_url_104, created_at, updated_at
		 FROM company_cache WHERE country = $1 AND name = $2`, country, name).
		Scan(&c.ID, &c.Name, &c.Country, &fd, &c.BusinessNo, &c.Source, &c.JobURL104, &c.CreatedAt, &c.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.FoundingDate = datePtr(fd)
	return &c, nil
}

func (s *Store) BatchGetCompanyCache(ctx context.Context, country string, names []string) (map[string]*CompanyCacheEntry, error) {
	rows, err := s.db.Pool.Query(ctx,
		`SELECT id, name, country, founding_date, business_no, source, job_url_104, created_at, updated_at
		 FROM company_cache WHERE country = $1 AND name = ANY($2)`, country, names)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]*CompanyCacheEntry, len(names))
	for rows.Next() {
		var c CompanyCacheEntry
		var fd *time.Time
		if err := rows.Scan(&c.ID, &c.Name, &c.Country, &fd, &c.BusinessNo, &c.Source, &c.JobURL104, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		c.FoundingDate = datePtr(fd)
		result[c.Name] = &c
	}
	return result, rows.Err()
}

func (s *Store) UpsertCompanyCache(ctx context.Context, req CompanyCacheSaveRequest) error {
	now := time.Now().UTC()
	existing, err := s.GetCompanyCache(ctx, req.Country, req.Name)
	if err != nil {
		return err
	}

	if existing != nil {
		_, err = s.db.Pool.Exec(ctx,
			`UPDATE company_cache SET
				founding_date = COALESCE($3, founding_date),
				business_no = COALESCE($4, business_no),
				source = COALESCE($5, source),
				job_url_104 = COALESCE($6, job_url_104),
				updated_at = $7
			 WHERE country = $1 AND name = $2`,
			req.Country, req.Name,
			parseDateNullable(req.FoundingDate), req.BusinessNo, req.Source, req.JobURL104, now)
	} else {
		id := newUUID()
		_, err = s.db.Pool.Exec(ctx,
			`INSERT INTO company_cache (id, name, country, founding_date, business_no, source, job_url_104, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)`,
			id, req.Name, req.Country,
			parseDateNullable(req.FoundingDate), req.BusinessNo, req.Source, req.JobURL104, now)
	}
	return err
}

// --- Date parsing helpers ---

func parseDateString(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}

func parseDateNullable(s *string) *time.Time {
	if s == nil {
		return nil
	}
	return parseDateString(*s)
}
