package database

import (
	"log/slog"
	"os"

	"gonum.org/v1/gonum/mat"
)

func ConvertToGonumMatrix(originalData [][]float32) *mat.Dense {
	rows := len(originalData)
	cols := len(originalData[0])
	flatData := make([]float64, 0, rows*cols)
	for _, row := range originalData {
		for _, val := range row {
			flatData = append(flatData, float64(val))
		}
	}
	return mat.NewDense(rows, cols, flatData)
}

func SaveMatrix(filename string, m *mat.Dense) error {
	file, err := os.Create(filename)
	if err != nil {
		slog.Error("Failed to create file for storing matrix", "error", err)
		return err
	}
	defer file.Close()

	data, err := m.MarshalBinary()
	if err != nil {
		slog.Error("Failed to marshal binary matrix", "error", err)
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		slog.Error("Failed to write binary matrix to file", "error", err)
	}
	return err
}

func LoadMatrix(filename string) (*mat.Dense, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		slog.Error("Failed to read file for storing matrix", "error", err)
		return nil, err
	}

	m := &mat.Dense{}
	err = m.UnmarshalBinary(data)
	if err != nil {
		slog.Error("Failed to unmarshal binary matrix", "error", err)
		return nil, err
	}

	return m, nil
}
