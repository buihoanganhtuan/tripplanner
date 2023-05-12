package planner

/*
A domain-level package Define domain types that model
real entities involved in the business, independent
from any underlying technology.
*/

// For consistency, we should select a convention for the response when errors occur and stick with it.
// The convention chosen here is: https://google.github.io/styleguide/jsoncstyleguide.xml#Reserved_Property_Names_in_the_error_object
type ErrorResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message,omitempty"`
	Errors  []ErrorDescriptor `json:"errors,omitempty"`
}
type ErrorDescriptor struct {
	Domain       string `json:"domain,omitempty"`
	Reason       string `json:"reason,omitempty"`
	Message      string `json:"message,omitempty"`
	Location     string `json:"location,omitempty"`
	LocationType string `json:"locationType,omitempty"`
}
