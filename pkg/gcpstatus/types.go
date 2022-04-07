package gcpstatus

import "time"

type GCPSserviceStatus []struct {
	ID           string    `json:"id"`            // incident id
	Number       string    `json:"number"`        // incident number
	Begin        time.Time `json:"begin"`         // incident begin time
	Created      time.Time `json:"created"`       // incident creation time
	End          time.Time `json:"end"`           // incident end time
	Modified     time.Time `json:"modified"`      // incident modified time
	ExternalDesc string    `json:"external_desc"` // external description for incident
	Updates      []struct {
		Created  time.Time `json:"created"`  // incident creation time
		Modified time.Time `json:"modified"` // incident modified time
		When     time.Time `json:"when"`     // incident update time
		Text     string    `json:"text"`     // incident description
		Status   string    `json:"status"`   // incident status
	} `json:"updates"`
	MostRecentUpdate struct {
		Created  time.Time `json:"created"`  // incident creation time
		Modified time.Time `json:"modified"` // incident modified time
		When     time.Time `json:"when"`     // incident update time
		Text     string    `json:"text"`     // incident description
		Status   string    `json:"status"`   // incident status
	} `json:"most_recent_update"`
	StatusImpact     string `json:"status_impact"` // incident status impact on GCP services
	Severity         string `json:"severity"`      // incident severity
	ServiceKey       string `json:"service_key"`   // incident service key
	ServiceName      string `json:"service_name"`  // GCP service name
	AffectedProducts []struct {
		Title string `json:"title"` // GCP affected product's title
		ID    string `json:"id"`    // GCP affected product's ID
	} `json:"affected_products"`
	URI string `json:"uri"` // incident URI
}
