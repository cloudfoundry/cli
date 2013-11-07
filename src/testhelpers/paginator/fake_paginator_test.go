package paginator_test

import (
	"testing"
	"testhelpers/paginator"
	"github.com/stretchr/testify/assert"
)

func TestFetchAllPages(t *testing.T) {
	p := &paginator.FakePaginator{
		TotalResults: []string{"a","b","c"},
	}

	output := []string{}

	results, apiReponse := p.Next()
	assert.True(t,apiReponse.IsSuccessful())

	output = append(output, results)

	assert.True(t,p.HasNext())
	assert.Equal(t,output,[]string{"a"})
	assert.True(t,p.HasNext())

	results, apiReponse = p.Next()
	assert.True(t,apiReponse.IsSuccessful())
	output = append(output, results)

	results, apiReponse = p.Next()
	assert.True(t,apiReponse.IsSuccessful())
	output = append(output, results)

	assert.Equal(t,output,[]string{"a","b","c"})
	assert.False(t,p.HasNext())
}
