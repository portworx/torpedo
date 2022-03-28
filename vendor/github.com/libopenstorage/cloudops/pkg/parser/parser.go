package parser

import (
	"io/ioutil"

	"github.com/libopenstorage/cloudops"
	"gopkg.in/yaml.v2"
)

// StorageDecisionMatrixParser parses a cloud storage decision matrix from yamls
// to StorageDecisionMatrix objects defined in cloudops
type StorageDecisionMatrixParser interface {
	// MarshalToYaml marshals the provided StorageDecisionMatrix
	// to a yaml file at the provided path
	MarshalToYaml(*cloudops.StorageDecisionMatrix, string) error
	// UnmarshalFromYaml unmarshals the yaml file at the provided path
	// into a StorageDecisionMatrix
	UnmarshalFromYaml(string) (*cloudops.StorageDecisionMatrix, error)
	// MarshalToBytes marshals the provided StorageDecisionMatrix to bytes
	MarshalToBytes(*cloudops.StorageDecisionMatrix) ([]byte, error)
	// UnmarshalFromBytes unmarshals the given yaml bytes into a StorageDecisionMatrix
	UnmarshalFromBytes([]byte) (*cloudops.StorageDecisionMatrix, error)
}

// NewStorageDecisionMatrixParser returns an implementation of StorageDecisionMatrixParser
func NewStorageDecisionMatrixParser() StorageDecisionMatrixParser {
	return &sdmParser{}
}

type sdmParser struct{}

func (s *sdmParser) MarshalToYaml(
	matrix *cloudops.StorageDecisionMatrix,
	filePath string,
) error {
	yamlBytes, err := s.MarshalToBytes(matrix)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, yamlBytes, 0777)
}

func (s *sdmParser) UnmarshalFromYaml(
	filePath string,
) (*cloudops.StorageDecisionMatrix, error) {
	yamlBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return s.UnmarshalFromBytes(yamlBytes)
}

func (s *sdmParser) MarshalToBytes(matrix *cloudops.StorageDecisionMatrix) ([]byte, error) {
	return yaml.Marshal(matrix)
}

func (s *sdmParser) UnmarshalFromBytes(yamlBytes []byte) (*cloudops.StorageDecisionMatrix, error) {
	matrix := &cloudops.StorageDecisionMatrix{}
	if err := yaml.Unmarshal(yamlBytes, matrix); err != nil {
		return nil, err
	}
	return matrix, nil
}
