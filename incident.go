package pagerduty

import (
	"encoding/json"
	"fmt"

	"github.com/google/go-querystring/query"
)

// Acknowledgement is the data structure of an acknoledgement of an incident.
type Acknowledgement struct {
	At           string    `json:"at,omitempty"`
	Acknowledger APIObject `json:"acknowledger,omitempty"`
}

// PendingAction is the data structure for any pending actions on an incident.
type PendingAction struct {
	Type string `json:"type,omitempty"`
	At   string `json:"at,omitempty"`
}

// Assignment is the data structure for an assignment of an incident
type Assignment struct {
	At       string    `json:"at,omitempty"`
	Assignee APIObject `json:"assignee,omitempty"`
}

// Incident is a normalized, de-duplicated event generated by a PagerDuty integration.
type Incident struct {
	APIObject
	IncidentNumber       uint              `json:"incident_number,omitempty"`
	CreatedAt            string            `json:"created_at,omitempty"`
	PendingActions       []PendingAction   `json:"pending_actions,omitempty"`
	IncidentKey          string            `json:"incident_key,omitempty"`
	Service              *APIObject        `json:"service,omitempty"`
	Assignments          []Assignment      `json:"assignments,omitempty"`
	Acknowledgements     []Acknowledgement `json:"acknowledgements,omitempty"`
	LastStatusChangeAt   string            `json:"last_status_change_at,omitempty"`
	LastStatusChangeBy   *APIObject        `json:"last_status_change_by,omitempty"`
	FirstTriggerLogEntry *APIObject        `json:"first_trigger_log_entry,omitempty"`
	EscalationPolicy     *APIObject        `json:"escalation_policy,omitempty"`
	Teams                []APIObject       `json:"teams,omitempty"`
	Urgency              string            `json:"urgency,omitempty"`
	Status               string            `json:"status,omitempty"`
	Id                   string            `json:"id,omitempty"`
	Priority             *APIObject        `json:"priority,omitempty"`
}

// ListIncidentsResponse is the response structure when calling the ListIncident API endpoint.
type ListIncidentsResponse struct {
	APIListObject
	Incidents []Incident `json:"incidents,omitempty"`
}

// ListIncidentsOptions is the structure used when passing parameters to the ListIncident API endpoint.
type ListIncidentsOptions struct {
	APIListObject
	Since       string   `url:"since,omitempty"`
	Until       string   `url:"until,omitempty"`
	DateRange   string   `url:"date_range,omitempty"`
	Statuses    []string `url:"statuses,omitempty,brackets"`
	IncidentKey string   `url:"incident_key,omitempty"`
	ServiceIDs  []string `url:"service_ids,omitempty,brackets"`
	TeamIDs     []string `url:"team_ids,omitempty,brackets"`
	UserIDs     []string `url:"user_ids,omitempty,brackets"`
	Urgencies   []string `url:"urgencies,omitempty,brackets"`
	TimeZone    string   `url:"time_zone,omitempty"`
	SortBy      string   `url:"sort_by,omitempty"`
	Includes    []string `url:"include,omitempty,brackets"`
}

// ListIncidents lists existing incidents.
func (c *Client) ListIncidents(o ListIncidentsOptions) (*ListIncidentsResponse, error) {
	v, err := query.Values(o)
	if err != nil {
		return nil, err
	}
	resp, err := c.get("/incidents?" + v.Encode())
	if err != nil {
		return nil, err
	}
	var result ListIncidentsResponse
	return &result, c.decodeJSON(resp, &result)
}

// CreateIncident is the structure POST'd to the incidents endpoint. It wraps a CreateIncidentValue
type CreateIncident struct {
	Incident CreateIncidentOptions `json:"incident"`
}

// createIncidentResponse is returned from the API when creating a response.
type createIncidentResponse struct {
	Incident Incident `json:incident`
}

// CreateIncidentOptions is the structure used when POSTing to the CreateIncident API endpoint.
type CreateIncidentOptions struct {
	Type             string       `json:"type"`
	Title            string       `json:"title"`
	Service          APIReference `json:"service"`
	Priority         APIReference `json:"priority"`
	IncidentKey      string       `json:"incident_key"`
	Body             APIDetails   `json:"body"`
	EscalationPolicy APIReference `json:"escalation_policy"`
}

// CreateIncident creates an incident synchronously without a corresponding event from a monitoring service.
func (c *Client) CreateIncident(from string, i *CreateIncident) (*Incident, error) {
	headers := make(map[string]string)
	headers["From"] = from
	resp, e := c.post("/incidents", i, &headers)
	if e != nil {
		return nil, e
	}

	var ii createIncidentResponse
	e = json.NewDecoder(resp.Body).Decode(&ii)
	if e != nil {
		return nil, e
	}

	return &ii.Incident, nil
}

// ManageIncidents acknowledges, resolves, escalates, or reassigns one or more incidents.
func (c *Client) ManageIncidents(from string, incidents []Incident) error {
	r := make(map[string][]Incident)
	headers := make(map[string]string)
	headers["From"] = from
	r["incidents"] = incidents
	_, e := c.put("/incidents", r, &headers)
	return e
}

// GetIncident shows detailed information about an incident.
func (c *Client) GetIncident(id string) (*Incident, error) {
	resp, err := c.get("/incidents/" + id)
	if err != nil {
		return nil, err
	}
	var result map[string]Incident
	if err := c.decodeJSON(resp, &result); err != nil {
		return nil, err
	}
	i, ok := result["incident"]
	if !ok {
		return nil, fmt.Errorf("JSON response does not have incident field")
	}
	return &i, nil
}

// IncidentNote is a note for the specified incident.
type IncidentNote struct {
	ID        string    `json:"id,omitempty"`
	User      APIObject `json:"user,omitempty"`
	Content   string    `json:"content,omitempty"`
	CreatedAt string    `json:"created_at,omitempty"`
}

// ListIncidentNotes lists existing notes for the specified incident.
func (c *Client) ListIncidentNotes(id string) ([]IncidentNote, error) {
	resp, err := c.get("/incidents/" + id + "/notes")
	if err != nil {
		return nil, err
	}
	var result map[string][]IncidentNote
	if err := c.decodeJSON(resp, &result); err != nil {
		return nil, err
	}
	notes, ok := result["notes"]
	if !ok {
		return nil, fmt.Errorf("JSON response does not have notes field")
	}
	return notes, nil
}

// CreateIncidentNote creates a new note for the specified incident.
func (c *Client) CreateIncidentNote(id, from string, note IncidentNote) error {
	data := make(map[string]IncidentNote)
	data["note"] = note
	// Creating a note requires the `From` header.
	// See https://api-reference.pagerduty.com/#!/Incidents/post_incidents_id_notes .
	headers := make(map[string]string)
	headers["From"] = from
	_, err := c.post("/incidents/"+id+"/notes", data, &headers)
	return err
}

// SnoozeIncident sets an incident to not alert for a specified period of time.
func (c *Client) SnoozeIncident(id string, duration uint) error {
	data := make(map[string]uint)
	data["duration"] = duration
	_, err := c.post("/incidents/"+id+"/snooze", data, nil)
	return err
}

// ListIncidentLogEntriesResponse is the response structure when calling the ListIncidentLogEntries API endpoint.
type ListIncidentLogEntriesResponse struct {
	APIListObject
	LogEntries []LogEntry `json:"log_entries,omitempty"`
}

// ListIncidentLogEntriesOptions is the structure used when passing parameters to the ListIncidentLogEntries API endpoint.
type ListIncidentLogEntriesOptions struct {
	APIListObject
	Includes   []string `url:"include,omitempty,brackets"`
	IsOverview bool     `url:"is_overview,omitempty"`
	TimeZone   string   `url:"time_zone,omitempty"`
}

// ListIncidentLogEntries lists existing log entries for the specified incident.
func (c *Client) ListIncidentLogEntries(id string, o ListIncidentLogEntriesOptions) (*ListIncidentLogEntriesResponse, error) {
	v, err := query.Values(o)
	if err != nil {
		return nil, err
	}
	resp, err := c.get("/incidents/" + id + "/log_entries?" + v.Encode())
	if err != nil {
		return nil, err
	}
	var result ListIncidentLogEntriesResponse
	return &result, c.decodeJSON(resp, &result)
}
