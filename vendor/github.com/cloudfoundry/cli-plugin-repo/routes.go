package main

import "github.com/tedsuo/rata"

const (
	Index = "Index"
	List  = "List"

	UI = "UI"
)

var Routes = rata.Routes([]rata.Route{
	{Path: "/list", Method: "GET", Name: List},

	//Deprecated URI
	{Path: "/ui/", Method: "GET", Name: UI},

	{Path: "/", Method: "GET", Name: Index},
	{Path: "/:file/", Method: "GET", Name: Index},
})
