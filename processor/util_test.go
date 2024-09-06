package processor_test

import (
	"encoding/json"
	"testing"

	"github.com/treaster/incant/processor"

	"github.com/stretchr/testify/require"
)

func TestEvalContentExpr(t *testing.T) {
	dataJson := []byte(`{
		"key1": "value1",
		"key2": [
			 "value2",
			 "value3"
		]
	}`)

	var data any
	err := json.Unmarshal(dataJson, &data)
	require.NoError(t, err)

	results := processor.EvalContentExpr("jq:.key2[]", data)
	require.Equal(t, 2, len(results))
	require.Equal(t, "value2", results[0])
	require.Equal(t, "value3", results[1])
}
