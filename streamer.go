package muta

type FileInfo struct {
	Name string
	Path string
	Ctx  interface{}
}

type Streamer func(*FileInfo, []byte) ([]byte, error)

func SrcStreamer(s ...string) Streamer {
	return func(fi *FileInfo, chunk []byte) ([]byte, error) {
		// If there is an incoming file pass the data along unmodified. This
		// func doesn't care to modify the data in any way
		if fi != nil {
			return chunk, nil
		}
		// Not generating files at the moment
		return nil, nil
	}
}

func Dest(d string) Streamer {
	return func(fi *FileInfo, chunk []byte) ([]byte, error) {
		return nil, nil
	}
}
