import sys
import struct
import numpy as np
from scipy.sparse import coo_matrix
from scipy.sparse.linalg import svds

def main():
    header_bytes = sys.stdin.buffer.read(32)
    if not header_bytes:
        return
    
    shape_rows, shape_cols, k, nnz = struct.unpack('<4q', header_bytes)

    rows_bytes = sys.stdin.buffer.read(nnz * 8)
    cols_bytes = sys.stdin.buffer.read(nnz * 8)
    vals_bytes = sys.stdin.buffer.read(nnz * 8)

    rows = np.frombuffer(rows_bytes, dtype=np.int64)
    cols = np.frombuffer(cols_bytes, dtype=np.int64)
    vals = np.frombuffer(vals_bytes, dtype=np.float64)

    sparse_mat = coo_matrix((vals, (rows, cols)), shape=(shape_rows, shape_cols))
    u, s, vt = svds(sparse_mat, k=k)

    idx = np.argsort(s)[::-1]
    u = np.ascontiguousarray(u[:, idx], dtype=np.float64)
    s = np.ascontiguousarray(s[idx], dtype=np.float64)
    vt = np.ascontiguousarray(vt[idx, :], dtype=np.float64)

    sys.stdout.buffer.write(struct.pack('<5q', u.shape[0], u.shape[1], len(s), vt.shape[0], vt.shape[1]))
    sys.stdout.buffer.write(u.tobytes())
    sys.stdout.buffer.write(s.tobytes())
    sys.stdout.buffer.write(vt.tobytes())

if __name__ == "__main__":
    main()