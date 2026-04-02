package career

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/seikaikyo/dashai-go/internal/shukuyo/engine"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

// --- 104 Job Search ---

// Job104Result is the response for 104 job search.
type Job104Result struct {
	CompanyURL string    `json:"company_url"`
	Jobs       []Job104  `json:"jobs"`
}

// Job104 is a single job from 104.com.tw.
type Job104 struct {
	Title   string `json:"title"`
	Company string `json:"company"`
	Location string `json:"location"`
	URL     string `json:"url"`
}

// Search104Jobs searches 104.com.tw for company jobs.
func Search104Jobs(companyName, birthDate, lang string) (*Job104Result, error) {
	searchURL := "https://www.104.com.tw/jobs/search/api/jobs"
	params := url.Values{
		"keyword": {companyName},
		"page":    {"1"},
	}

	req, err := http.NewRequest("GET", searchURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.104.com.tw/jobs/search/?keyword="+url.QueryEscape(companyName))
	req.Header.Set("Accept", "application/json, text/plain, */*")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("104 search failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Data []struct {
			CustName       string `json:"custName"`
			JobName        string `json:"jobName"`
			JobNameSnippet string `json:"jobNameSnippet"`
			JobAddrNoDesc  string `json:"jobAddrNoDesc"`
			JobAddress     string `json:"jobAddress"`
			Link           struct {
				Job  string `json:"job"`
				Cust string `json:"cust"`
			} `json:"link"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("104 parse failed: %w", err)
	}

	var jobs []Job104
	companyURL := ""
	for _, d := range raw.Data {
		if !strings.Contains(d.CustName, companyName) && !strings.Contains(companyName, d.CustName) {
			continue
		}
		loc := d.JobAddrNoDesc
		if loc == "" {
			loc = d.JobAddress
		}
		title := d.JobName
		if title == "" {
			title = d.JobNameSnippet
		}
		jobURL := ""
		if d.Link.Job != "" {
			jobURL = "https:" + d.Link.Job
		}
		if companyURL == "" && d.Link.Cust != "" {
			companyURL = "https:" + d.Link.Cust
		}
		jobs = append(jobs, Job104{
			Title:    title,
			Company:  d.CustName,
			Location: loc,
			URL:      jobURL,
		})
	}

	return &Job104Result{
		CompanyURL: companyURL,
		Jobs:       jobs,
	}, nil
}

// --- Global Company Search ---

// GlobalSearchResult is the response for global company search.
type GlobalSearchResult struct {
	Name         string `json:"name"`
	FoundingDate string `json:"founding_date"`
	Country      string `json:"country"`
	CountryName  string `json:"country_name"`
	Source       string `json:"source"`
	Relation     string `json:"relation_type,omitempty"`
	Direction    string `json:"direction,omitempty"`
	Level        string `json:"level,omitempty"`
}

// SearchGlobal searches for a company by country and returns founding date + compatibility.
func SearchGlobal(companyName, country, birthDate string) (*GlobalSearchResult, error) {
	var foundingDate, source string
	var err error

	switch country {
	case "tw":
		foundingDate, source, err = searchGCIS(companyName)
	case "jp":
		foundingDate, source, err = searchGBizInfo(companyName)
	case "us":
		foundingDate, source, err = searchOpenCorporates(companyName)
	default:
		return nil, fmt.Errorf("unsupported country: %s (supported: tw, jp, us)", country)
	}

	if err != nil {
		return nil, err
	}

	countryNames := map[string]string{"tw": "Taiwan", "jp": "Japan", "us": "United States"}

	result := &GlobalSearchResult{
		Name:         companyName,
		FoundingDate: foundingDate,
		Country:      country,
		CountryName:  countryNames[country],
		Source:       source,
	}

	// Add compatibility if birth date provided
	if birthDate != "" && foundingDate != "" {
		bd, err1 := engine.ParseDate(birthDate)
		cd, err2 := engine.ParseDate(foundingDate)
		if err1 == nil && err2 == nil {
			compat := engine.Compatibility(bd, cd)
			result.Relation = compat.Relation.Group
			result.Direction = compat.Relation.Direction
			result.Level = CareerLevel(compat.Relation.Group)
		}
	}

	return result, nil
}

// searchGCIS searches Taiwan GCIS (via proxy to avoid IP blocking).
func searchGCIS(companyName string) (string, string, error) {
	// GCIS API through proxy (direct calls blocked from US IPs)
	apiID := "6BBA2268-1367-4B42-9CCA-BC17499EBE8C"
	proxyBase := os.Getenv("GCIS_PROXY_URL")
	if proxyBase == "" {
		proxyBase = "https://shukuyo.dashai.dev/proxy/gcis"
	}

	params := url.Values{
		"$format": {"json"},
		"$filter": {fmt.Sprintf("Company_Name like %s and Company_Status eq 01", companyName)},
		"$top":    {"5"},
	}

	reqURL := fmt.Sprintf("%s/%s?%s", proxyBase, apiID, params.Encode())
	resp, err := httpClient.Get(reqURL)
	if err != nil {
		return "", "", fmt.Errorf("GCIS search failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var results []struct {
		CompanyName   string `json:"Company_Name"`
		SetupDate     string `json:"Company_Setup_Date"`
	}
	if err := json.Unmarshal(body, &results); err != nil {
		return "", "", fmt.Errorf("GCIS parse failed: %w", err)
	}

	for _, r := range results {
		if strings.Contains(r.CompanyName, companyName) || strings.Contains(companyName, r.CompanyName) {
			if len(r.SetupDate) == 7 {
				// ROC date: 1100101 → 2021-01-01
				rocYear := 0
				fmt.Sscanf(r.SetupDate[:3], "%d", &rocYear)
				mm := r.SetupDate[3:5]
				dd := r.SetupDate[5:7]
				return fmt.Sprintf("%d-%s-%s", rocYear+1911, mm, dd), "gcis", nil
			}
		}
	}

	return "", "", fmt.Errorf("company not found in GCIS: %s", companyName)
}

// searchGBizInfo searches Japan gBizINFO.
func searchGBizInfo(companyName string) (string, string, error) {
	token := os.Getenv("GBIZINFO_API_TOKEN")
	if token == "" {
		return "", "", fmt.Errorf("GBIZINFO_API_TOKEN not configured")
	}

	params := url.Values{
		"name": {companyName},
		"page": {"1"},
	}

	req, err := http.NewRequest("GET", "https://api.info.gbiz.go.jp/hojin/v2/hojin?"+params.Encode(), nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("X-hojinInfo-api-token", token)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("gBizINFO search failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var raw struct {
		HojinInfos []struct {
			Name                string `json:"name"`
			DateOfEstablishment string `json:"date_of_establishment"`
			FoundingYear        string `json:"founding_year"`
			Status              string `json:"status"`
		} `json:"hojin-infos"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", "", fmt.Errorf("gBizINFO parse failed: %w", err)
	}

	for _, h := range raw.HojinInfos {
		if h.Status == "閉鎖" {
			continue
		}
		date := h.DateOfEstablishment
		if date != "" {
			// Format: YYYY/MM/DD or YYYY-MM-DDTHH:MM:SS
			date = strings.ReplaceAll(date, "/", "-")
			if idx := strings.Index(date, "T"); idx > 0 {
				date = date[:idx]
			}
			return date, "gbizinfo", nil
		}
		if h.FoundingYear != "" {
			return h.FoundingYear + "-01-01", "gbizinfo", nil
		}
	}

	return "", "", fmt.Errorf("company not found in gBizINFO: %s", companyName)
}

// searchOpenCorporates searches US companies via OpenCorporates.
func searchOpenCorporates(companyName string) (string, string, error) {
	token := os.Getenv("OPENCORPORATES_API_KEY")
	if token == "" {
		return "", "", fmt.Errorf("OPENCORPORATES_API_KEY not configured")
	}

	params := url.Values{
		"q":                 {companyName},
		"jurisdiction_code": {"us"},
		"api_token":         {token},
	}

	resp, err := httpClient.Get("https://api.opencorporates.com/v0.4.8/companies/search?" + params.Encode())
	if err != nil {
		return "", "", fmt.Errorf("OpenCorporates search failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var raw struct {
		Results struct {
			Companies []struct {
				Company struct {
					Name              string `json:"name"`
					IncorporationDate string `json:"incorporation_date"`
				} `json:"company"`
			} `json:"companies"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return "", "", fmt.Errorf("OpenCorporates parse failed: %w", err)
	}

	nameUpper := strings.ToUpper(companyName)
	for _, c := range raw.Results.Companies {
		if strings.Contains(strings.ToUpper(c.Company.Name), nameUpper) && c.Company.IncorporationDate != "" {
			return c.Company.IncorporationDate, "opencorporates", nil
		}
	}

	// Fallback: first result
	if len(raw.Results.Companies) > 0 {
		c := raw.Results.Companies[0]
		if c.Company.IncorporationDate != "" {
			return c.Company.IncorporationDate, "opencorporates", nil
		}
	}

	return "", "", fmt.Errorf("company not found in OpenCorporates: %s", companyName)
}
