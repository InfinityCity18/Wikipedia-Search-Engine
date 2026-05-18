package pythonsvd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os/exec"

	"gonum.org/v1/gonum/mat"
)

func TruncatedSVD(rowsCount, colsCount, k int, sparseMap map[int]map[int]float64) (*mat.Dense, *mat.Dense, *mat.Dense, error) {
	var rows []int64
	var cols []int64
	var vals []float64

	for r, colMap := range sparseMap {
		for c, val := range colMap {
			if val != 0.0 {
				rows = append(rows, int64(r))
				cols = append(cols, int64(c))
				vals = append(vals, val)
			}
		}
	}
	nnz := int64(len(vals))

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, int64(rowsCount))
	binary.Write(buf, binary.LittleEndian, int64(colsCount))
	binary.Write(buf, binary.LittleEndian, int64(k))
	binary.Write(buf, binary.LittleEndian, nnz)

	binary.Write(buf, binary.LittleEndian, rows)
	binary.Write(buf, binary.LittleEndian, cols)
	binary.Write(buf, binary.LittleEndian, vals)

	cmd := exec.Command("python3", "pythonsvd.py")
	cmd.Stdin = buf

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		return nil, nil, nil, fmt.Errorf("python execution error: %w\nstderr: %s", err, stderrBuf.String())
	}

	var uR, uC, sLen, vR, vC int64
	binary.Read(&stdoutBuf, binary.LittleEndian, &uR)
	binary.Read(&stdoutBuf, binary.LittleEndian, &uC)
	binary.Read(&stdoutBuf, binary.LittleEndian, &sLen)
	binary.Read(&stdoutBuf, binary.LittleEndian, &vR)
	binary.Read(&stdoutBuf, binary.LittleEndian, &vC)

	uData := make([]float64, uR*uC)
	sData := make([]float64, sLen)
	vData := make([]float64, vR*vC)

	binary.Read(&stdoutBuf, binary.LittleEndian, &uData)
	binary.Read(&stdoutBuf, binary.LittleEndian, &sData)
	binary.Read(&stdoutBuf, binary.LittleEndian, &vData)

	U := mat.NewDense(int(uR), int(uC), uData)

	S := mat.NewDense(int(sLen), int(sLen), nil)
	for i, val := range sData {
		S.Set(i, i, val)
	}

	V := mat.NewDense(int(vR), int(vC), vData)

	return U, S, V, nil
}
