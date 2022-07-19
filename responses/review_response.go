package responses

type Review_Owner_Response struct {
	Username   string  `json:"full_name,omitempty"`
	Rating     float64 `json:"rating,omitempty"`
	Trip       int     `json:"trip,omitempty"`
	User_Email string  `json:"user_email,omitempty"`
}
