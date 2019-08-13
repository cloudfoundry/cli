package ccv3

import (
	"net/http"
	"strconv"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	log "github.com/sirupsen/logrus"
)

const BATCH_SIZE = 10

type PaginatedResult struct {
	Page     PaginatedResources
	Warnings Warnings
	Err      error
}

func (client Client) paginate(request *cloudcontroller.Request, obj interface{}, appendToExternalList func(interface{}) error) (Warnings, error) {
	fullWarningsList := Warnings{}

	for {
		wrapper := NewPaginatedResources(obj)
		response := cloudcontroller.Response{
			DecodeJSONResponseInto: &wrapper,
		}

		err := client.connection.Make(request, &response)
		fullWarningsList = append(fullWarningsList, response.Warnings...)
		if err != nil {
			return fullWarningsList, err
		}

		list, err := wrapper.Resources()
		if err != nil {
			return fullWarningsList, err
		}

		for _, item := range list {
			err = appendToExternalList(item)
			if err != nil {
				return fullWarningsList, err
			}
		}

		if wrapper.NextPage() == "" {
			break
		}

		request, err = client.newHTTPRequest(requestOptions{
			URL:    wrapper.NextPage(),
			Method: http.MethodGet,
		})
		if err != nil {
			return fullWarningsList, err
		}
	}

	return fullWarningsList, nil
}

func (client Client) paginate2(requestOptions requestOptions, obj interface{}, appendToExternalList func(interface{}) error) (Warnings, error) {
	fullWarningsList := Warnings{}

	pageOne, warnings, err := client.fetchPage2(requestOptions, obj, 1)
	log.Debugf("PAGE 1: %v", pageOne)
	fullWarningsList = append(fullWarningsList, warnings...)
	if err != nil {
		return fullWarningsList, err
	}

	resourcesByPageNumber := make([]PaginatedResources, pageOne.Pagination.TotalPages)
	remainingPages := pageOne.Pagination.TotalPages - 1

	resourcesByPageNumber[0] = pageOne

	resultsChannel := make(chan PaginatedResult, remainingPages)

	for i := 0; i < remainingPages; i++ {
		go client.work(i+2, requestOptions, obj, resultsChannel)
	}

	for i := 0; i < remainingPages; i++ {
		log.Debugf("Processing Result %d", i)
		result := <-resultsChannel
		fullWarningsList = append(fullWarningsList, result.Warnings...)
		if result.Err != nil {
			return fullWarningsList, result.Err
		}
		resourcesByPageNumber[result.Page.PageNumber-1] = result.Page
	}

	log.Debugf("Appending resources")
	for _, page := range resourcesByPageNumber {
		list, err := page.Resources()
		if err != nil {
			return fullWarningsList, err
		}

		for _, item := range list {
			err = appendToExternalList(item)
			if err != nil {
				return fullWarningsList, err
			}
		}
	}
	log.Debugf("RETURNING")
	return fullWarningsList, nil
}

func (client Client) fetchPage(request *cloudcontroller.Request, obj interface{}, pageIndex int) (PaginatedResources, Warnings, error) {
	wrapper := NewPaginatedResources(obj)

	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &wrapper,
	}

	err := client.connection.Make(request, &response)

	wrapper.PageNumber = pageIndex

	return *wrapper, response.Warnings, err
}

func (client Client) fetchPage2(requestOptions requestOptions, obj interface{}, pageIndex int) (PaginatedResources, Warnings, error) {
	requestOptions.Query = append(requestOptions.Query, Query{Key: "page", Values: []string{strconv.Itoa(pageIndex)}})
	request, err := client.newHTTPRequest(requestOptions)
	if err != nil {
		return PaginatedResources{}, Warnings{}, err
	}

	wrapper := NewPaginatedResources(obj)

	response := cloudcontroller.Response{
		DecodeJSONResponseInto: &wrapper,
	}

	err = client.connection.Make(request, &response)

	wrapper.PageNumber = pageIndex

	return *wrapper, response.Warnings, err
}

func (client Client) work(pageNum int, requestOptions requestOptions, obj interface{}, resultsChannel chan<- PaginatedResult) {
	page, warnings, err := client.fetchPage2(requestOptions, obj, pageNum)
	resultsChannel <- PaginatedResult{Page: page, Warnings: warnings, Err: err}
}
