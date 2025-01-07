package internal

import (
	"net/http"

	"github.com/tedsuo/rata"
)

const (
	CreatePolicies = "PostPolicies"
	DeletePolicies = "DeletePolicies"
	ListPolicies   = "ListPolicies"
)

// Routes is a list of routes used by the rata library to construct request
// URLs.
var Routes = rata.Routes{
	{Path: "/policies", Method: http.MethodPost, Name: CreatePolicies},
	{Path: "/policies/delete", Method: http.MethodPost, Name: DeletePolicies},
	{Path: "/policies", Method: http.MethodGet, Name: ListPolicies},
}
