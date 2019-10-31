package marshaler

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

type PromqlMarshaler struct {
	fallback runtime.Marshaler
}

func NewPromqlMarshaler(fallback runtime.Marshaler) *PromqlMarshaler {
	return &PromqlMarshaler{
		fallback: fallback,
	}
}

func (m *PromqlMarshaler) Marshal(v interface{}) ([]byte, error) {
	switch q := v.(type) {
	case *logcache_v1.PromQL_InstantQueryResult:
		result, err := m.assembleInstantQueryResult(q)
		if err != nil {
			return nil, err
		}

		return appendNewLine(json.Marshal(result))
	case *logcache_v1.PromQL_RangeQueryResult:
		result, err := m.assembleRangeQueryResult(q)
		if err != nil {
			return nil, err
		}

		return appendNewLine(json.Marshal(result))
	default:
		return appendNewLine(m.fallback.Marshal(v))
	}
}

func appendNewLine(bytes []byte, err error) ([]byte, error) {
	return append(bytes, byte('\n')), err
}

type queryResult struct {
	Status    string     `json:"status"`
	Data      resultData `json:"data"`
	ErrorType string     `json:"errorType,omitempty"`
	Error     string     `json:"error,omitempty"`
}

type resultData struct {
	ResultType string          `json:"resultType"`
	Result     json.RawMessage `json:"result,omitempty"`
}

type sample struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value"`
}

type series struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"`
}

func (m *PromqlMarshaler) assembleInstantQueryResult(v *logcache_v1.PromQL_InstantQueryResult) (*queryResult, error) {
	var data resultData
	var err error

	switch v.GetResult().(type) {
	case *logcache_v1.PromQL_InstantQueryResult_Scalar:
		data, err = assembleScalarResultData(v.GetScalar())
	case *logcache_v1.PromQL_InstantQueryResult_Vector:
		data, err = assembleVectorResultData(v.GetVector())
	case *logcache_v1.PromQL_InstantQueryResult_Matrix:
		data, err = assembleMatrixResultData(v.GetMatrix())
	}

	if err != nil {
		return nil, err
	}

	return &queryResult{
		Status: "success",
		Data:   data,
	}, nil
}

func (m *PromqlMarshaler) assembleRangeQueryResult(v *logcache_v1.PromQL_RangeQueryResult) (*queryResult, error) {
	var data resultData
	var err error

	switch v.GetResult().(type) {
	case *logcache_v1.PromQL_RangeQueryResult_Matrix:
		data, err = assembleMatrixResultData(v.GetMatrix())
	}

	if err != nil {
		return nil, err
	}

	return &queryResult{
		Status: "success",
		Data:   data,
	}, nil
}

func assembleScalarResultData(v *logcache_v1.PromQL_Scalar) (resultData, error) {
	point, err := assemblePoint(v.GetTime(), v.GetValue())
	if err != nil {
		return resultData{}, err
	}

	data, err := json.Marshal(point)
	if err != nil {
		return resultData{}, err
	}

	return resultData{
		ResultType: "scalar",
		Result:     data,
	}, nil
}

func assembleVectorResultData(v *logcache_v1.PromQL_Vector) (resultData, error) {
	// NOTE: This is required to make sure that JSON marshals an empty result
	// set as `[]` and not `null`.
	samples := make([]interface{}, 0)

	for _, s := range v.GetSamples() {
		p := s.GetPoint()
		point, err := assemblePoint(p.GetTime(), p.GetValue())
		if err != nil {
			return resultData{}, err
		}

		metric := s.GetMetric()
		if metric == nil {
			metric = make(map[string]string, 0)
		}

		samples = append(samples, sample{
			Metric: metric,
			Value:  point,
		})
	}

	data, err := json.Marshal(samples)
	if err != nil {
		return resultData{}, err
	}

	return resultData{
		ResultType: "vector",
		Result:     data,
	}, nil
}

func assembleMatrixResultData(v *logcache_v1.PromQL_Matrix) (resultData, error) {
	// NOTE: This is required to make sure that JSON marshals an empty result
	// set as `[]` and not `null`.
	result := make([]interface{}, 0)

	for _, s := range v.GetSeries() {
		var values [][]interface{}
		for _, p := range s.GetPoints() {
			point, err := assemblePoint(p.GetTime(), p.GetValue())
			if err != nil {
				return resultData{}, err
			}

			values = append(values, point)
		}

		metric := s.GetMetric()
		if metric == nil {
			metric = make(map[string]string, 0)
		}

		result = append(result, series{
			Metric: metric,
			Values: values,
		})
	}

	data, err := json.Marshal(result)
	if err != nil {
		return resultData{}, err
	}

	return resultData{
		ResultType: "matrix",
		Result:     data,
	}, nil
}

func assemblePoint(time string, value float64) ([]interface{}, error) {
	t, err := strconv.ParseFloat(time, 64)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse float %s: %s", time, err.Error())
	}
	formattedTime := fmt.Sprintf("%.3f", t)
	return []interface{}{
		json.RawMessage(formattedTime),
		strconv.FormatFloat(value, 'f', -1, 64),
	}, nil
}

func (m *PromqlMarshaler) NewEncoder(w io.Writer) runtime.Encoder {
	fallbackEncoder := m.fallback.NewEncoder(w)
	jsonEncoder := json.NewEncoder(w)

	return runtime.EncoderFunc(func(v interface{}) error {
		switch q := v.(type) {
		case *logcache_v1.PromQL_InstantQueryResult:
			result, err := m.assembleInstantQueryResult(q)
			if err != nil {
				return err
			}

			return jsonEncoder.Encode(result)
		case *logcache_v1.PromQL_RangeQueryResult:
			result, err := m.assembleRangeQueryResult(q)
			if err != nil {
				return err
			}

			return jsonEncoder.Encode(result)
		default:
			return fallbackEncoder.Encode(v)
		}
	})
}

// The special marshaling for PromQL results is currently only implemented
// for encoding.
func (m *PromqlMarshaler) Unmarshal(data []byte, v interface{}) error {
	var result queryResult

	switch q := v.(type) {
	case *logcache_v1.PromQL_InstantQueryResult:
		err := json.Unmarshal(data, &result)
		if err != nil {
			return err
		}

		r, err := m.disassembleInstantQueryResult(result)
		if err != nil {
			return err
		}
		*q = *r
	case *logcache_v1.PromQL_RangeQueryResult:
		err := json.Unmarshal(data, &result)
		if err != nil {
			return err
		}

		r, err := m.disassembleRangeQueryResult(result)
		if err != nil {
			return err
		}
		*q = *r
	default:
		return m.fallback.Unmarshal(data, v)
	}

	return nil
}

func (m *PromqlMarshaler) disassembleInstantQueryResult(q queryResult) (*logcache_v1.PromQL_InstantQueryResult, error) {
	switch q.Data.ResultType {
	case "scalar":
		r, err := unmarshalScalarResultData(q.Data.Result)
		if err != nil {
			return nil, err
		}

		return &logcache_v1.PromQL_InstantQueryResult{
			Result: &logcache_v1.PromQL_InstantQueryResult_Scalar{
				Scalar: r,
			},
		}, nil
	case "vector":
		r, err := unmarshalVectorResultData(q.Data.Result)
		if err != nil {
			return nil, err
		}

		return &logcache_v1.PromQL_InstantQueryResult{
			Result: &logcache_v1.PromQL_InstantQueryResult_Vector{
				Vector: r,
			},
		}, nil
	case "matrix":
		r, err := unmarshalMatrixResultData(q.Data.Result)
		if err != nil {
			return nil, err
		}

		return &logcache_v1.PromQL_InstantQueryResult{
			Result: &logcache_v1.PromQL_InstantQueryResult_Matrix{
				Matrix: r,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown instant query resultType '%s'", q.Data.ResultType)
	}
}

func (m *PromqlMarshaler) disassembleRangeQueryResult(q queryResult) (*logcache_v1.PromQL_RangeQueryResult, error) {
	switch q.Data.ResultType {
	case "matrix":
		r, err := unmarshalMatrixResultData(q.Data.Result)
		if err != nil {
			return nil, err
		}

		return &logcache_v1.PromQL_RangeQueryResult{
			Result: &logcache_v1.PromQL_RangeQueryResult_Matrix{
				Matrix: r,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown range query resultType '%s'", q.Data.ResultType)
	}
}

func unmarshalScalarResultData(data []byte) (*logcache_v1.PromQL_Scalar, error) {
	var point []interface{}
	err := json.Unmarshal(data, &point)
	if err != nil {
		return nil, err
	}

	time, value, err := disassemblePoint(point)
	if err != nil {
		return nil, err
	}

	return &logcache_v1.PromQL_Scalar{
		Time:  time,
		Value: value,
	}, nil
}

func unmarshalVectorResultData(data []byte) (*logcache_v1.PromQL_Vector, error) {
	var samples []sample
	err := json.Unmarshal(data, &samples)
	if err != nil {
		return nil, err
	}

	var resultSamples []*logcache_v1.PromQL_Sample
	for _, sampleValue := range samples {
		time, value, err := disassemblePoint(sampleValue.Value)
		if err != nil {
			return nil, err
		}

		resultSamples = append(resultSamples, &logcache_v1.PromQL_Sample{
			Metric: sampleValue.Metric,
			Point: &logcache_v1.PromQL_Point{
				Time:  time,
				Value: value,
			},
		})
	}
	return &logcache_v1.PromQL_Vector{
		Samples: resultSamples,
	}, nil
}

func unmarshalMatrixResultData(data []byte) (*logcache_v1.PromQL_Matrix, error) {
	var values []series
	err := json.Unmarshal(data, &values)
	if err != nil {
		return nil, err
	}

	var serieses []*logcache_v1.PromQL_Series
	for _, value := range values {
		var points []*logcache_v1.PromQL_Point
		for _, point := range value.Values {
			time, value, err := disassemblePoint(point)
			if err != nil {
				return nil, err
			}

			points = append(points, &logcache_v1.PromQL_Point{
				Time:  time,
				Value: value,
			})
		}
		serieses = append(serieses, &logcache_v1.PromQL_Series{
			Metric: value.Metric,
			Points: points,
		})
	}

	return &logcache_v1.PromQL_Matrix{
		Series: serieses,
	}, nil
}

func disassemblePoint(point []interface{}) (string, float64, error) {
	if len(point) != 2 {
		return "", 0, fmt.Errorf("invalid length of point, got %d, expected 2", len(point))
	}

	t, ok := point[0].(float64)
	if !ok {
		return "", 0, fmt.Errorf("invalid type of point timestamp, got %T, expected number", point[0])
	}

	v, ok := point[1].(string)
	if !ok {
		return "", 0, fmt.Errorf("invalid type of value, got %T, expected string", point[1])
	}

	decodedValue, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse value: %q", err)
	}

	return strconv.FormatFloat(t, 'f', 3, 64), decodedValue, nil
}

func (m *PromqlMarshaler) NewDecoder(r io.Reader) runtime.Decoder {
	fallbackDecoder := m.fallback.NewDecoder(r)
	jsonDecoder := json.NewDecoder(r)

	return runtime.DecoderFunc(func(v interface{}) error {
		var result queryResult

		switch q := v.(type) {
		case *logcache_v1.PromQL_InstantQueryResult:
			err := jsonDecoder.Decode(&result)
			if err != nil {
				return err
			}

			r, err := m.disassembleInstantQueryResult(result)
			if err != nil {
				return err
			}
			*q = *r

			return nil
		case *logcache_v1.PromQL_RangeQueryResult:
			err := jsonDecoder.Decode(&result)
			if err != nil {
				return err
			}

			r, err := m.disassembleRangeQueryResult(result)
			if err != nil {
				return err
			}
			*q = *r

			return nil
		}

		return fallbackDecoder.Decode(v)
	})
}

func (m *PromqlMarshaler) ContentType() string {
	return `application/json`
}
