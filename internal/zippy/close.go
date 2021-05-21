package zippy

func (z ZipReader) Close() {
	z.reader.Close()
}
